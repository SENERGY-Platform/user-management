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
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type LogConsumer struct {
	Prefix string
}

func (this LogConsumer) Accept(log testcontainers.Log) {
	fmt.Print(this.Prefix, string(log.Content))
}

func DatabaseExports(ctx context.Context, wg *sync.WaitGroup, mysqlHost string, rancherUrl string, permSearchUrl string, influxDbHost string, kafkaUrl string, importDeployUrl string, pipelineApiUrl string) (hostPort string, ipAddress string, err error) {
	log.Println("start analytics-serving-service")
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "ghcr.io/senergy-platform/analytics-serving-service:prod",
			Env: map[string]string{
				"MYSQL_HOST":                 mysqlHost,
				"MYSQL_USER":                 "root",
				"MYSQL_PW":                   "secret",
				"MYSQL_DB":                   "mysql",
				"DOCKER_PULL":                "true",
				"DRIVER":                     "ew",
				"TRANSFER_IMAGE":             "ghcr.io/senergy-platform/hello-world:test",
				"RANCHER2_ENDPOINT":          rancherUrl + "/",
				"RANCHER_ACCESS_KEY":         "foo",
				"RANCHER_SECRET_KEY":         "bar",
				"PERMISSION_V2_URL":          permSearchUrl,
				"API_PORT":                   "8080",
				"SERVER_PORT":                "8080",
				"INFLUX_DB_HOST":             influxDbHost,
				"KAFKA_BOOTSTRAP":            kafkaUrl,
				"PIPELINE_API_ENDPOINT":      pipelineApiUrl,
				"IMPORT_DEPLOY_API_ENDPOINT": importDeployUrl,
				"EXPORT_DATABASE_ID_PREFIX":  "urn:infai:ses:export-db:",
				"KAFKA_REPLICATION_FACTOR":   "1",
			},
			ExposedPorts:    []string{"8080/tcp"},
			WaitingFor:      wait.ForListeningPort("8080/tcp"),
			AlwaysPullImage: true,
			/*
				LogConsumerCfg: &testcontainers.LogConsumerConfig{
					Opts:      nil,
					Consumers: []testcontainers.LogConsumer{LogConsumer{Prefix: "ANALYTICS-SERVING-SERVICE:"}},
				},
				//*/
		},
		Started: true,
	})
	if err != nil {
		PrintDockerLogs(c, "ANALYTICS-SERVING-SERVICE")
		return "", "", err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		//PrintDockerLogs(c, "ANALYTICS-SERVING-SERVICE")
		timeout, _ := context.WithTimeout(context.Background(), 5*time.Second)
		log.Println("DEBUG: remove container analytics-serving-service", c.Terminate(timeout))
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
