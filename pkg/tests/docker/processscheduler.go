package docker

import (
	"context"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"log"
	"net/http"
	"sync"
)

func ProcessScheduler(ctx context.Context, wg *sync.WaitGroup, mongoUrl string) (hostPort string, ipAddress string, err error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return "", "", err
	}
	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "ghcr.io/senergy-platform/process-scheduler",
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
	ipAddress = container.Container.NetworkSettings.IPAddress
	go Dockerlog(pool, ctx, container, "SCHEDULER")
	err = pool.Retry(func() error {
		log.Println("try scheduler connection...")
		_, err := http.Get("http://" + ipAddress + ":8080/schedules")
		return err
	})
	return hostPort, ipAddress, err
}
