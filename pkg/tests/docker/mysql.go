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
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"log"
	"sync"
)

func MysqlContainer(ctx context.Context, wg *sync.WaitGroup) (hostPort string, ipAddress string, err error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return "", "", err
	}
	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mariadb",
		Tag:        "10.5",
		Env: []string{
			"MYSQL_ROOT_PASSWORD=secret",
			"MYSQL_DATABASE=mysql",
		},
	}, func(config *docker.HostConfig) {
		config.Tmpfs = map[string]string{"/var/lib/mysql": "rw"}
	})
	if err != nil {
		return "", "", err
	}
	//go Dockerlog(pool, ctx, container, "MY-SQL")
	wg.Add(1)
	go func() {
		<-ctx.Done()
		log.Println("DEBUG: remove container " + container.Container.Name)
		container.Close()
		wg.Done()
	}()
	hostPort = container.GetPort("3306/tcp")
	conStr := fmt.Sprintf("root:secret@(localhost:%s)/mysql?parseTime=true", hostPort)
	err = pool.Retry(func() error {
		log.Println("try mysql connection...")
		var err error
		db, err := sql.Open("mysql", conStr)
		if err != nil {
			log.Println("ERROR:", err)
			return err
		}
		err = db.Ping()
		if err != nil {
			log.Println("ERROR:", err)
			return err
		}
		return nil
	})
	return hostPort, container.Container.NetworkSettings.IPAddress, err
}
