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
	rancherUrl, err := LocalUrlToDockerUrl(mocks.RancherMock(ctx, wg))
	if err != nil {
		return config, err
	}

	_, importsIp, err := Imports(ctx, wg, importsDbUrl, importRepoUrl, permissionsUrl, config.KafkaBootstrap, rancherUrl)
	if err != nil {
		return config, err
	}
	config.ImportsDeploymentUrl = "http://" + importsIp + ":8080"

	_, brokerExportsDbIp, err := MongoContainer(ctx, wg)
	if err != nil {
		return config, err
	}
	brokerExportsDbUrl := "mongodb://" + brokerExportsDbIp + ":27017"

	_, brokerExportsIp, err := BrokerExports(ctx, wg, brokerExportsDbUrl, rancherUrl)
	if err != nil {
		return config, err
	}
	config.BrokerExportsUrl = "http://" + brokerExportsIp + ":8080"

	_, dbExportsDbIp, err := MysqlContainer(ctx, wg)
	if err != nil {
		return config, err
	}
	_, influxIp, err := InfluxdbContainer(ctx, wg)
	if err != nil {
		return config, err
	}
	_, dbExportsIp, err := DatabaseExports(ctx, wg, dbExportsDbIp, rancherUrl, permissionsUrl, influxIp)
	if err != nil {
		return config, err
	}
	config.DatabaseExportsUrl = "http://" + dbExportsIp + ":8080"

	_, operatorDbIp, err := MongoContainer(ctx, wg)
	if err != nil {
		return config, err
	}
	_, operatorIp, err := AnalyticsOperatorRepo(ctx, wg, operatorDbIp)
	if err != nil {
		return config, err
	}
	config.AnalyticsOperatorRepoUrl = "http://" + operatorIp + ":5000"

	_, flowDbIp, err := MongoContainer(ctx, wg)
	if err != nil {
		return config, err
	}
	_, flowIp, err := AnalyticsFlowRepo(ctx, wg, flowDbIp)
	if err != nil {
		return config, err
	}
	config.AnalyticsFlowRepoUrl = "http://" + flowIp + ":5000"

	return config, nil
}
