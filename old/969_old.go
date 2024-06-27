// +build !windows

// Copyright 2014-2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package api

import (
	"testing"

	docker "github.com/fsouza/go-dockerclient"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/assert"
)

const (
	emptyVolumeName1                  = "empty-volume-1"
	emptyVolumeContainerPath1         = "/my/empty-volume-1"
	expectedEmptyVolumeGeneratedPath1 = "/ecs-empty-volume/" + emptyVolumeName1

	emptyVolumeName2                  = "empty-volume-2"
	emptyVolumeContainerPath2         = "/my/empty-volume-2"
	expectedEmptyVolumeGeneratedPath2 = "/ecs-empty-volume/" + emptyVolumeName2

	expectedEmptyVolumeContainerImage = "amazon/ecs-emptyvolume-base"
	expectedEmptyVolumeContainerTag   = "autogenerated"
	expectedEmptyVoluemContainerCmd   = "not-applicable"

	validTaskArn   = "arn:aws:ecs:region:account-id:task/task-id"
	invalidTaskArn = "invalid:task::arn"

	expectedCgroupRoot = "/ecs/task-id"

	taskVCPULimit   = 2.0
	taskMemoryLimit = 512
)

// TestBuildCgroupRootHappyPath builds cgroup root from valid taskARN
func TestBuildCgroupRootHappyPath(t *testing.T) {
	task := Task{
		Arn: validTaskArn,
	}

	cgroupRoot, err := task.BuildCgroupRoot()

	assert.NoError(t, err)
	assert.Equal(t, expectedCgroupRoot, cgroupRoot)
}

// TestBuildCgroupRootErrorPath validates the cgroup path build error path
func TestBuildCgroupRootErrorPath(t *testing.T) {
	task := Task{
		Arn: invalidTaskArn,
	}

	cgroupRoot, err := task.BuildCgroupRoot()

	assert.Error(t, err)
	assert.Empty(t, cgroupRoot)
}

// TestBuildLinuxResourceSpecCPUMem validates the linux resource spec builder
func TestBuildLinuxResourceSpecCPUMem(t *testing.T) {
	taskMemoryLimit := int64(taskMemoryLimit)

	task := &Task{
		Arn:         validTaskArn,
		VCPULimit:   float64(taskVCPULimit),
		MemoryLimit: taskMemoryLimit,
	}

	expectedTaskCPUQuota := int64(taskVCPULimit * defaultCPUPeriod)
	expectedTaskCPUPeriod := uint64(defaultCPUPeriod)
	expectedLinuxResourceSpec := specs.LinuxResources{
		CPU: &specs.LinuxCPU{
			Quota:  &expectedTaskCPUQuota,
			Period: &expectedTaskCPUPeriod,
		},
		Memory: &specs.LinuxMemory{
			Limit: &taskMemoryLimit,
		},
	}

	linuxResourceSpec := task.BuildLinuxResourceSpec()

	assert.EqualValues(t, expectedLinuxResourceSpec, linuxResourceSpec)
}

// TestBuildLinuxResourceSpecCPU validates the linux resource spec builder
func TestBuildLinuxResourceSpecCPU(t *testing.T) {
	task := &Task{
		Arn:       validTaskArn,
		VCPULimit: float64(taskVCPULimit),
	}

	expectedTaskCPUQuota := int64(taskVCPULimit * defaultCPUPeriod)
	expectedTaskCPUPeriod := uint64(defaultCPUPeriod)
	expectedLinuxResourceSpec := specs.LinuxResources{
		CPU: &specs.LinuxCPU{
			Quota:  &expectedTaskCPUQuota,
			Period: &expectedTaskCPUPeriod,
		},
	}

	linuxResourceSpec := task.BuildLinuxResourceSpec()

	assert.EqualValues(t, expectedLinuxResourceSpec, linuxResourceSpec)
}

// TestOverrideCgroupParent validates the cgroup parent override
func TestOverrideCgroupParentHappyPath(t *testing.T) {
	task := &Task{
		Arn:         validTaskArn,
		VCPULimit:   float64(taskVCPULimit),
		MemoryLimit: int64(taskMemoryLimit),
	}

	hostConfig := &docker.HostConfig{}
	task.overrideCgroupParent(hostConfig)

	assert.NotEmpty(t, hostConfig)
	assert.Equal(t, expectedCgroupRoot, hostConfig.CgroupParent)
}

// TestOverrideCgroupParentErrorPath validates the error path for
// cgroup parent update
func TestOverrideCgroupParentErrorPath(t *testing.T) {
	task := &Task{
		Arn:         invalidTaskArn,
		VCPULimit:   float64(taskVCPULimit),
		MemoryLimit: int64(taskMemoryLimit),
	}

	hostConfig := &docker.HostConfig{}
	task.overrideCgroupParent(hostConfig)

	assert.Empty(t, hostConfig.CgroupParent)
}

// TestPlatformHostConfigOverride validates the platform host config overrides
func TestPlatformHostConfigOverride(t *testing.T) {
	task := &Task{
		Arn:         validTaskArn,
		VCPULimit:   float64(taskVCPULimit),
		MemoryLimit: int64(taskMemoryLimit),
	}

	hostConfig := &docker.HostConfig{}
	task.platformHostConfigOverride(hostConfig)

	assert.NotEmpty(t, hostConfig)
	assert.Equal(t, expectedCgroupRoot, hostConfig.CgroupParent)
}

// TestCgroupEnabledAll validates the happy path for cgroup enabled
func TestCgroupEnabledAll(t *testing.T) {
	task := &Task{
		VCPULimit:   float64(taskVCPULimit),
		MemoryLimit: int64(taskMemoryLimit),
	}

	assert.True(t, task.CgroupEnabled())
}

// TestCgroupEnabledPartial validates the happy path for partial resource limits
func TestCgroupEnabledPartial(t *testing.T) {
	task := &Task{
		VCPULimit: float64(taskVCPULimit),
	}

	assert.True(t, task.CgroupEnabled())
}

// TestCgroupDisabled validates the cgroup disabled path
func TestCgroupDisabled(t *testing.T) {
	task := &Task{
		Arn: validTaskArn,
	}

	assert.False(t, task.CgroupEnabled())
}

// TestCgroupDisabledMem validates the cgroup disabled path
func TestCgroupDisabledMem(t *testing.T) {
	task := &Task{
		Arn:         validTaskArn,
		MemoryLimit: int64(taskMemoryLimit),
	}

	assert.False(t, task.CgroupEnabled())
}
