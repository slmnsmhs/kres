// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package common

import (
	"github.com/siderolabs/gen/xslices"

	"github.com/neglectedta/kres/internal/dag"
	"github.com/neglectedta/kres/internal/output/renovate"
	"github.com/neglectedta/kres/internal/project/meta"
)

// Renovate is a node that represents the renovate configuration.
type Renovate struct {
	dag.BaseNode

	meta *meta.Options

	CustomManagers []CustomManager `yaml:"customManagers,omitempty"`
	PackageRules   []PackageRule   `yaml:"packageRules,omitempty"`
	Enabled        bool            `yaml:"enabled"`
}

// CustomManager represents a custom manager.
type CustomManager struct {
	CustomType         string   `yaml:"customType"`
	DataSourceTemplate string   `yaml:"datasourceTemplate,omitempty"`
	DepNameTemplate    string   `yaml:"depNameTemplate,omitempty"`
	VersioningTemplate string   `yaml:"versioningTemplate"`
	FileMatch          []string `yaml:"fileMatch"`
	MatchStrings       []string `yaml:"matchStrings"`
}

// PackageRule represents a package rule.
type PackageRule struct {
	Enabled           *bool    `yaml:"enabled,omitempty"`
	AllowedVersions   string   `yaml:"allowedVersions,omitempty"`
	Versioning        string   `yaml:"versioning,omitempty"`
	MatchDataSources  []string `yaml:"matchDataSources,omitempty"`
	MatchFileNames    []string `yaml:"matchFileNames,omitempty"`
	MatchPaths        []string `yaml:"matchPaths,omitempty"`
	MatchPackageNames []string `yaml:"matchPackageNames,omitempty"`
}

// NewRenovate creates a new Renovate node.
func NewRenovate(meta *meta.Options) *Renovate {
	return &Renovate{
		BaseNode: dag.NewBaseNode("renovate"),

		meta: meta,

		Enabled: true,
	}
}

// CompileRenovate implements renovate.Compiler.
func (r *Renovate) CompileRenovate(o *renovate.Output) error {
	if !r.Enabled {
		return nil
	}

	o.Enable()
	o.CustomManagers(xslices.Map(r.CustomManagers, func(cm CustomManager) renovate.CustomManager {
		return renovate.CustomManager{
			CustomType:         cm.CustomType,
			DataSourceTemplate: cm.DataSourceTemplate,
			DepNameTemplate:    cm.DepNameTemplate,
			FileMatch:          cm.FileMatch,
			MatchStrings:       cm.MatchStrings,
			VersioningTemplate: cm.VersioningTemplate,
		}
	}))
	o.PackageRules(xslices.Map(r.PackageRules, func(pr PackageRule) renovate.PackageRule {
		return renovate.PackageRule{
			Enabled:           pr.Enabled,
			AllowedVersions:   pr.AllowedVersions,
			MatchDataSources:  pr.MatchDataSources,
			MatchFileNames:    pr.MatchFileNames,
			MatchPaths:        pr.MatchPaths,
			MatchPackageNames: pr.MatchPackageNames,
			Versioning:        pr.Versioning,
		}
	}))

	return nil
}
