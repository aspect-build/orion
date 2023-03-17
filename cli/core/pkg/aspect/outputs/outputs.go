/*
 * Copyright 2022 Aspect Build Systems, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package outputs

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aspect-build/silo/cli/core/pkg/bazel"
	"github.com/aspect-build/silo/cli/core/pkg/ioutils"
)

const (
	BbClientdStateFlag       = "bb-clientd-output-state-prefix"
	MaximumStateFileSizeFlag = "bb-clientd-output-state-maximum-size-bytes"
)

type Outputs struct {
	ioutils.Streams
	bzl bazel.Bazel
}

func New(streams ioutils.Streams, bzl bazel.Bazel) *Outputs {
	return &Outputs{
		Streams: streams,
		bzl:     bzl,
	}
}

func (runner *Outputs) Run(_ context.Context, cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("a label is required as the first argument to aspect outputs")
	}
	query := args[0]
	var mnemonicFilter string
	if len(args) > 1 {
		mnemonicFilter = args[1]
	}

	var bbclientdStateFile string
	bbclientdStatePrefix, err := cmd.Flags().GetString(BbClientdStateFlag)
	if err != nil {
		return fmt.Errorf("cannot parse the bb_clientd state file path prefix flag: %w", err)
	}
	maximumStateFileSizeBytes, err := cmd.Flags().GetInt64(MaximumStateFileSizeFlag)
	if err != nil {
		return fmt.Errorf("cannot parse the bb_clientd state file maximum size flag: %w", err)
	}

	if bbclientdStatePrefix != "" {
		// This can also be done with `bazel-info` if we prefer a sub process.
		// and the pipe handling for that.
		bo, err := runner.bzl.AbsPathRelativeToWorkspace("bazel-out")
		if err != nil {
			return err
		}
		bazelOut, err := os.Readlink(bo)
		if err != nil {
			return err
		}
		parts := strings.Split(bazelOut, "/")
		// Index from the end to work with non-standard locations.
		outputBase := parts[len(parts)-4]
		bbclientdStateFile = path.Join(bbclientdStatePrefix, outputBase)
	}

	// TODO: To maintain performance this should be aware of the requisite remote execution flags
	// used by the previous build. As this would drop the analysis cache otherwise.
	agc, err := runner.bzl.AQuery(query)
	if err != nil {
		return err
	}
	outs := bazel.ParseOutputs(agc)

	// Special case pseudo-mnemonic indicating we should compute an overall hash
	// for any executables in the aquery result
	if mnemonicFilter == "ExecutableHash" {
		hashes, err := printExecutableHashes(outs, bbclientdStateFile, maximumStateFileSizeBytes)
		if err != nil {
			return err
		}
		for label, hash := range hashes {
			fmt.Printf("%s %s\n", label, hash)
		}
		return nil
	}

	for _, a := range outs {
		if len(mnemonicFilter) > 0 {
			if a.Mnemonic == mnemonicFilter {
				fmt.Printf("%s\n", a.Path)
			}
		} else {
			fmt.Printf("%s %s\n", a.Mnemonic, a.Path)
		}
	}
	return nil
}
