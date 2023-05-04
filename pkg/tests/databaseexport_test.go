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

package tests

import (
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"github.com/SENERGY-Platform/user-management/pkg/ctrl"
	"testing"
)

func getDbExportElement(name string, dbId string) map[string]interface{} {
	return map[string]interface{}{
		"Name":             name,
		"FilterType":       "deviceId",
		"Filter":           name,
		"EntityName":       name,
		"ServiceName":      name,
		"Topic":            name,
		"Offset":           "foo",
		"ExportDatabaseID": dbId,
	}
}

type ExportDatabaseRequest struct {
	Name          string `json:"Name" validate:"required"`
	Description   string `json:"Description"`
	Type          string `json:"Type" validate:"required"`
	Deployment    string `json:"deployment" validate:"required"`
	Url           string `json:"Url" validate:"required"`
	EwFilterTopic string `json:"EwFilterTopic" validate:"required"`
	Public        bool   `json:"Public"`
}

func initPublicExportDatabase(config configuration.Config, user ctrl.Token, dbId *string) func(t *testing.T) {
	return func(t *testing.T) {
		temp := ctrl.ExportIdWrapper{}
		err := user.Impersonate().PostJSON(
			config.DatabaseExportsUrl+"/databases",
			ExportDatabaseRequest{
				Name:          "edb1",
				Description:   "",
				Type:          "influxdb",
				Deployment:    "foo",
				Url:           "bar",
				EwFilterTopic: "batz",
				Public:        true,
			},
			&temp)
		if err != nil {
			t.Error(err)
			t.Logf("%#v", err)
			return
		}
		*dbId = temp.Id
	}
}

func initDatabaseExportState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token, dbId string, ids *[]string) func(t *testing.T) {
	return func(t *testing.T) {
		temp := ctrl.ExportIdWrapper{}
		err := user1.Impersonate().PostJSON(
			config.DatabaseExportsUrl+"/instance",
			getDbExportElement("1", dbId),
			&temp)
		if err != nil {
			t.Error(err)
			t.Logf("%#v", err)
			return
		}
		*ids = append(*ids, temp.Id)

		temp = ctrl.ExportIdWrapper{}
		err = user1.Impersonate().PostJSON(
			config.DatabaseExportsUrl+"/instance",
			getDbExportElement("2", dbId),
			&temp)
		if err != nil {
			t.Error(err)
			return
		}
		*ids = append(*ids, temp.Id)

		temp = ctrl.ExportIdWrapper{}
		err = user2.Impersonate().PostJSON(
			config.DatabaseExportsUrl+"/instance",
			getDbExportElement("3", dbId),
			&temp)
		if err != nil {
			t.Error(err)
			return
		}
		*ids = append(*ids, temp.Id)
	}
}

func checkDatabaseExportsState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token, ids []string) func(t *testing.T) {
	return func(t *testing.T) {
		if len(ids) != 3 {
			t.Error(ids)
		}
		temp := ctrl.ExportListIdWrapper{}
		err := user1.Impersonate().GetJSON(config.DatabaseExportsUrl+"/instance", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp.Instances) != 0 {
			t.Error(temp)
		}

		temp = ctrl.ExportListIdWrapper{}
		err = user2.Impersonate().GetJSON(config.DatabaseExportsUrl+"/instance", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp.Instances) != 1 {
			t.Error(temp)
			return
		}
		if temp.Instances[0].Id != ids[2] {
			t.Error(temp)
		}
	}
}
