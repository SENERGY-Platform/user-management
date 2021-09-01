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
	"errors"
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"io"
)

func DeleteAnalyticsOperatorRepoUser(token Token, conf configuration.Config) error {
	if conf.AnalyticsOperatorRepoUrl == "" || conf.AnalyticsOperatorRepoUrl == "-" {
		return nil
	}
	ids, err := getAnalyticsOperatorIds(token, conf)
	if err != nil {
		return err
	}
	err = deleteAnalyticsOperator(token, conf, ids)
	if err != nil {
		return err
	}
	return nil
}

func deleteAnalyticsOperator(token Token, conf configuration.Config, ids []string) error {
	if len(ids) > 0 {
		resp, err := token.Impersonate().DeleteWithBody(conf.AnalyticsOperatorRepoUrl+"/operator", ids)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			temp, _ := io.ReadAll(resp.Body)
			return errors.New("deleteAnalyticsOperator(): " + string(temp))
		}
	}
	return nil
}

func getAnalyticsOperatorIds(token Token, config configuration.Config) (ids []string, err error) {
	temp := OperatorList{}
	//limit=0 -> mongodb: all elements
	err = token.Impersonate().GetJSON(config.AnalyticsOperatorRepoUrl+"/operator?limit=0&offset=0", &temp)
	if err != nil {
		return ids, err
	}
	for _, element := range temp.Operators {
		if element.UserId == token.GetUserId() { //filter public operators of other users
			ids = append(ids, element.Id)
		}
	}
	return ids, err
}

type Operator struct {
	UnderscoreIdWrapper
	Public bool   `json:"pub"`
	UserId string `json:"userId"`
}

type OperatorList struct {
	Operators []Operator `json:"operators"`
}
