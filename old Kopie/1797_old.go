// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stack

import (
	"fmt"
	"strings"
)

const fmtTaskStackName = "task-%s"

// TaskStackName holds the name of a Copilot one-off task stack.
type TaskStackName string

// TaskName returns the name of the task family, generated from the stack name
func (t TaskStackName) TaskName() string {
	return strings.SplitN(string(t), "-", 2)[1]
}

// NameForService returns the stack name for a service.
func NameForService(app, env, svc string) string {
	// stack name limit constrained by CFN https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/cfn-using-console-create-stack-parameters.html
	const maxLen = 128
	stackName := fmt.Sprintf("%s-%s-%s", app, env, svc)

	if len(stackName) > maxLen {
		return stackName[:maxLen]
	}
	return stackName
}

// NameForEnv returns the stack name for an environment.
func NameForEnv(app, env string) string {
	return fmt.Sprintf("%s-%s", app, env)
}

// NameForTask returns the stack name for a task.
func NameForTask(task string) string {
	return fmt.Sprintf("task-%s", task)
	return TaskStackName(fmt.Sprintf(fmtTaskStackName, task))
}
