// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package task

import (
	"fmt"

	"github.com/aws/copilot-cli/internal/pkg/aws/ec2"
	"github.com/aws/copilot-cli/internal/pkg/aws/ecs"
	"github.com/aws/copilot-cli/internal/pkg/deploy"
)

const (
	fmtErrPublicSubnetsFromEnv  = "get public subnet IDs from environment %s: %w "
	fmtErrSecurityGroupsFromEnv = "get security groups from environment %s: %w"

	envSecurityGroupCFNLogicalIDTagKey   = "aws:cloudformation:logical-id"
	envSecurityGroupCFNLogicalIDTagValue = "EnvironmentSecurityGroup"
)

// Names for tag filters
var (
	fmtErrNoSubnetFoundInEnv = fmt.Sprintf("no subnets found from environment %s (need %s and %s tag?)", "%s", deploy.AppTagKey, deploy.EnvTagKey)

	tagFilterNameForApp = fmt.Sprintf(ec2.TagFilterName, deploy.AppTagKey)
	tagFilterNameForEnv = fmt.Sprintf(ec2.TagFilterName, deploy.EnvTagKey)
)

// EnvRunner can run an Amazon ECS task in the VPC and the cluster of an environment.
type EnvRunner struct {
	// Count of the tasks to be launched.
	Count int
	// Group Name of the tasks that use the same task definition.
	GroupName string

	// App and Env in which the tasks will be launched.
	App string
	Env string

	// Interfaces to interact with dependencies. Must not be nil.
	VPCGetter     VPCGetter
	ClusterGetter ClusterGetter
	Starter       Runner
}

// Run runs tasks in the environment of the application, and returns the tasks.
func (r *EnvRunner) Run() ([]*Task, error) {
	if err := r.validateDependencies(); err != nil {
		return nil, err
	}

	cluster, err := r.ClusterGetter.ClusterARN(r.App, r.Env)
	if err != nil {
		return nil, fmt.Errorf("get cluster for environment %s: %w", r.Env, err)
	}

	filters := r.filtersForVPCFromAppEnv()

	subnets, err := r.VPCGetter.PublicSubnetIDs(filters...)
	if err != nil {
		return nil, fmt.Errorf(fmtErrPublicSubnetsFromEnv, r.Env, err)
	}
	if len(subnets) == 0 {
		return nil, errNoSubnetFound
	}

	// Use only environment security group https://github.com/aws/copilot-cli/issues/1882.
	securityGroups, err := r.VPCGetter.SecurityGroups(append(filters, ec2.Filter{
		Name:   fmt.Sprintf(ec2.TagFilterName, envSecurityGroupCFNLogicalIDTagKey),
		Values: []string{envSecurityGroupCFNLogicalIDTagValue},
	})...)
	if err != nil {
		return nil, fmt.Errorf(fmtErrSecurityGroupsFromEnv, r.Env, err)
	}

	ecsTasks, err := r.Starter.RunTask(ecs.RunTaskInput{
		Cluster:        cluster,
		Count:          r.Count,
		Subnets:        subnets,
		SecurityGroups: securityGroups,
		TaskFamilyName: taskFamilyName(r.GroupName),
		StartedBy:      startedBy,
	})
	if err != nil {
		return nil, &errRunTask{
			groupName: r.GroupName,
			parentErr: err,
		}
	}
	return convertECSTasks(ecsTasks), nil
}

func (r *EnvRunner) filtersForVPCFromAppEnv() []ec2.Filter {
	return []ec2.Filter{
		{
			Name:   tagFilterNameForEnv,
			Values: []string{r.Env},
		},
		{
			Name:   tagFilterNameForApp,
			Values: []string{r.App},
		},
	}
}

func (r *EnvRunner) validateDependencies() error {
	if r.VPCGetter == nil {
		return errVPCGetterNil
	}

	if r.ClusterGetter == nil {
		return errClusterGetterNil
	}

	if r.Starter == nil {
		return errStarterNil
	}

	return nil
}
