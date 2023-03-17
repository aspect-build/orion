package outputs

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aspect-build/silo/cli/core/pkg/bazel"
	"golang.org/x/exp/slices"
	"golang.org/x/mod/sumdb/dirhash"
	"golang.org/x/sync/errgroup"
)

// addExecutableHash appends the exePath to hashFiles entry of the label
func addExecutableHash(hashFiles map[string][]string, label string, exePath string) {
	_, err := os.Stat(exePath)
	if os.IsNotExist(err) {
		fmt.Printf("%s output %s is not on disk, did you build it? Skipping...\n", label, exePath)
		return
	}

	hashFiles[label] = append(hashFiles[label], exePath)
}

// addRunfilesHash iterates through the runfiles entries of the manifest, appending all files
// contained (or files inside directories) to the hashFiles entry of the label
func addRunfilesHash(hashFiles map[string][]string, label string, manifestPath string) error {
	_, err := os.Stat(manifestPath)
	if os.IsNotExist(err) {
		fmt.Printf("%s manifest %s is not on disk, did you build it? Skipping...\n", label, manifestPath)
		return nil
	}
	runfiles, err := os.Open(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to open runfiles manifest %s: %w\n", manifestPath, err)
	}
	defer runfiles.Close()

	fileScanner := bufio.NewScanner(runfiles)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		// Manifest entries are in the form
		// execroot/path /some/absolute/path
		entry := strings.Split(fileScanner.Text(), " ")
		// key := entry[0]
		abspath := entry[1]
		fileinfo, err := os.Stat(abspath)

		if err != nil {
			return fmt.Errorf("failed to stat runfiles manifest entry %s: %w\n", abspath, err)
		}

		if fileinfo.IsDir() {
			// TODO(alexeagle): I think the abspath means we'll get more hashed than we mean to
			// we should pass some other value to the second arg "prefix"
			direntries, err := dirhash.DirFiles(abspath, abspath)
			if err != nil {
				return fmt.Errorf("failed to recursively list directory %s: %w\n", abspath, err)
			}
			hashFiles[label] = append(hashFiles[label], direntries...)
		} else {
			hashFiles[label] = append(hashFiles[label], abspath)
		}
	}
	return nil
}

// Resolve symlinks and directories to find extra outs
// that can later be hashed.
//
// map from Label to the files/directories which should be hashed
func resolveOuts(outs []bazel.Output) (map[string][]string, error) {
	hashFiles := make(map[string][]string)

	for _, a := range outs {
		if a.Mnemonic == "ExecutableSymlink" {
			// NB: This stats the file.
			// TODO: We could probably optimize that away.
			addExecutableHash(hashFiles, a.Label, a.Path)
		} else if a.Mnemonic == "SourceSymlinkManifest" {
			// NB: This must access and stat the files to see if they are directories
			// TODO: An optimization could compare them to the output state, maybe.
			if err := addRunfilesHash(hashFiles, a.Label, a.Path); err != nil {
				return nil, err
			}
		}
		// Other Mnemonics are discarded.
	}

	return hashFiles, nil
}

// Split the `hashFiles` map into two,
// one that contains source files and symlinks that point to source files
// and another that contains generated files and symlinks that point to them.
func splitHashFiles(hashFiles map[string][]string) (sourceFiles, outputFiles map[string][]string, err error) {
	sourceFiles = make(map[string][]string)
	outputFiles = make(map[string][]string)

	for label, files := range hashFiles {
		for _, file := range files {
			resolved, err := filepath.EvalSymlinks(file)
			if err != nil {
				return nil, nil, fmt.Errorf("could not resolve path: %w\n", err)
			}

			// TODO: make this more robust,
			// if we have the output base we can perform a full prefix match.
			// If we elect to find the output base in the main program.
			// We can plumb the real bazel-out in here.
			// TODO: And consider cross-platform filepath comparisons.
			if strings.Contains(resolved, "/bazel-out/") {
				outputFiles[label] = append(outputFiles[label], file)
			} else {
				sourceFiles[label] = append(sourceFiles[label], file)
			}
		}
	}

	return
}

