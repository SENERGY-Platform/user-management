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
	"errors"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func GetFreePort() (string, error) {
	portInt, err := getFreePort()
	if err != nil {
		return "", err
	}
	return strconv.Itoa(portInt), nil
}

func Dockerlog(pool *dockertest.Pool, ctx context.Context, repo *dockertest.Resource, name string) {
	out := &LogWriter{logger: log.New(os.Stdout, "["+name+"]", 0)}
	err := pool.Client.Logs(docker.LogsOptions{
		Stdout:       true,
		Stderr:       true,
		Context:      ctx,
		Container:    repo.Container.ID,
		Follow:       true,
		OutputStream: out,
		ErrorStream:  out,
	})
	if err != nil && err != context.Canceled {
		log.Println("DEBUG-ERROR: unable to start docker log", name, err)
	}
}

type LogWriter struct {
	logger *log.Logger
}

func (this *LogWriter) Write(p []byte) (n int, err error) {
	this.logger.Print(string(p))
	return len(p), nil
}

func GetHostIp() (string, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return "", err
	}
	networks, _ := pool.Client.ListNetworks()
	for _, network := range networks {
		if network.Name == "bridge" {
			return network.IPAM.Config[0].Gateway, nil
		}
	}
	return "", errors.New("no bridge network found")
}

//transform local-address to address in docker container
func LocalUrlToDockerUrl(localUrl string) (dockerUrl string, err error) {
	hostIp, err := GetHostIp()
	if err != nil {
		return "", err
	}
	urlStruct := strings.Split(localUrl, ":")
	dockerUrl = "http://" + hostIp + ":" + urlStruct[len(urlStruct)-1]
	log.Println("DEBUG: url transformation:", localUrl, "-->", dockerUrl)
	return
}
