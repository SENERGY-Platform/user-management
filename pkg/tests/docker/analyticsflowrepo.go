/*
 * Copyright 2021 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package docker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func AnalyticsFlowRepo(ctx context.Context, wg *sync.WaitGroup, mongoIp string, operatorRepoUrl string, permUrl string, piplelineRegUrl string) (hostPort string, ipAddress string, err error) {
	log.Println("start analytics-flow-repo")
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "ghcr.io/senergy-platform/analytics-flow-repo-v2:latest",
			Env: map[string]string{
				"MONGO_URL":             mongoIp,
				"OPERATOR_REPO_URL":     operatorRepoUrl,
				"PERMISSIONS_V2_URL":    permUrl,
				"PIPELINE_REGISTRY_URL": piplelineRegUrl,
			},
			ExposedPorts:    []string{"8080/tcp"},
			WaitingFor:      wait.ForListeningPort("8080/tcp"),
			AlwaysPullImage: true,

			/*
				LogConsumerCfg: &testcontainers.LogConsumerConfig{
					Opts:      nil,
					Consumers: []testcontainers.LogConsumer{LogConsumer{Prefix: "ANALYTICS-FLOW-REPO:"}},
				},
			//*/
		},
		Started: true,
	})
	if err != nil {
		PrintDockerLogs(c, "ANALYTICS-FLOW-REPO")
		return "", "", err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		timeout, _ := context.WithTimeout(context.Background(), 5*time.Second)
		log.Println("DEBUG: remove container analytics-flow-repo", c.Terminate(timeout))
	}()

	ipAddress, err = c.ContainerIP(ctx)
	if err != nil {
		return "", "", err
	}
	temp, err := c.MappedPort(ctx, "8080/tcp")
	if err != nil {
		return "", "", err
	}
	hostPort = temp.Port()

	return hostPort, ipAddress, err
}
