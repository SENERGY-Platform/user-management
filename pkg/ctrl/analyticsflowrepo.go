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
	"github.com/SENERGY-Platform/analytics-flow-repo-v2/client"
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
)

func DeleteAnalyticsFlowRepoUser(token Token, conf configuration.Config) error {
	if conf.AnalyticsFlowRepoUrl == "" || conf.AnalyticsFlowRepoUrl == "-" {
		return nil
	}
	ids, err := getAnalyticsFlowIds(token, conf)
	if err != nil {
		return err
	}
	for _, id := range ids {
		err = deleteAnalyticsFlow(token, conf, id)
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteAnalyticsFlow(token Token, conf configuration.Config, id string) error {
	_, err := client.NewClient(conf.AnalyticsFlowRepoUrl).DeleteFlow(token.Token, token.GetUserId(), id)
	return err
}

func getAnalyticsFlowIds(token Token, config configuration.Config) (ids []string, err error) {
	ids = []string{}

	flows, _, err := client.NewClient(config.AnalyticsFlowRepoUrl).GetFlows(token.Token, token.GetUserId())
	if err != nil {
		return ids, err
	}

	for _, element := range flows.Flows {
		if element.UserId == token.GetUserId() { //filter public flows of other users
			ids = append(ids, element.Id.Hex())
		}
	}
	return ids, err
}
