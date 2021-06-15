/*
 * Copyright 2018 InfAI (CC SES)
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

package main

import (
	"errors"
	"log"
	"net/http"
	"net/url"
)

type User struct {
	Id         string                 `json:"id"`
	Name       string                 `json:"username"`
	Attributes map[string]interface{} `json:"attributes"`
}

func GetUserByName(name string, conf Config) (user User, err error) {
	token, err := EnsureAccess(conf)
	if err != nil {
		return user, err
	}
	temp := []User{}
	err = token.GetJSON(conf.KeycloakUrl+"/auth/admin/realms/"+conf.KeycloakRealm+"/users?username="+url.QueryEscape(name), &temp)
	if err != nil {
		return user, err
	}
	temp = filterExact(temp, name)
	size := len(temp)
	switch {
	case size > 1:
		err = errors.New("ambiguous username")
	case size < 1:
		err = errors.New("unable to find user")
	default:
		user = temp[0]
	}
	return
}

func filterExact(users []User, name string) (result []User) {
	for _, user := range users {
		if user.Name == name {
			result = append(result, user)
		}
	}
	return
}

func GetUserById(id string, conf Config) (user User, err error) {
	token, err := EnsureAccess(conf)
	if err != nil {
		return user, err
	}
	err = token.GetJSON(conf.KeycloakUrl+"/auth/admin/realms/"+conf.KeycloakRealm+"/users/"+url.QueryEscape(id), &user)
	return
}

func DeleteKeycloakUser(id string, conf Config) (err error) {
	token, err := EnsureAccess(conf)
	if err != nil {
		log.Println("ERROR: unable to ensure access", err)
		return err
	}
	resp, err := token.Delete(conf.KeycloakUrl + "/auth/admin/realms/" + conf.KeycloakRealm + "/users/" + url.QueryEscape(id))
	if err != nil || (resp != nil && resp.StatusCode == http.StatusNotFound) {
		log.Println("WARNING: user dosnt exist; command will be ignored")
		err = nil
	}
	return err
}
