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

package api

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/user-management/pkg/api/util"
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"github.com/SENERGY-Platform/user-management/pkg/ctrl"
	"github.com/julienschmidt/httprouter"
	"io"
	"log"
	"net/http"
	"sync"
)

type api struct {
	eventHandler *ctrl.EventHandler
	conf         configuration.Config
}

func Start(ctx context.Context, conf configuration.Config) (wg *sync.WaitGroup, err error) {
	wg = &sync.WaitGroup{}

	eventHandler, err := ctrl.InitEventConn(ctx, wg, conf)
	if err != nil {
		return
	}
	apiInstance := &api{
		eventHandler: eventHandler,
		conf:         conf,
	}
	log.Println("start server on port: ", conf.ServerPort)
	httpHandler := apiInstance.getRoutes()
	corsHandler := util.NewCors(httpHandler)
	logg := util.NewLogger(corsHandler, conf.LogLevel)
	go func() { log.Println(http.ListenAndServe(":"+conf.ServerPort, logg)) }()
	return
}

func (api *api) getRoutes() (router *httprouter.Router) {
	router = httprouter.New()

	router.GET("/user/id/:id", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		id := ps.ByName("id")
		user, err := ctrl.GetUserById(id, api.conf)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(res).Encode(user)
	})

	router.DELETE("/user/id/:id", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		id := ps.ByName("id")
		token, err := GetParsedToken(r)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		if token.GetUserId() != id && !token.IsAdmin() {
			http.Error(res, "access denied", http.StatusForbidden)
			return
		}
		err = api.eventHandler.DeleteUser(id)
		if err != nil {
			http.Error(res, err.Error(), http.StatusPreconditionFailed)
			return
		}
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(res).Encode("ok")
	})

	router.DELETE("/user", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		token, err := GetParsedToken(r)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		err = api.eventHandler.DeleteUser(token.GetUserId())
		if err != nil {
			http.Error(res, err.Error(), http.StatusPreconditionFailed)
			return
		}
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(res).Encode("ok")
	})

	router.GET("/user/id/:id/name", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		id := ps.ByName("id")
		user, err := ctrl.GetUserById(id, api.conf)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(res).Encode(user.Name)
	})

	router.GET("/user/name/:name", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		name := ps.ByName("name")
		user, err := ctrl.GetUserByName(name, api.conf)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(res).Encode(user)
	})

	router.GET("/user/name/:name/id", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		name := ps.ByName("name")
		user, err := ctrl.GetUserByName(name, api.conf)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(res).Encode(user.Id)
	})

	router.GET("/sessions", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		usertoken, err := GetParsedToken(r)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		token, err := ctrl.EnsureAccess(api.conf)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		var result interface{}
		err = token.GetJSON(api.conf.KeycloakUrl+"/auth/admin/realms/"+api.conf.KeycloakRealm+"/users/"+usertoken.GetUserId()+"/sessions", &result)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		} else {
			res.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(res).Encode(result)
			if err != nil {
				log.Println("ERROR: unable to respond", err)
			}
		}
	})

	router.PUT("/password", func(res http.ResponseWriter, request *http.Request, params httprouter.Params) {
		usertoken, err := GetParsedToken(request)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		passwordRequest := PasswordRequest{}
		err = json.NewDecoder(request.Body).Decode(&passwordRequest)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		token, err := ctrl.EnsureAccess(api.conf)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		r, w := io.Pipe()
		go func() {
			defer w.Close()
			err = json.NewEncoder(w).Encode(map[string]interface{}{
				"type":      "password",
				"value":     passwordRequest.Password,
				"temporary": false,
			})
			if err != nil {
				log.Println("ERROR:", err)
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
		}()
		resp, err := token.Put(api.conf.KeycloakUrl+"/auth/admin/realms/"+api.conf.KeycloakRealm+"/users/"+usertoken.GetUserId()+"/reset-password", "application/json", r)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.WriteHeader(resp.StatusCode)
		_, err = io.Copy(res, resp.Body)
		if err != nil {
			log.Println("ERROR: /password io.Copy ", err)
		}
	})

	router.PUT("/info", func(res http.ResponseWriter, request *http.Request, params httprouter.Params) {
		usertoken, err := GetParsedToken(request)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		infoRequest := UserInfoRequest{}
		err = json.NewDecoder(request.Body).Decode(&infoRequest)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		token, err := ctrl.EnsureAccess(api.conf)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		r, w := io.Pipe()
		go func() {
			defer w.Close()
			err = json.NewEncoder(w).Encode(map[string]interface{}{
				"firstName": infoRequest.FirstName,
				"lastName":  infoRequest.LastName,
				"email":     infoRequest.Email,
			})
			if err != nil {
				log.Println("ERROR:", err)
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
		}()
		resp, err := token.Put(api.conf.KeycloakUrl+"/auth/admin/realms/"+api.conf.KeycloakRealm+"/users/"+usertoken.GetUserId(), "application/json", r)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.WriteHeader(resp.StatusCode)
		_, err = io.Copy(res, resp.Body)
		if err != nil {
			log.Println("ERROR: /info io.Copy ", err)
		}
	})

	router.GET("/roles", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		roles, err := ctrl.GetRoles(api.conf)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(res).Encode(roles)
	})

	return
}

type PasswordRequest struct {
	Password string `json:"password"`
}

type UserInfoRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}
