// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package common

import (
	"fmt"
	"strings"

	"github.com/drone/drone-yaml/yaml"

	"github.com/neglectedta/kres/internal/config"
	"github.com/neglectedta/kres/internal/dag"
	"github.com/neglectedta/kres/internal/output/drone"
	"github.com/neglectedta/kres/internal/output/makefile"
	"github.com/neglectedta/kres/internal/project/meta"
)

// Docker provides build infrastructure via docker buildx.
type Docker struct { //nolint:govet
	dag.BaseNode

	meta *meta.Options

	DockerImage    string   `yaml:"dockerImage"`
	AllowInsecure  bool     `yaml:"allowInsecure"`
	ExtraBuildArgs []string `yaml:"extraBuildArgs"`

	DockerResourceRequests *yaml.ResourceObject `yaml:"dockerResourceRequests"`
}

// NewDocker initializes Docker.
func NewDocker(meta *meta.Options) *Docker {
	meta.BuildArgs = append(meta.BuildArgs, "USERNAME", "REGISTRY")

	return &Docker{
		BaseNode: dag.NewBaseNode("setup-ci"),

		meta: meta,

		DockerImage: "docker:" + config.DindContainerImageVersion,
	}
}

// CompileDrone implements drone.Compiler.
func (docker *Docker) CompileDrone(output *drone.Output) error {
	output.
		VolumeHostPath("outer-docker-socket", "/var/ci-docker", "/var/outer-run").
		VolumeTemporary("docker-socket", "/var/run").
		VolumeTemporary("buildx", "/root/.docker/buildx").
		VolumeTemporary("ssh", "/root/.ssh").
		VolumeHostPathStandalone("dev", "/dev")

	docker.BuildBaseDroneSteps(output)

	return nil
}

// BuildBaseDroneSteps builds the base steps which start the pipeline.
func (docker *Docker) BuildBaseDroneSteps(output drone.StepService) {
	resources := (*yaml.Resources)(nil)
	if docker.DockerResourceRequests != nil {
		resources = &yaml.Resources{
			Requests: docker.DockerResourceRequests,
		}
	}

	output.Service(&yaml.Container{
		Name:       "docker",
		Image:      docker.DockerImage,
		Entrypoint: []string{"dockerd"},
		Privileged: true,
		Commands: []string{
			"--dns=8.8.8.8",
			"--dns=8.8.4.4",
			"--mtu=1500",
			"--log-level=error",
		},
		Resources: resources,
		Volumes: []*yaml.VolumeMount{
			{
				Name:      "dev",
				MountPath: "/dev",
			},
		},
	})

	builderName := "local"
	extraArgs := []string{""}

	if docker.AllowInsecure {
		builderName += "-insecure"

		extraArgs = append(extraArgs, "--buildkitd-flags", "'--allow-insecure-entitlement security.insecure'")
	}

	output.Step(
		drone.CustomStep(docker.Name(),
			"sleep 5",
			"git fetch --tags",
			"install-ci-key",
			fmt.Sprintf("docker buildx create --driver docker-container --platform linux/amd64 --name %s%s --use unix:///var/outer-run/docker.sock", builderName, strings.Join(extraArgs, " ")),
			"docker buildx inspect --bootstrap",
		).EnvironmentFromSecret("SSH_KEY", "ssh_key"),
	)
}

// CompileMakefile implements makefile.Compiler.
func (docker *Docker) CompileMakefile(output *makefile.Output) error {
	// Only compile if the Dockerfile is used.
	if docker.meta.ContainerImageFrontend != config.ContainerImageFrontendDockerfile {
		return nil
	}

	buildArgs := makefile.RecursiveVariable("COMMON_ARGS", "--file=Dockerfile").
		Push("--provenance=false").
		Push("--progress=$(PROGRESS)").
		Push("--platform=$(PLATFORM)").
		Push("--build-arg=BUILDKIT_MULTI_PLATFORM=$(BUILDKIT_MULTI_PLATFORM)").
		Push("--push=$(PUSH)")

	for _, arg := range docker.meta.BuildArgs {
		buildArgs.Push(fmt.Sprintf("--build-arg=%s=\"$(%s)\"", arg, arg))
	}

	for _, arg := range docker.ExtraBuildArgs {
		buildArgs.Push(fmt.Sprintf("--build-arg=%s=\"$(%s)\"", arg, arg))
	}

	output.VariableGroup(makefile.VariableGroupCommon).
		Variable(makefile.OverridableVariable("REGISTRY", "ghcr.io")).
		Variable(makefile.OverridableVariable("USERNAME", docker.meta.GitHubOrganization)).
		Variable(makefile.OverridableVariable("REGISTRY_AND_USERNAME", "$(REGISTRY)/$(USERNAME)"))

	output.VariableGroup(makefile.VariableGroupDocker).
		Variable(makefile.SimpleVariable("BUILD", "docker buildx build")).
		Variable(makefile.OverridableVariable("PLATFORM", "linux/amd64")).
		Variable(makefile.OverridableVariable("PROGRESS", "auto")).
		Variable(makefile.OverridableVariable("PUSH", "false")).
		Variable(makefile.OverridableVariable("CI_ARGS", "")).
		Variable(makefile.OverridableVariable("BUILDKIT_MULTI_PLATFORM", "")).
		Variable(buildArgs)

	output.Target("target-%").
		Description("Builds the specified target defined in the Dockerfile. The build result will only remain in the build cache.").
		Script(`@$(BUILD) --target=$* $(COMMON_ARGS) $(TARGET_ARGS) $(CI_ARGS) .`)

	output.Target("registry-%").
		Description("Builds the specified target defined in the Dockerfile and the output is an image. The image is pushed to the registry if PUSH=true.").
		Script(`@$(MAKE) target-$* TARGET_ARGS="--tag=$(REGISTRY)/$(USERNAME)/$(IMAGE_NAME):$(IMAGE_TAG)" BUILDKIT_MULTI_PLATFORM=1`)

	output.Target("local-%").
		Description("Builds the specified target defined in the Dockerfile using the local output type. The build result will be output to the specified local destination.").
		Script(`@$(MAKE) target-$* TARGET_ARGS="--output=type=local,dest=$(DEST) $(TARGET_ARGS)"` + FixLocalDestLocationsScript)

	return nil
}
