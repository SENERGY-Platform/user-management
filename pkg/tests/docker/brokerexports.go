package docker

import (
	"context"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"log"
	"net/http"
	"sync"
)

func BrokerExports(ctx context.Context, wg *sync.WaitGroup, mongoUrl string, rancherUrl string) (hostPort string, ipAddress string, err error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return "", "", err
	}
	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "ghcr.io/senergy-platform/kafka2mqtt-manager",
		Tag:        "dev",
		Env: []string{
			"MONGO_URL=" + mongoUrl,
			"MONGO_REPL_SET=false",
			"DEBUG=true",
			"DOCKER_PULL=true",
			"VERIFY_INPUT=false",
			"TRANSFER_IMAGE=ghcr.io/senergy-platform/hello-world:test",
			"DEPLOY_MODE=rancher2",
			"RANCHER_URL=" + rancherUrl + "/",
			"RANCHER_ACCESS_KEY=foo",
			"RANCHER_SECRET_KEY=bar",
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
	go Dockerlog(pool, ctx, container, "BROKER-EXPORT")
	ipAddress = container.Container.NetworkSettings.IPAddress
	err = pool.Retry(func() error {
		log.Println("try BrokerExports connection...")
		_, err := http.Get("http://" + ipAddress + ":8080/instances")
		return err
	})
	return hostPort, ipAddress, err
}
