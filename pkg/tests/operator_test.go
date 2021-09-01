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

func initOperatorState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token, ids *[]string) func(t *testing.T) {
	return func(t *testing.T) {
		temp := ctrl.UnderscoreIdWrapper{}
		err := user1.Impersonate().PutJSON(
			config.AnalyticsOperatorRepoUrl+"/operator",
			map[string]interface{}{
				"name":  "1",
				"image": "ghcr.io/senergy-platform/hello-world:test",
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*ids = append(*ids, temp.Id)

		temp = ctrl.UnderscoreIdWrapper{}
		err = user1.Impersonate().PutJSON(
			config.AnalyticsOperatorRepoUrl+"/operator",
			map[string]interface{}{
				"name":  "2",
				"image": "ghcr.io/senergy-platform/hello-world:test",
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*ids = append(*ids, temp.Id)

		temp = ctrl.UnderscoreIdWrapper{}
		err = user2.Impersonate().PutJSON(
			config.AnalyticsOperatorRepoUrl+"/operator",
			map[string]interface{}{
				"name":  "3",
				"image": "ghcr.io/senergy-platform/hello-world:test",
				"pub":   true,
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*ids = append(*ids, temp.Id)
		temp = ctrl.UnderscoreIdWrapper{}
		err = user2.Impersonate().PutJSON(
			config.AnalyticsOperatorRepoUrl+"/operator",
			map[string]interface{}{
				"name":  "4",
				"image": "ghcr.io/senergy-platform/hello-world:test",
				"pub":   false,
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*ids = append(*ids, temp.Id)
	}
}

func checkOperatorState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token, ids []string) func(t *testing.T) {
	return func(t *testing.T) {
		if len(ids) != 4 {
			t.Error(ids)
			return
		}
		temp := ctrl.OperatorList{}
		err := user1.Impersonate().GetJSON(config.AnalyticsOperatorRepoUrl+"/operator", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		//one public operator from user2
		if len(temp.Operators) != 1 {
			t.Error(temp)
			return
		}
		if temp.Operators[0].Id != ids[2] {
			t.Error(temp)
		}

		temp = ctrl.OperatorList{}
		err = user2.Impersonate().GetJSON(config.AnalyticsOperatorRepoUrl+"/operator", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp.Operators) != 2 {
			t.Error(temp)
			return
		}
		if temp.Operators[0].Id != ids[2] {
			t.Error(temp)
		}
		if temp.Operators[1].Id != ids[3] {
			t.Error(temp)
		}
	}
}
