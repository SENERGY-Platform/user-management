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
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"log"
	"net/http"
	"sync"
)

func DatabaseExports(ctx context.Context, wg *sync.WaitGroup, mysqlHost string, rancherUrl string, permSearchUrl string, influxDbHost string) (hostPort string, ipAddress string, err error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return "", "", err
	}
	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "ghcr.io/senergy-platform/analytics-serving-service",
		Tag:        "dev",
		Env: []string{
			"MYSQL_HOST=" + mysqlHost,
			"MYSQL_USER=root",
			"MYSQL_PW=secret",
			"MYSQL_DB=mysql",
			"DOCKER_PULL=true",
			"DRIVER=rancher2",
			"TRANSFER_IMAGE=ghcr.io/senergy-platform/hello-world:test",
			"RANCHER2_ENDPOINT=" + rancherUrl + "/",
			"RANCHER_ACCESS_KEY=foo",
			"RANCHER_SECRET_KEY=bar",
			"PERMISSION_API_ENDPOINT=" + permSearchUrl,
			"API_PORT=8080",
			"INFLUX_DB_HOST=" + influxDbHost,
		},
	}, func(config *docker.HostConfig) {
	})
	if err != nil {
		return "", "", err
	}
	wg.Add(1)
	go func() {
		<-ctx.Done()
		log.Println("DEBUG: remove container " + container.Container.Name)
		container.Close()
		wg.Done()
	}()
	hostPort = container.GetPort("8080/tcp")
	go Dockerlog(pool, ctx, container, "DATABASE-EXPORT")
	ipAddress = container.Container.NetworkSettings.IPAddress
	err = pool.Retry(func() error {
		log.Println("try DatabaseExports connection...")
		_, err := http.Get("http://" + ipAddress + ":8080/instances")
		return err
	})
	return hostPort, ipAddress, err
}