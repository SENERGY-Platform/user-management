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

func initFlowEngineState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token, ids *[]string) func(t *testing.T) {
	return func(t *testing.T) {
		temp := ctrl.IdWrapper{}
		err := user1.Impersonate().PostJSON(
			config.AnalyticsFlowEngineUrl+"/pipeline",
			map[string]interface{}{
				"name":    "1",
				"flowId":  "1",
				"metrics": false,
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*ids = append(*ids, temp.Id)

		temp = ctrl.IdWrapper{}
		err = user1.Impersonate().PostJSON(
			config.AnalyticsFlowEngineUrl+"/pipeline",
			map[string]interface{}{
				"name":    "2",
				"flowId":  "2",
				"metrics": false,
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*ids = append(*ids, temp.Id)

		temp = ctrl.IdWrapper{}
		err = user2.Impersonate().PostJSON(
			config.AnalyticsFlowEngineUrl+"/pipeline",
			map[string]interface{}{
				"name":    "3",
				"flowId":  "3",
				"metrics": false,
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*ids = append(*ids, temp.Id)
	}
}

func checkFlowEngineState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token, ids []string) func(t *testing.T) {
	return func(t *testing.T) {
		if len(ids) != 3 {
			t.Error(ids)
			return
		}
		temp := []ctrl.IdWrapper{}
		err := user1.Impersonate().GetJSON(config.AnalyticsPipelineUrl+"/pipeline", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp) != 0 {
			t.Error(temp)
			return
		}

		temp = []ctrl.IdWrapper{}
		err = user2.Impersonate().GetJSON(config.AnalyticsPipelineUrl+"/pipeline", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp) != 1 {
			t.Error(temp)
			return
		}
		if temp[0].Id != ids[2] {
			t.Error(temp)
		}
	}
}
