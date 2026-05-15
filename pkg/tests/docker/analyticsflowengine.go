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

func AnalyticsFlowEngine(ctx context.Context, wg *sync.WaitGroup, pipelineApiUrl string, parserUrl string, rancherUrl string, mqttUrl string, permUrl string) (hostPort string, ipAddress string, err error) {
	log.Println("start analytics-flow-engine")
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "ghcr.io/senergy-platform/analytics-flow-engine",
			Env: map[string]string{
				"DRIVER":                  "rancher",
				"RANCHER2_ENDPOINT":       rancherUrl + "/",
				"PARSER_API_ENDPOINT":     parserUrl,
				"PIPELINE_API_ENDPOINT":   pipelineApiUrl,
				"BROKER_ADDRESS":          mqttUrl,
				"PERMISSION_API_ENDPOINT": permUrl,
				//"DEBUG":                 "true",
			},
			ExposedPorts:    []string{"8000/tcp"},
			WaitingFor:      wait.ForListeningPort("8000/tcp"),
			AlwaysPullImage: true,
			/*
				LogConsumerCfg: &testcontainers.LogConsumerConfig{
					Opts:      nil,
					Consumers: []testcontainers.LogConsumer{LogConsumer{Prefix: "ANALYTICS-FLOW-ENGINE:"}},
				},
				//*/
		},
		Started: true,
	})
	if err != nil {
		PrintDockerLogs(c, "ANALYTICS-FLOW-ENGINE")
		return "", "", err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		//PrintDockerLogs(c, "ANALYTICS-FLOW-ENGINE")
		timeout, _ := context.WithTimeout(context.Background(), 5*time.Second)
		log.Println("DEBUG: remove container analytics-flow-engine", c.Terminate(timeout))
	}()

	ipAddress, err = c.ContainerIP(ctx)
	if err != nil {
		return "", "", err
	}
	temp, err := c.MappedPort(ctx, "8000/tcp")
	if err != nil {
		return "", "", err
	}
	hostPort = temp.Port()

	return hostPort, ipAddress, err
}
