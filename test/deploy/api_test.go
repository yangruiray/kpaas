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

package deploy

import (
	"context"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/util/wait"

	pb "github.com/kpaas-io/kpaas/pkg/deploy/protos"
	"github.com/kpaas-io/kpaas/pkg/deploy/server"
	"github.com/sirupsen/logrus"
)

var (
	client pb.DeployContollerClient
	conn   *grpc.ClientConn
	stopCh chan struct{}
)

func setup() {
	if _skip {
		return
	}

	logrus.SetLevel(logrus.DebugLevel)

	serverAddress := _remoteServerAddress

	if _launchLocalServer {
		var port uint16 = 9999
		// Setup and start gRpc server
		options := server.ServerOptions{
			Port:       port,
			LogFileLoc: "./tmp/logs",
		}
		stopCh = make(chan struct{})
		go server.New(options).Run(stopCh)
		serverAddress = fmt.Sprintf("localhost:%d", port)
	}

	// Create a gRpc client
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var err error
	conn, err = grpc.DialContext(ctx, serverAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		fmt.Println("did not connect:", err)
		os.Exit(1)
	}
	client = pb.NewDeployContollerClient(conn)
}

func tearDown() {
	if _skip {
		return
	}

	conn.Close()

	if _launchLocalServer {
		stopCh <- struct{}{}
	}
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func TestTestConnection(t *testing.T) {
	if _skip {
		t.SkipNow()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r, err := client.TestConnection(ctx, testConnectionData.request.(*pb.TestConnectionRequest))
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, true, r.Passed)
}

func TestCheckNodes(t *testing.T) {
	if _skip {
		t.SkipNow()
	}

	// CheckNodes request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r, err := client.CheckNodes(ctx, checkNodesData.request.(*pb.CheckNodesRequest))
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, true, r.Accepted)

	// GetCheckNodesResult request
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var actualReply *pb.GetCheckNodesResultReply
	// Call GetCheckNodesResult repeatly until the related task is done or failed.
	err = wait.Poll(10*time.Second, 1*time.Minute, func() (done bool, err error) {
		actualReply, err = client.GetCheckNodesResult(ctx, getCheckNodesResultData.request.(*pb.GetCheckNodesResultRequest))
		if err != nil {
			return false, err
		}
		if actualReply.Status == "failed" || actualReply.Status == "done" {
			return true, nil
		}
		return false, nil
	})
	assert.NoError(t, err)
	assert.NotNil(t, actualReply)
	sortCheckNodesResult(actualReply)
	expectedReply := getCheckNodesResultData.reply.(*pb.GetCheckNodesResultReply)
	sortCheckNodesResult(expectedReply)
	assert.Equal(t, expectedReply, actualReply)
}

func TestDeploy(t *testing.T) {
	if _skip {
		t.SkipNow()
	}

	// Deploy request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := client.Deploy(ctx, deployData.request.(*pb.DeployRequest))
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, true, res.Accepted)

	// GetDeployResult request
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var actualReply *pb.GetDeployResultReply
	// Call GetDeployResult repeatly until the related task is done or failed.
	err = wait.Poll(10*time.Second, 10*time.Minute, func() (done bool, err error) {
		actualReply, err := client.GetDeployResult(ctx, getDeployResultData.request.(*pb.GetDeployResultRequest))
		if err != nil {
			return false, err
		}
		if actualReply.Status == "failed" || actualReply.Status == "done" {
			return true, nil
		}
		return false, nil
	})
	assert.NoError(t, err)
	assert.NotNil(t, actualReply)
	sortDeployItemResults(actualReply.Items)
	expectedReply := getCheckNodesResultData.reply.(*pb.GetDeployResultReply)
	sortDeployItemResults(expectedReply.Items)
	assert.Equal(t, expectedReply, actualReply)
}

func sortItemCheckResults(results []*pb.ItemCheckResult) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].Item.Name <= results[j].Item.Name
	})
}

func sortCheckNodesResult(r *pb.GetCheckNodesResultReply) {
	for _, nodeCheckResult := range r.Nodes {
		sortItemCheckResults(nodeCheckResult.Items)
	}
}

func sortDeployItemResults(r []*pb.DeployItemResult) {
	sort.Slice(r, func(i, j int) bool {
		itemI := r[i].DeployItem
		itemJ := r[j].DeployItem
		// Sort by {NodeName, Role}
		if itemI.NodeName != itemJ.NodeName {
			return itemI.NodeName <= itemJ.NodeName
		}
		return itemI.Role <= itemJ.Role
	})
}