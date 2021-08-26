package docker

import (
	"context"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"log"
	"net/http"
	"sync"
)

func WaitingRoom(ctx context.Context, wg *sync.WaitGroup, mongoUrl string) (hostPort string, ipAddress string, err error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return "", "", err
	}
	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "ghcr.io/senergy-platform/device-waiting-room",
		Tag:        "dev",
		Env: []string{
			"MONGO_URL=" + mongoUrl,
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
	go Dockerlog(pool, ctx, container, "WAITING-ROOM")
	ipAddress = container.Container.NetworkSettings.IPAddress
	err = pool.Retry(func() error {
		log.Println("try waiting-room connection...")
		_, err := http.Get("http://" + ipAddress + ":8080/devices")
		return err
	})
	return hostPort, ipAddress, err
}
