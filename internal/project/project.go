// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package project provides top-level view of the whole project consisting of building blocks.
package project

import (
	"github.com/neglectedta/kres/internal/config"
	"github.com/neglectedta/kres/internal/dag"
	"github.com/neglectedta/kres/internal/output"
)

// Contents is a DAG of the project.
type Contents struct {
	dag.BaseGraph
}

// Compile the project to specified outputs.
func (project *Contents) Compile(outputs []output.Writer) error {
	for _, output := range outputs {
		visited := map[dag.Node]struct{}{}

		if err := project.CompileTo(output, visited); err != nil {
			return err
		}
	}

	return nil
}

// CompileTo project to specified output.
func (project *Contents) CompileTo(out output.Writer, visited map[dag.Node]struct{}) error {
	return dag.Walk(project, func(node dag.Node) error {
		return out.Compile(node)
	}, visited, -1)
}

// LoadConfig walks the tree and loads the config into every node.
func (project *Contents) LoadConfig(config *config.Provider) error {
	visited := map[dag.Node]struct{}{}

	return dag.Walk(project, func(node dag.Node) error {
		if err := config.Load(node); err != nil {
			return err
		}

		// allow for nodes to implement an AfterLoad function to edit the loaded config
		// before compilation starts.
		if loadHook, ok := node.(interface{ AfterLoad() error }); ok {
			return loadHook.AfterLoad()
		}

		return nil
	}, visited, -1)
}
