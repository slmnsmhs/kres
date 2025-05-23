// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package common

import (
	"github.com/neglectedta/kres/internal/config"
	"github.com/neglectedta/kres/internal/dag"
	"github.com/neglectedta/kres/internal/output/makefile"
	"github.com/neglectedta/kres/internal/project/meta"
)

// All builds Makefile `all` target.
type All struct { //nolint:govet
	dag.BaseNode

	meta *meta.Options
}

// NewAll initializes All.
func NewAll(meta *meta.Options) *All {
	return &All{
		BaseNode: dag.NewBaseNode("all"),

		meta: meta,
	}
}

// CompileMakefile implements makefile.Compiler.
func (all *All) CompileMakefile(output *makefile.Output) error {
	if all.meta.ContainerImageFrontend != config.ContainerImageFrontendDockerfile {
		return nil
	}

	output.Target("all").
		Depends(dag.GatherMatchingInputNames(all, dag.Not(dag.Implements[makefile.SkipAsMakefileDependency]()))...)

	return nil
}
