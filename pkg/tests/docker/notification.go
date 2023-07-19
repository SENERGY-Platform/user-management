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
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"sync"
	"time"
)

func NotificationContainer(ctx context.Context, wg *sync.WaitGroup, mongoIp string, vaultUrl string, keycloakUrl string) (hostPort string, ipAddress string, err error) {
	log.Println("start notifier")
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "ghcr.io/senergy-platform/notifier:dev",
			Env: map[string]string{
				"MONGO_ADDR":   mongoIp,
				"KEYCLOAK_URL": keycloakUrl,
				"VAULT_URL":    vaultUrl,
			},
			ExposedPorts:    []string{"5000/tcp"},
			WaitingFor:      wait.ForListeningPort("5000/tcp"),
			AlwaysPullImage: true,
			//SkipReaper:      true,
		},
		Started: true,
	})
	if err != nil {
		return "", "", err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		timeout, _ := context.WithTimeout(context.Background(), 5*time.Second)
		log.Println("DEBUG: remove container notifier", c.Terminate(timeout))
	}()
	//err = Dockerlog(ctx, c, "NOTIFIER")
	if err != nil {
		return "", "", err
	}

	ipAddress, err = c.ContainerIP(ctx)
	if err != nil {
		return "", "", err
	}
	temp, err := c.MappedPort(ctx, "5000/tcp")
	if err != nil {
		return "", "", err
	}
	hostPort = temp.Port()

	return hostPort, ipAddress, err
}
