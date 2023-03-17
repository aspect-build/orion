package outputs

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/buildbarn/bb-clientd/pkg/outputpathpersistency"
	outputpathpersistency_pb "github.com/buildbarn/bb-remote-execution/pkg/proto/outputpathpersistency"
	"golang.org/x/exp/slices"
)

// Splits the directory and file (basename) parts.
func splitDirectory(path string) (string, string) {
	final := strings.LastIndex(path, "/")
	b := path[final+1:]
	d := path[:final]

	return d, b
}

type entry struct {
	components []string
	directory  *outputpathpersistency_pb.Directory
	reader     outputpathpersistency.Reader
}

type stack []entry

func (s *stack) push(e entry) entry {
	*s = append(*s, e)
	return e
}

func (s *stack) peekIndex() (entry, int) {
	if len(*s) == 0 {
		log.Fatal("popping an empty stack.")
	}
	i := len(*s) - 1
	e := (*s)[i]
	return e, i
}

func (s *stack) peek() entry {
	e, _ := s.peekIndex()
	return e
}

func (s *stack) pop() entry {
	e, i := s.peekIndex()
	*s = (*s)[:i]
	return e
}

// Lookup the hashes for files in bbclientd's state file.
//
// Artifact paths are compared to the state file with paths relative to `bazel-out`.
// It is preferred to call the function with such paths,
// but longer, absolute, paths can be used too.
// Then this will try to strip all leading components,
// before matching against the output state file.
//
// Returns:
//
//   FileHashes: A map of found artifacts and their hashes.
//               The paths are returned unmodified.
//   []string:   A slice with any artifacts that were not found.
//   error:      error
//
// # Missing artifacts and build consistency with IDE tools:
//
// The lookup may not find all requested files,
// the output path persistency contains the last build,
// and the machinations of `bazel-out` are outside the scope for this document.
// This requires a Bazel built with the `Remote Output Service` patches.
// Then `bazel-out` is seemingly kept in sync.
// But if you run other builds in the workspace it can differ.
// Here automatic tools for IDEs and the like are a cause for concern.
// See if `tools/bazel` can be used to specify the required Bazel client.
//
// It is up to the caller to manage the build information.
// But this function returns a list of missing artifacts
// to help diagnose such inconsistencies.
//
// # Implementation notes:
//
// A stack is maintained as we progress through the directory tree,
// so the lookup can backtrack to previous directories to find other subdirectories.
// This must retain the `Directory` message for stack entries,
// as well as the wrapped `Reader`,
// so we can traverse them again.
// The reader only contains bookkeeping and a pointer to the real file reader
// so the memory overhead is negligible.
//
// # On absolute or relative paths:
//
// To fit in with the other code this is written to accept absolute paths
// and will just discard anything leading up to `bazel-out` in the matching phase.
// But the function would do better to require relative paths,
// to avoid this documentation.
// And the hashing would then give the same result for all users.
// Whereas now the hashing depends on where the Bazel cache is located on the user's machine.
// And the default locations contains the username.
//
// # On symbolic links:
//
// This resolves all artifact paths to operate on the real targets if they are symlinks
// But just like with absolute paths will report the path the user gave,
// this is done through a resolution map to backtrack the symlink resolution.
func (s *stateFileReader) Lookup(artifacts []string) (map[string]string, []string, error) {
	resolution := make(map[string]string, len(artifacts))
	resolvedArtifacts := make([]string, 0, len(artifacts))
	result := make(map[string]string)

	for _, artifact := range artifacts {
		resolved, err := filepath.EvalSymlinks(artifact)
		if err != nil {
			return nil, nil, fmt.Errorf("could not resolve path: %w\n", err)
		}
		resolution[resolved] = artifact
		resolvedArtifacts = append(resolvedArtifacts, resolved)
	}

	slices.Sort(resolvedArtifacts)
	res, missing, err := s.lookupArtifacts(resolvedArtifacts)

	for resolved, filenode := range res {
		original_artifact := resolution[resolved]
		result[original_artifact] = filenode
	}

	return res, missing, err
}

// The real Lookup function
//
// Takes a list of artifacts to look up in the output state file.
// Symbolic links are not handled at all, that is up to the caller.
func (s *stateFileReader) lookupArtifacts(artifacts []string) (map[string]string, []string, error) {
	if !slices.IsSorted(artifacts) {
		log.Fatal("the `artifacts` must be sorted for the depth-first algorithm.")
	}
	res := make(map[string]string, len(artifacts))

	missing := []string{}
	stack := stack{}
	stack.push(entry{[]string{""}, s.rootDirectory, s.reader})

ARTIFACT:
	for _, artifact := range artifacts {
		directory, file := splitDirectory(artifact)

		components := strings.Split(directory, "/")
		// Allow and discard leading components, and proceed with a relative path.
		cut := slices.Index(components, "bazel-out")
		if cut >= 0 {
			components = components[cut+1:]
		}

		// We want an explicit root component in the stack,
		// to prefix-matching easier.
		components = append([]string{""}, components...)

		current := stack.peek()

		for _, component := range components {
			if slices.Contains([]string{".", ".."}, component) {
				return nil, nil, fmt.Errorf("Lookup does not support '.' or '..' path components: '%s'", artifact)
			}
		}

		for depth, component := range components {
			// Look for equal components.
			if depth < len(current.components) && current.components[depth] == component {
				continue
			}

			// Look for upward components.
			// With sorted inputs it is safe to throw away the `Directory`s and `Reader`s that we pop.
			// We will not need them anymore.
			resets := len(current.components) - depth
			for i := 0; i < resets; i++ {
				current = stack.pop()
			}
		}
		remaining := components[len(current.components):]

		TRAVERSE:
		for _, component := range remaining {
			// Look for downward components.
			for _, c := range current.directory.Directories {
				if c.Name == component {
					region := c.FileRegion
					reader, dir, err := current.reader.ReadDirectory(region)
					if err != nil {
						// Unexpected state file content.
						return nil, nil, err
					}
					child := append(current.components, component)
					current = stack.push(entry{child, dir, reader})
					continue TRAVERSE
				}
			}
			return nil, nil, fmt.Errorf("Could not find path component '%s' for '%s'.", component, artifact)
		}

		for _, f := range current.directory.Files {
			if f.Name == file {
				res[artifact] = f.Digest.Hash
				continue ARTIFACT
			}
		}
		// SymlinkNodes from the REAPI are not handled here.
		// This should just be called with real artifacts.

		missing = append(missing, artifact)
	}

	return res, missing, nil
}

type stateFileReader struct {
	reader        outputpathpersistency.Reader
	rootDirectory *outputpathpersistency_pb.Directory
	closer        io.Closer
}

func NewStateFileReader(statefile string, maximumStateFileSizeBytes int64) (*stateFileReader, error) {
	rawReader, err := os.Open(statefile)
	if err != nil {
		return nil, err
	}
	reader, root, err := outputpathpersistency.NewFileReader(rawReader, maximumStateFileSizeBytes)
	if err != nil {
		return nil, err
	}

	return &stateFileReader{
		reader:        reader,
		rootDirectory: root.Contents,
		closer:        rawReader,
	}, err
}

func (s *stateFileReader) Close() error {
	return s.closer.Close()
}