// Equivalent to the per-file hashing from `dirhash.Hash1`
// The hash and filepath can later be combined to calculate the "h1:" directory hash.
func hashFile(file string) (string, error) {
	hasher := sha256.New()
	f, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(hasher, f); err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// Print the hashes for executable targets
// This traverses the runfile tree for all (runtime) dependencies,
// if `bbclientdStatePath` is given it uses `bb_clientd`'s
// persistent output state to avoid file operations.
func printExecutableHashes(outs []bazel.Output, bbclientdStatePath string, maximumStateFileSizeBytes int64) (map[string]string, error) {

	result := make(map[string]string)
	resultHashes := make(map[string][][]string)

	hashFiles, err := resolveOuts(outs)
	if err != nil {
		return nil, err
	}

	// All source files and generated files if there is no bbclientd state file.
	var readFileContentHashFiles map[string][]string

	// The next two blocks/loops iterate over the files in the two maps
	// and compute their individual hashes.
	// They will write into the same result map,
	// and the resulting overall hash can be calculated afterwards.

	// Output files can be optimized.
	//
	// TODO: We could check whether the inner arrays of the `checkOutputStateHashFiles`
	// are empty and avoid reading the state file entirely.
	if bbclientdStatePath != "" {
		// Generated files has their hashes in the bbclientd state file.
		var checkOutputStateHashFiles map[string][]string
		readFileContentHashFiles, checkOutputStateHashFiles, err = splitHashFiles(hashFiles)
		if err != nil {
			return nil, err
		}

		reader, err := NewStateFileReader(bbclientdStatePath, maximumStateFileSizeBytes)
		if err != nil {
			return nil, err
		}
		defer reader.Close()

		for label, paths := range checkOutputStateHashFiles {
			// NB: This reads once per label.
			// It would be better to flatten the dict and call `Lookup` once
			// and then rebuild the dictionary form.
			hashes, missing, err := reader.Lookup(paths)
			if err != nil {
				return nil, err
			}
			if len(missing) > 0 {
				return nil, fmt.Errorf("unable to find hashes for requested artifacts: %v, did you build with Remote Output Service and bb-clientd?", missing)
			}

			hashStrings := make([][]string, 0, len(hashes))
			for path, hash := range hashes {
				hashStrings = append(hashStrings, []string{path, hash})
			}

			resultHashes[label] = append(resultHashes[label], hashStrings...)
		}
	} else {
		readFileContentHashFiles = hashFiles
	}

	// All source files must be read from disk, and build artifacts without Remote Output Service.
	for label, files := range readFileContentHashFiles {
		hashStrings := make([][]string, len(files))

		g := new(errgroup.Group)
		for i, file := range files {
			i, file := i, file
			g.Go(func() error {
				fileHash, err := hashFile(file)
				if err != nil {
					return err
				}

				hashStrings[i] = []string{file, fileHash}
				return nil
			})
		}
		err := g.Wait()
		if err != nil {
			return nil, err
		}
		resultHashes[label] = append(resultHashes[label], hashStrings...)
	}

	// Calculate the directory hash for the results.
	for label, hashStrings := range resultHashes {
		overallHash, err := hash1(hashStrings)
		if err != nil {
			return nil, err
		}
		result[label] = overallHash
	}

	return result, nil
}

// Calculate the directory hash for already hashed files.
// This follows the implementation for `dirhash.Hash1`
// but omits all the file operations.
//
// hash1 is the "h1:" directory hash function, using SHA-256.
func hash1(fileHashes [][]string) (string, error) {
	h := sha256.New()

	slices.SortFunc(fileHashes, func(a, b []string) bool { return a[0] < b[0] })

	for _, fileHash := range fileHashes {
		file := fileHash[0]
		hash := fileHash[1]
		if strings.Contains(file, "\n") {
			// NB: Bazel will not create such targets, but it is allowed from a unix-perspective.
			return "", errors.New("dirhash: filenames with newlines are not supported")
		}
		fmt.Fprintf(h, "%s  %s\n", hash, file)
	}

	return "h1:" + base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}
