package docker

import (
	"context"
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"github.com/SENERGY-Platform/user-management/pkg/tests/mocks"
	"sync"
)

func Start(basectx context.Context, wg *sync.WaitGroup, origConfig configuration.Config) (config configuration.Config, err error) {
	config = origConfig
	ctx, cancel := context.WithCancel(basectx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	_, zkIp, err := Zookeeper(ctx, wg)
	if err != nil {
		return config, err
	}
	zookeeperUrl := zkIp + ":2181"

	config.KafkaBootstrap, err = Kafka(ctx, wg, zookeeperUrl)
	if err != nil {
		return config, err
	}

	_, wrDbIp, err := MongoContainer(ctx, wg)
	if err != nil {
		return config, err
	}
	_, wrIp, err := WaitingRoom(ctx, wg, "mongodb://"+wrDbIp+":27017")
	if err != nil {
		return config, err
	}
	config.WaitingRoomUrl = "http://" + wrIp + ":8080"

	_, psDbIp, err := MongoContainer(ctx, wg)
	if err != nil {
		return config, err
	}
	_, psIp, err := ProcessScheduler(ctx, wg, "mongodb://"+psDbIp+":27017")
	if err != nil {
		return config, err
	}
	config.ProcessSchedulerUrl = "http://" + psIp + ":8080"

	_, dDbIp, err := MongoContainer(ctx, wg)
	if err != nil {
		return config, err
	}
	_, dIp, err := Dashboard(ctx, wg, dDbIp)
	if err != nil {
		return config, err
	}
	config.DashboardServiceUrl = "http://" + dIp + ":8080"

	_, importsDbIp, err := MongoContainer(ctx, wg)
	if err != nil {
		return config, err
	}
	importsDbUrl := "mongodb://" + importsDbIp + ":27017"
	permissionsUrl, err := LocalUrlToDockerUrl(mocks.PermissionsMock(ctx, wg))
	if err != nil {
		return config, err
	}
	importRepoUrl, err := LocalUrlToDockerUrl(mocks.ImportsRepoMock(ctx, wg))
	if err != nil {
		return config, err
	}

	_, importsIp, err := Imports(ctx, wg, importsDbUrl, importRepoUrl, permissionsUrl, config.KafkaBootstrap)
	if err != nil {
		return config, err
	}

	config.ImportsDeploymentUrl = "http://" + importsIp + ":8080"

	return config, nil
}
