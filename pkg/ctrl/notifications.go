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

func DeleteNotificationUser(token Token, conf configuration.Config) error {
	if conf.NotifierUrl == "" || conf.NotifierUrl == "-" {
		return nil
	}
	ids, err := getNotificationIds(token, conf)
	if err != nil {
		return err
	}
	err = deleteNotifications(token, conf, ids)
	if err != nil {
		return err
	}

	ids, err = getBrokerIds(token, conf)
	if err != nil {
		return err
	}
	err = deleteBrokers(token, conf, ids)
	if err != nil {
		return err
	}

	err = deletePlatformBrokerConfig(token, conf)
	if err != nil {
		return err
	}

	return nil
}

func deleteNotifications(token Token, conf configuration.Config, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	resp, err := token.Impersonate().Delete(conf.NotifierUrl+"/notifications", ids)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		temp, _ := io.ReadAll(resp.Body)
		return errors.New("deleteNotifications(): " + string(temp))
	}
	return nil
}

func getNotificationIds(token Token, config configuration.Config) (ids []string, err error) {
	temp := NotificationList{}
	//limit=0 -> mongodb: all elements
	err = token.Impersonate().GetJSON(config.NotifierUrl+"/notifications?limit=0&offset=0", &temp)
	if err != nil {
		return ids, err
	}
	for _, element := range temp.Notifications {
		ids = append(ids, element.Id)
	}
	return ids, err
}

func getBrokerIds(token Token, config configuration.Config) (ids []string, err error) {
	temp := BrokerList{}
	//limit=0 -> mongodb: all elements
	err = token.Impersonate().GetJSON(config.NotifierUrl+"/brokers?limit=0&offset=0", &temp)
	if err != nil {
		return ids, err
	}
	for _, element := range temp.Brokers {
		ids = append(ids, element.Id)
	}
	return ids, err
}

func deleteBrokers(token Token, conf configuration.Config, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	resp, err := token.Impersonate().Delete(conf.NotifierUrl+"/brokers", ids)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		temp, _ := io.ReadAll(resp.Body)
		return errors.New("deleteBrokers(): " + string(temp))
	}
	return nil
}

func deletePlatformBrokerConfig(token Token, conf configuration.Config) error {
	resp, err := token.Impersonate().Delete(conf.NotifierUrl+"/platform-broker", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		temp, _ := io.ReadAll(resp.Body)
		return errors.New("deletePlatformBrokerConfig(): " + string(temp))
	}
	return nil
}

type NotificationList struct {
	Notifications []UnderscoreIdWrapper `json:"notifications"`
}

type BrokerList struct {
	Brokers []IdWrapper `json:"brokers"`
}
