// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package js

import (
	"github.com/neglectedta/kres/internal/dag"
	"github.com/neglectedta/kres/internal/output/dockerfile"
	"github.com/neglectedta/kres/internal/output/dockerfile/step"
	"github.com/neglectedta/kres/internal/output/drone"
	"github.com/neglectedta/kres/internal/output/ghworkflow"
	"github.com/neglectedta/kres/internal/output/makefile"
	"github.com/neglectedta/kres/internal/project/meta"
)

// UnitTests runs unit-tests for Go packages.
type UnitTests struct {
	meta *meta.Options

	dag.BaseNode
}

// NewUnitTests initializes UnitTests.
func NewUnitTests(meta *meta.Options, name string) *UnitTests {
	return &UnitTests{
		BaseNode: dag.NewBaseNode(name),
		meta:     meta,
	}
}

// CompileDockerfile implements dockerfile.Compiler.
func (tests *UnitTests) CompileDockerfile(output *dockerfile.Output) error {
	output.Stage(tests.Name()).
		Description("runs js unit-tests").
		From("js").
		Step(step.Script(`bun add -d @happy-dom/global-registrator`).
			MountCache(tests.meta.JSCachePath, tests.meta.GitHubRepository, step.CacheLocked)).
		Step(step.Script(`bun run test`).
			MountCache(tests.meta.JSCachePath, tests.meta.GitHubRepository).
			Env("CI", "true"))

	return nil
}

// CompileMakefile implements makefile.Compiler.
func (tests *UnitTests) CompileMakefile(output *makefile.Output) error {
	output.VariableGroup(makefile.VariableGroupCommon).
		Variable(makefile.OverridableVariable("TESTPKGS", "./..."))

	output.Target(tests.Name()).
		Description("Performs unit tests").
		Script("@$(MAKE) target-$@").
		Phony()

	return nil
}

// CompileDrone implements drone.Compiler.
func (tests *UnitTests) CompileDrone(output *drone.Output) error {
	output.Step(drone.MakeStep(tests.Name()).
		DependsOn(dag.GatherMatchingInputNames(tests, dag.Implements[drone.Compiler]())...),
	)

	return nil
}

// CompileGitHubWorkflow implements ghworkflow.Compiler.
func (tests *UnitTests) CompileGitHubWorkflow(output *ghworkflow.Output) error {
	output.AddStep(
		"default",
		ghworkflow.Step(tests.Name()).SetMakeStep(tests.Name()),
	)

	return nil
}
