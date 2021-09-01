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
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"log"
)

func DeleteUser(userId string, conf configuration.Config) (err error) {
	token, err := CreateToken("users-service", userId)
	if err != nil {
		log.Println("ERROR: unable to create jwt for userId", userId, err)
		return err
	}
	err = DeleteWaitingRoomUser(token, conf)
	if err != nil {
		log.Println("ERROR: DeleteWaitingRoomUser()", err)
		return err
	}
	err = DeleteDashboardUser(token, conf)
	if err != nil {
		log.Println("ERROR: DeleteDashboardUser()", err)
		return err
	}
	err = DeleteProcessSchedulerUser(token, conf)
	if err != nil {
		log.Println("ERROR: DeleteProcessSchedulerUser()", err)
		return err
	}
	err = DeleteImportsUser(token, conf)
	if err != nil {
		log.Println("ERROR: DeleteImportsUser()", err)
		return err
	}
	err = DeleteBrokerExportsUser(token, conf)
	if err != nil {
		log.Println("ERROR: DeleteBrokerExportsUser()", err)
		return err
	}
	err = DeleteDatabaseExportsUser(token, conf)
	if err != nil {
		log.Println("ERROR: DeleteDatabaseExportsUser()", err)
		return err
	}
	err = DeleteAnalyticsOperatorRepoUser(token, conf)
	if err != nil {
		log.Println("ERROR: DeleteAnalyticsOperatorRepoUser()", err)
		return err
	}
	err = DeleteKeycloakUser(userId, conf)
	if err != nil {
		log.Println("ERROR: DeleteKeycloakUser()", err)
		return err
	}
	return nil
}

type IdWrapper struct {
	Id string `json:"id"`
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
