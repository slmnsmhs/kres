// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package golang

import (
	"fmt"
	"path/filepath"

	"github.com/neglectedta/kres/internal/config"
	"github.com/neglectedta/kres/internal/dag"
	"github.com/neglectedta/kres/internal/output/dockerfile"
	"github.com/neglectedta/kres/internal/output/dockerfile/step"
	"github.com/neglectedta/kres/internal/output/makefile"
	"github.com/neglectedta/kres/internal/project/meta"
)

// DeepCopy provides goimports deepcopyer.
type DeepCopy struct {
	dag.BaseNode

	meta *meta.Options

	Version string `yaml:"version"`
}

// NewDeepCopy builds DeepCopy node.
func NewDeepCopy(meta *meta.Options) *DeepCopy {
	meta.BuildArgs = append(meta.BuildArgs, "DEEPCOPY_VERSION")

	return &DeepCopy{
		BaseNode: dag.NewBaseNode("deepcopy"),

		meta: meta,

		Version: config.DeepCopyVersion,
	}
}

// CompileMakefile implements makefile.Compiler.
func (deepcopy *DeepCopy) CompileMakefile(output *makefile.Output) error {
	output.VariableGroup(makefile.VariableGroupCommon).
		Variable(makefile.OverridableVariable("DEEPCOPY_VERSION", deepcopy.Version))

	return nil
}

// ToolchainBuild implements common.ToolchainBuilder hook.
func (deepcopy *DeepCopy) ToolchainBuild(stage *dockerfile.Stage) error {
	stage.
		Step(step.Arg("DEEPCOPY_VERSION")).
		Step(step.Script(fmt.Sprintf(
			`go install github.com/siderolabs/deep-copy@${DEEPCOPY_VERSION} \
	&& mv /go/bin/deep-copy %s/deep-copy`, deepcopy.meta.BinPath)).
			MountCache(filepath.Join(deepcopy.meta.CachePath, "go-build"), deepcopy.meta.GitHubRepository).
			MountCache(filepath.Join(deepcopy.meta.GoPath, "pkg"), deepcopy.meta.GitHubRepository),
		)

	return nil
}
