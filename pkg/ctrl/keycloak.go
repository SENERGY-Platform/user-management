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

package ctrl

import (
	"fmt"
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"log"
	"net/http"
	"net/url"
)

type User struct {
	Id   string `json:"id"`
	Name string `json:"username"`
	//Enabled    bool                   `json:"enabled"`
	//FirstName  string                 `json:"firstName"`
	//LastName   string                 `json:"lastName"`
	Attributes map[string]interface{} `json:"attributes"`
}

type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

//func GetUserByName(name string, conf configuration.Config) (user User, err error) {
//	token, err := EnsureAccess(conf)
//	if err != nil {
//		return user, err
//	}
//	temp := []User{}
//	err = token.GetJSON(conf.KeycloakUrl+"/auth/admin/realms/"+conf.KeycloakRealm+"/users?username="+url.QueryEscape(name), &temp)
//	if err != nil {
//		return user, err
//	}
//	temp = filterExact(temp, name)
//	size := len(temp)
//	switch {
//	case size > 1:
//		err = errors.New("ambiguous username")
//	case size < 1:
//		err = errors.New("unable to find user")
//	default:
//		user = temp[0]
//	}
//	return
//}

func filterExact(users []User, name string) (result []User) {
	for _, user := range users {
		if user.Name == name {
			result = append(result, user)
		}
	}
	return
}

func GetUserById(id string, conf configuration.Config) (user User, err error) {
	token, err := EnsureAccess(conf)
	if err != nil {
		return user, err
	}
	err = token.GetJSON(conf.KeycloakUrl+"/auth/admin/realms/"+conf.KeycloakRealm+"/users/"+url.QueryEscape(id), &user)
	return
}

func DeleteKeycloakUser(id string, conf configuration.Config) (err error) {
	token, err := EnsureAccess(conf)
	if err != nil {
		log.Println("ERROR: unable to ensure access", err)
		return err
	}
	resp, err := token.Delete(conf.KeycloakUrl+"/auth/admin/realms/"+conf.KeycloakRealm+"/users/"+url.QueryEscape(id), nil)
	if err != nil || (resp != nil && resp.StatusCode == http.StatusNotFound) {
		log.Println("WARNING: user dosnt exist; command will be ignored")
		err = nil
	}
	return err
}

func GetUsers(excludeID string, conf configuration.Config) ([]User, error) {
	return getUsers(conf.KeycloakUrl+"/auth/admin/realms/"+conf.KeycloakRealm+"/users", excludeID, conf)
}

func GetUsersGroups(id string, conf configuration.Config) ([]Group, error) {
	var groups []Group
	pageNum := 0
	for {
		token, err := EnsureAccess(conf)
		if err != nil {
			return nil, err
		}
		var page []Group
		if err = token.GetJSON(conf.KeycloakUrl+"/auth/admin/realms/"+conf.KeycloakRealm+"/users/"+id+"/groups"+fmt.Sprintf("?max=%d&first=%d", conf.KeycloakPageMax, conf.KeycloakPageMax*pageNum), &page); err != nil {
			return nil, err
		}
		if len(page) == 0 {
			break
		}
		groups = append(groups, page...)
		pageNum++
	}
	return groups, nil
}

func GetGroupMembersCombined(groups []Group, excludeID string, conf configuration.Config) ([]User, error) {
	var users []User
	userSet := make(map[string]struct{})
	for _, group := range groups {
		members, err := getUsers(conf.KeycloakUrl+"/auth/admin/realms/"+conf.KeycloakRealm+"/groups/"+url.QueryEscape(group.ID)+"/members", excludeID, conf)
		if err != nil {
			return nil, err
		}
		for _, user := range members {
			if _, ok := userSet[user.Id]; ok {
				continue
			}
			userSet[user.Id] = struct{}{}
			users = append(users, user)
		}
	}
	return users, nil
}

func getUsers(url string, excludeID string, conf configuration.Config) ([]User, error) {
	var users []User
	pageNum := 0
	for {
		token, err := EnsureAccess(conf)
		if err != nil {
			return nil, err
		}
		var page []User
		if err = token.GetJSON(url+fmt.Sprintf("?max=%d&first=%d", conf.KeycloakPageMax, conf.KeycloakPageMax*pageNum), &page); err != nil {
			return nil, err
		}
		if len(page) == 0 {
			break
		}
		if excludeID != "" {
			for _, user := range page {
				if user.Id == excludeID {
					continue
				}
				users = append(users, user)
			}
		} else {
			users = append(users, page...)
		}
		pageNum++
	}
	return users, nil
}
