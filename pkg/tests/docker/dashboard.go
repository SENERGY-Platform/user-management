package docker

import (
	"context"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"log"
	"net/http"
	"sync"
)

func Dashboard(ctx context.Context, wg *sync.WaitGroup, mongoIp string) (hostPort string, ipAddress string, err error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return "", "", err
	}
	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "fgseitsrancher.wifa.intern.uni-leipzig.de:5000/dashboard",
		Tag:        "dev",
		Env: []string{
			"MONGO=" + mongoIp,
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
	ipAddress = container.Container.NetworkSettings.IPAddress
	err = pool.Retry(func() error {
		log.Println("try dashboard connection...")
		_, err := http.Get("http://" + ipAddress + ":8080/dashboards")
		return err
	})
	return hostPort, ipAddress, err
}
