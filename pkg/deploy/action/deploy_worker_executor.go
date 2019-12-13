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

package action

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/kpaas-io/kpaas/pkg/deploy/consts"
	deployMachine "github.com/kpaas-io/kpaas/pkg/deploy/machine"
	"github.com/kpaas-io/kpaas/pkg/deploy/operation/worker"
	"github.com/kpaas-io/kpaas/pkg/deploy/protos"
)

type deployWorkerExecutor struct {
	logger  *logrus.Entry
	machine *deployMachine.Machine
	action  *DeployWorkerAction
}

func (executor *deployWorkerExecutor) Execute(act Action) *protos.Error {

	action, ok := act.(*DeployWorkerAction)
	if !ok {
		return errOfTypeMismatched(new(DeployWorkerAction), act)
	}

	executor.action = action

	executor.initLogger()

	executor.logger.Info("start to execute deploy worker executor")

	if err := executor.connectSSH(); err != nil {

		return err
	}
	defer executor.disconnectSSH()

	if err := executor.installKubelet(); err != nil {
		return err
	}

	if err := executor.joinCluster(); err != nil {
		return err
	}

	if err := executor.appendLabel(); err != nil {
		return err
	}

	if err := executor.appendAnnotation(); err != nil {
		return err
	}

	if err := executor.appendTaint(); err != nil {
		return err
	}

	executor.logger.Info("deploy worker finished")

	return nil
}

func (executor *deployWorkerExecutor) connectSSH() *protos.Error {

	executor.logger.Debug("Start to connect ssh")

	var err error
	executor.machine, err = deployMachine.NewMachine(executor.action.config.Node.GetNode())
	if err != nil {
		pbError := &protos.Error{
			Reason:     "Connect ssh error",                                                                                                                                           // 连接SSH失败。
			Detail:     fmt.Sprintf("SSH connect to %s(%s) failed , error: %v.", executor.action.config.Node.GetNode().GetName(), executor.action.config.Node.GetNode().GetIp(), err), // 连接%s(%s)失败，失败原因：%v。
			FixMethods: "Please check node reliability, make SSH service is available.",                                                                                               // 请检查节点的可用性，确保SSH服务可用。
		}
		executor.logger.WithField("error", pbError).Error("connect ssh error")
		return pbError
	}

	executor.logger.Debug("ssh connected")
	return nil
}

func (executor *deployWorkerExecutor) initLogger() {
	executor.logger = logrus.WithFields(logrus.Fields{
		consts.LogFieldAction: executor.action.GetName(),
		"nodeName":            executor.action.config.Node.GetNode().GetName(),
		"nodeIP":              executor.action.config.Node.GetNode().GetIp(),
	})
}

func (executor *deployWorkerExecutor) installKubelet() *protos.Error {

	executor.logger.Debug("Start to install kubelet")

	operation := worker.NewInstallKubelet(
		&worker.InstallKubeletConfig{
			Machine: executor.machine,
			Logger:  executor.logger,
			Node:    executor.action.config.Node,
			Cluster: executor.action.config.ClusterConfig,
		},
	)

	if err := operation.Execute(); err != nil {
		executor.logger.WithField("error", err).Error("install kubelet error")
		return err
	}

	executor.logger.Info("Finish to install kubelet action")
	return nil
}

func (executor *deployWorkerExecutor) joinCluster() *protos.Error {

	executor.logger.Debug("Start to join cluster")

	operation := worker.NewJoinCluster(
		&worker.JoinClusterConfig{
			Machine: executor.machine,
			Logger:  executor.logger,
			Node:    executor.action.config.Node,
			Cluster: executor.action.config.ClusterConfig,
		},
	)

	if err := operation.Execute(); err != nil {
		executor.logger.WithField("error", err).Error("join cluster error")
		return err
	}

	executor.logger.Info("Finish to join cluster action")
	return nil
}

func (executor *deployWorkerExecutor) appendLabel() *protos.Error {

	executor.logger.Debug("Start to append label")

	operation := worker.NewAppendLabel(
		&worker.AppendLabelConfig{
			Machine: executor.machine,
			Logger:  executor.logger,
			Node:    executor.action.config.Node,
			Cluster: executor.action.config.ClusterConfig,
		},
	)

	if err := operation.Execute(); err != nil {
		executor.logger.WithField("error", err).Error("append label error")
		return err
	}

	executor.logger.Info("Finish to append label action")
	return nil
}

func (executor *deployWorkerExecutor) appendAnnotation() *protos.Error {

	executor.logger.Debug("Start to append annotation")

	operation := worker.NewAppendAnnotation(
		&worker.AppendAnnotationConfig{
			Machine: executor.machine,
			Logger:  executor.logger,
			Node:    executor.action.config.Node,
			Cluster: executor.action.config.ClusterConfig,
		},
	)

	if err := operation.Execute(); err != nil {
		executor.logger.WithField("error", err).Error("append annotation error")
		return err
	}

	executor.logger.Info("Finish to append annotation action")
	return nil
}

func (executor *deployWorkerExecutor) appendTaint() *protos.Error {

	executor.logger.Debug("Start to append taint")

	operation := worker.NewAppendTaint(
		&worker.AppendTaintConfig{
			Machine: executor.machine,
			Logger:  executor.logger,
			Node:    executor.action.config.Node,
			Cluster: executor.action.config.ClusterConfig,
		},
	)

	if err := operation.Execute(); err != nil {
		executor.logger.WithField("error", err).Error("append taint error")
		return err
	}

	executor.logger.Info("Finish to append taint action")
	return nil
}

func (executor *deployWorkerExecutor) disconnectSSH() {

	executor.logger.Debug("Start to disconnect ssh")

	executor.machine.Close()

	executor.logger.Debug("ssh disconnected")
}