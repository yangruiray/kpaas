// Copyright 2019 Shanghai JingDuo Information Technology co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package init

import (
	"bytes"
	"fmt"

	"github.com/kpaas-io/kpaas/pkg/deploy/command"
	"github.com/kpaas-io/kpaas/pkg/deploy/machine"
	"github.com/kpaas-io/kpaas/pkg/deploy/operation"
	pb "github.com/kpaas-io/kpaas/pkg/deploy/protos"
)

type InitHostNameOperation struct {
	shellCmd       *command.ShellCommand
	NodeInitAction *operation.NodeInitAction
}

func (itOps *InitHostNameOperation) RunCommands(node *pb.Node, initAction *operation.NodeInitAction, logChan chan<- *bytes.Buffer) (stdOut, stdErr []byte, err error) {

	m, err := machine.NewMachine(node)
	if err != nil {
		return nil, nil, err
	}

	defer m.Close()

	logBuffer := &bytes.Buffer{}

	itOps.NodeInitAction = initAction

	currentName := node.Name
	if currentName == "" {
		return nil, nil, fmt.Errorf("node name can not be empty")
	}

	itOps.shellCmd = command.NewShellCommand(m, "hostnamectl", fmt.Sprintf("set-hostname %v", currentName)).
		WithDescription("初始化 host 文件").
		WithExecuteLogWriter(logBuffer)

	// run commands
	stdOut, stdErr, err = itOps.shellCmd.Execute()

	// write to log channel
	logChan <- logBuffer

	return
}
