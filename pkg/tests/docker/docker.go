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
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"github.com/SENERGY-Platform/permissions-v2/pkg/model"
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"github.com/SENERGY-Platform/user-management/pkg/tests/mocks"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
)

func Start(basectx context.Context, wg *sync.WaitGroup, origConfig configuration.Config) (config configuration.Config, getDeviceRepoCalls func() []string, err error) {
	config = origConfig
	ctx, cancel := context.WithCancel(basectx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	_, zkIp, err := Zookeeper(ctx, wg)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	zookeeperUrl := zkIp + ":2181"

	config.KafkaBootstrap, err = Kafka(ctx, wg, zookeeperUrl)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}

	_, wrDbIp, err := MongoContainer(ctx, wg)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	_, wrIp, err := WaitingRoom(ctx, wg, "mongodb://"+wrDbIp+":27017")
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	config.WaitingRoomUrl = "http://" + wrIp + ":8080"

	_, psDbIp, err := MongoContainer(ctx, wg)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	_, psIp, err := ProcessScheduler(ctx, wg, "mongodb://"+psDbIp+":27017")
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	config.ProcessSchedulerUrl = "http://" + psIp + ":8080"
	_, dDbIp, err := MongoContainer(ctx, wg)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	_, dIp, err := Dashboard(ctx, wg, dDbIp)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	config.DashboardServiceUrl = "http://" + dIp + ":8080"

	_, importsDbIp, err := MongoContainer(ctx, wg)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	importsDbUrl := "mongodb://" + importsDbIp + ":27017"
	permissionsUrl, err := LocalUrlToDockerUrl(mocks.PermissionsMock(ctx, wg))
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	importRepoUrl, err := LocalUrlToDockerUrl(mocks.ImportsRepoMock(ctx, wg))
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	rancherUrl, err := LocalUrlToDockerUrl(mocks.RancherMock(ctx, wg))
	if err != nil {
		return config, getDeviceRepoCalls, err
	}

	_, permV2Ip, err := PermissionsV2(ctx, wg, importsDbUrl)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	permissionsV2Url := "http://" + permV2Ip + ":8080"

	_, importsIp, err := Imports(ctx, wg, importsDbUrl, importRepoUrl, permissionsUrl, config.KafkaBootstrap, rancherUrl, permissionsV2Url)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	config.ImportsDeploymentUrl = "http://" + importsIp + ":8080"

	_, brokerExportsDbIp, err := MongoContainer(ctx, wg)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	brokerExportsDbUrl := "mongodb://" + brokerExportsDbIp + ":27017"

	_, brokerExportsIp, err := BrokerExports(ctx, wg, brokerExportsDbUrl, rancherUrl, permissionsV2Url)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	config.BrokerExportsUrl = "http://" + brokerExportsIp + ":8080"

	_, dbExportsDbIp, err := MysqlContainer(ctx, wg)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	_, influxIp, err := InfluxdbContainer(ctx, wg)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	_, dbExportsIp, err := DatabaseExports(ctx, wg, dbExportsDbIp, rancherUrl, permissionsV2Url, influxIp)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	config.DatabaseExportsUrl = "http://" + dbExportsIp + ":8080"
	log.Println("DatabaseExportsUrl = ", config.DatabaseExportsUrl)

	_, operatorDbIp, err := MongoContainer(ctx, wg)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	_, operatorIp, err := AnalyticsOperatorRepo(ctx, wg, operatorDbIp)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	config.AnalyticsOperatorRepoUrl = "http://" + operatorIp + ":5000"

	_, flowDbIp, err := MongoContainer(ctx, wg)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	_, flowIp, err := AnalyticsFlowRepo(ctx, wg, flowDbIp)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	config.AnalyticsFlowRepoUrl = "http://" + flowIp + ":5000"

	_, pipelineDbIp, err := MongoContainer(ctx, wg)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	_, pipelineIp, err := AnalyticsPipeline(ctx, wg, pipelineDbIp)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	config.AnalyticsPipelineUrl = "http://" + pipelineIp + ":8000"

	parserMockUrl, err := LocalUrlToDockerUrl(mocks.AnalyticsParserMock(ctx, wg))
	if err != nil {
		return config, getDeviceRepoCalls, err
	}

	_, mqttIp, err := Mqtt(ctx, wg)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}

	_, engineIp, err := AnalyticsFlowEngine(ctx, wg, config.AnalyticsPipelineUrl, parserMockUrl, rancherUrl, "tcp://"+mqttIp+":1883")
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	config.AnalyticsFlowEngineUrl = "http://" + engineIp + ":8000"

	_, notifierDbIp, err := MongoContainer(ctx, wg)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}

	vaultUrl, err := mocks.MockVault(ctx)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	vaultUrl, err = LocalUrlToDockerUrl(vaultUrl)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}

	keycloakUrl, err := mocks.MockKeycloak(ctx)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	keycloakUrl, err = LocalUrlToDockerUrl(keycloakUrl)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}

	_, notifierIp, err := NotificationContainer(ctx, wg, notifierDbIp, vaultUrl, keycloakUrl)
	if err != nil {
		return config, getDeviceRepoCalls, err
	}
	config.NotifierUrl = "http://" + notifierIp + ":5000"

	deviceRepoCalls := []string{}
	getDeviceRepoCalls = func() []string {
		return deviceRepoCalls
	}

	deviceRepoMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deviceRepoCalls = append(deviceRepoCalls, r.Method+" "+r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer deviceRepoMock.Close()
		<-ctx.Done()
	}()
	config.DeviceRepositoryUrl = deviceRepoMock.URL

	_, err, _ = client.New(permissionsV2Url).SetTopic(client.InternalAdminToken, client.Topic{
		Id: "import-instances",
		DefaultPermissions: model.ResourcePermissions{
			UserPermissions: map[string]model.PermissionsMap{
				"admin": {Read: true, Write: true, Execute: true, Administrate: true},
			},
		},
	})
	if err != nil {
		return config, getDeviceRepoCalls, err
	}

	_, err, _ = client.New(permissionsV2Url).SetTopic(client.InternalAdminToken, client.Topic{
		Id: "import-types",
		DefaultPermissions: model.ResourcePermissions{
			UserPermissions: map[string]model.PermissionsMap{
				"user1": {Read: true, Write: true, Execute: true, Administrate: true},
				"user2": {Read: true, Write: true, Execute: true, Administrate: true},
			},
		},
	})
	if err != nil {
		return config, getDeviceRepoCalls, err
	}

	_, err, _ = client.New(permissionsV2Url).SetTopic(client.InternalAdminToken, client.Topic{
		Id: "export-instances",
		DefaultPermissions: model.ResourcePermissions{
			UserPermissions: map[string]model.PermissionsMap{
				"admin": {Read: true, Write: true, Execute: true, Administrate: true},
			},
		},
	})
	if err != nil {
		return config, getDeviceRepoCalls, err
	}

	_, err, _ = client.New(permissionsV2Url).SetTopic(client.InternalAdminToken, client.Topic{
		Id: "devices",
		DefaultPermissions: model.ResourcePermissions{
			UserPermissions: map[string]model.PermissionsMap{
				"user1": {Read: true, Write: true, Execute: true, Administrate: true},
				"user2": {Read: true, Write: true, Execute: true, Administrate: true},
			},
		},
	})
	if err != nil {
		return config, getDeviceRepoCalls, err
	}

	_, err, _ = client.New(permissionsV2Url).SetPermission(client.InternalAdminToken, "devices", "1", client.ResourcePermissions{
		UserPermissions: map[string]model.PermissionsMap{
			"admin": {Read: true, Write: true, Execute: true, Administrate: true},
		},
	})
	if err != nil {
		return config, getDeviceRepoCalls, err
	}

	_, err, _ = client.New(permissionsV2Url).SetPermission(client.InternalAdminToken, "devices", "2", client.ResourcePermissions{
		UserPermissions: map[string]model.PermissionsMap{
			"admin": {Read: true, Write: true, Execute: true, Administrate: true},
		},
	})
	if err != nil {
		return config, getDeviceRepoCalls, err
	}

	_, err, _ = client.New(permissionsV2Url).SetPermission(client.InternalAdminToken, "devices", "3", client.ResourcePermissions{
		UserPermissions: map[string]model.PermissionsMap{
			"admin": {Read: true, Write: true, Execute: true, Administrate: true},
		},
	})
	if err != nil {
		return config, getDeviceRepoCalls, err
	}

	return config, getDeviceRepoCalls, nil
}
