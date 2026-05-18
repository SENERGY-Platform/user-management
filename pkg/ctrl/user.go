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

package ctrl

import (
	devicerepo "github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
)

func DeleteUser(userId string, conf configuration.Config) (err error) {
	token, err := CreateToken("users-service", userId)
	if err != nil {
		conf.GetLogger().Error("unable to create jwt for userId", "userId", userId, "error", err)
		return err
	}
	err, _ = devicerepo.NewClient(conf.DeviceRepositoryUrl, nil).DeleteUser(devicerepo.InternalAdminToken, userId)
	if err != nil {
		conf.GetLogger().Error("unable to delete user from device-repository", "userId", userId, "error", err)
		return err
	}
	err = DeleteWaitingRoomUser(token, conf)
	if err != nil {
		conf.GetLogger().Error("unable to delete user from waiting-room", "userId", userId, "error", err)
		return err
	}
	err = DeleteDashboardUser(token, conf)
	if err != nil {
		conf.GetLogger().Error("unable to delete user from dashboard-service", "userId", userId, "error", err)
		return err
	}
	err = DeleteProcessSchedulerUser(token, conf)
	if err != nil {
		conf.GetLogger().Error("unable to delete user from process-scheduler", "userId", userId, "error", err)
		return err
	}
	err = DeleteImportsUser(token, conf)
	if err != nil {
		conf.GetLogger().Error("unable to delete user from imports-service", "userId", userId, "error", err)
		return err
	}
	err = DeleteBrokerExportsUser(token, conf)
	if err != nil {
		conf.GetLogger().Error("unable to delete user from broker-exports-service", "userId", userId, "error", err)
		return err
	}
	err = DeleteDatabaseExportsUser(token, conf)
	if err != nil {
		conf.GetLogger().Error("unable to delete user from database-exports-service", "userId", userId, "error", err)
		return err
	}

	if conf.RemoveExportDatabaseMetadataOnUserDelete {
		err = DeleteExportDatabasesUser(token, conf)
		if err != nil {
			conf.GetLogger().Error("unable to delete user from export-databases-service", "userId", userId, "error", err)
			return err
		}
	}

	err = DeleteAnalyticsOperatorRepoUser(token, conf)
	if err != nil {
		conf.GetLogger().Error("unable to delete user from analytics-operator-repo", "userId", userId, "error", err)
		return err
	}
	err = DeleteAnalyticsFlowRepoUser(token, conf)
	if err != nil {
		conf.GetLogger().Error("unable to delete user from analytics-flow-repo", "userId", userId, "error", err)
		return err
	}
	err = DeleteAnalyticsFlowEngineUser(token, conf)
	if err != nil {
		conf.GetLogger().Error("unable to delete user from analytics-flow-engine", "userId", userId, "error", err)
		return err
	}
	err = DeleteNotificationUser(token, conf)
	if err != nil {
		conf.GetLogger().Error("unable to delete user from notification-service", "userId", userId, "error", err)
		return err
	}
	err = DeleteKeycloakUser(userId, conf)
	if err != nil {
		conf.GetLogger().Error("unable to delete user from keycloak", "userId", userId, "error", err)
		return err
	}
	return nil
}

type IdWrapper struct {
	Id string `json:"id"`
}

type DataWrapper[T any] struct {
	Data T `json:"data"`
}

type ExportListIdWrapper struct {
	Instances []ExportIdWrapper `json:"instances"`
}

type ExportIdWrapper struct {
	Id string `json:"ID"`
}

type UnderscoreIdWrapper struct {
	Id string `json:"_id"`
}
