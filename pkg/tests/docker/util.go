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
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func PrintDockerLogs(c testcontainers.Container, name string) {
	if c == nil {
		return
	}
	reader, err2 := c.Logs(context.Background())
	if err2 != nil {
		log.Println("ERROR: unable to get container log", name)
		return
	}
	buf := new(strings.Builder)
	io.Copy(buf, reader)
	fmt.Println(name + " LOGS: ------------------------------------------")
	fmt.Println(buf.String())
	fmt.Println("\n---------------------------------------------------------------")
}

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

func Dockerlog(container testcontainers.Container, name string) error {
	container.FollowOutput(&LogWriter{logger: log.New(os.Stdout, "["+name+"] ", log.LstdFlags)})
	err := container.StartLogProducer(context.Background())
	if err != nil {
		return err
	}
	return nil
}

type LogWriter struct {
	logger *log.Logger
}

func (this *LogWriter) Accept(l testcontainers.Log) {
	this.Write(l.Content)
}

func (this *LogWriter) Write(p []byte) (n int, err error) {
	this.logger.Print(string(p))
	return len(p), nil
}

func waitretry(timeout time.Duration, f func(ctx context.Context, target wait.StrategyTarget) error) func(ctx context.Context, target wait.StrategyTarget) error {
	return func(ctx context.Context, target wait.StrategyTarget) (err error) {
		return retry(timeout, func() error {
			return f(ctx, target)
		})
	}
}

func retry(timeout time.Duration, f func() error) (err error) {
	err = errors.New("initial")
	start := time.Now()
	for i := int64(1); err != nil && time.Since(start) < timeout; i++ {
		err = f()
		if err != nil {
			log.Println("ERROR: :", err)
			wait := time.Duration(i) * time.Second
			if time.Since(start)+wait < timeout {
				log.Println("ERROR: retry after:", wait.String())
				time.Sleep(wait)
			} else {
				time.Sleep(time.Since(start) + wait - timeout)
				return f()
			}
		}
	}
	return err
}

func Forward(ctx context.Context, fromPort int, toAddr string) error {
	log.Println("forward", fromPort, "to", toAddr)
	incoming, err := net.Listen("tcp", fmt.Sprintf(":%d", fromPort))
	if err != nil {
		return err
	}
	go func() {
		defer log.Println("closed forward incoming")
		<-ctx.Done()
		incoming.Close()
	}()
	go func() {
		for {
			client, err := incoming.Accept()
			if err != nil {
				log.Println("FORWARD ERROR:", err)
				return
			}
			go handleForwardClient(client, toAddr)
		}
	}()
	return nil
}

func handleForwardClient(client net.Conn, addr string) {
	//log.Println("new forward client")
	target, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println("FORWARD ERROR:", err)
		return
	}
	go func() {
		defer target.Close()
		defer client.Close()
		io.Copy(target, client)
	}()
	go func() {
		defer target.Close()
		defer client.Close()
		io.Copy(client, target)
	}()
}

func GetHostIp() (string, error) {
	provider, err := testcontainers.NewDockerProvider(testcontainers.DefaultNetwork("bridge"))
	if err != nil {
		return "", err
	}
	return provider.GetGatewayIP(context.Background())
}

// transform local-address to address in docker container
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
