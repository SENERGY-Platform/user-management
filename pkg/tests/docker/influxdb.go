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
	"github.com/influxdata/influxdb/client/v2"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"log"
	"sync"
	"time"
)

func InfluxdbContainer(ctx context.Context, wg *sync.WaitGroup) (hostPort string, ipAddress string, err error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return "", "", err
	}
	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "influxdb",
		Tag:        "1.6.3",
		Env: []string{
			"INFLUXDB_DB=connectionlog",
			"INFLUXDB_ADMIN_ENABLED=true",
			"INFLUXDB_ADMIN_USER=root",
			"INFLUXDB_ADMIN_PASSWORD=",
		},
	}, func(config *docker.HostConfig) {})
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
	hostPort = container.GetPort("8086/tcp")
	err = pool.Retry(func() error {
		log.Println("try InfluxdbContainer connection...")
		client, err := client.NewHTTPClient(client.HTTPConfig{
			Addr:     "http://" + container.Container.NetworkSettings.IPAddress + ":8086",
			Username: "root",
			Password: "",
			Timeout:  time.Duration(1) * time.Second,
		})
		if err != nil {
			log.Println(err)
			return err
		}
		defer client.Close()
		_, _, err = client.Ping(1 * time.Second)
		if err != nil {
			log.Println(err)
			return err
		}
		return err
	})
	return hostPort, container.Container.NetworkSettings.IPAddress, err
}
