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
	"context"
	"encoding/json"
	"github.com/SmartEnergyPlatform/jwt-http-router"
	"github.com/SmartEnergyPlatform/util/http/cors"
	"github.com/SmartEnergyPlatform/util/http/logger"
	"github.com/SmartEnergyPlatform/util/http/response"
	"io"
	"log"
	"net/http"
	"sync"
)

type api struct {
	eventHandler *EventHandler
	conf         Config
}

func StartApi(ctx context.Context, conf Config) (wg *sync.WaitGroup, err error) {
	wg = &sync.WaitGroup{}

	eventHandler, err := InitEventConn(ctx, wg, conf)
	if err != nil {
		return
	}
	apiInstance := &api{
		eventHandler: eventHandler,
		conf:         conf,
	}
	log.Println("start server on port: ", conf.ServerPort)
	httpHandler := apiInstance.getRoutes()
	corseHandler := cors.New(httpHandler)
	logg := logger.New(corseHandler, conf.LogLevel)
	go func() { log.Println(http.ListenAndServe(":"+conf.ServerPort, logg)) }()
	return
}

func (api *api) getRoutes() (router *jwt_http_router.Router) {
	router = jwt_http_router.New(jwt_http_router.JwtConfig{
		ForceUser: api.conf.ForceUser == "true",
		ForceAuth: api.conf.ForceAuth == "true",
		PubRsa:    api.conf.JwtPubRsa,
	})

	router.GET("/user/id/:id", func(res http.ResponseWriter, r *http.Request, ps jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		id := ps.ByName("id")
		user, err := GetUserById(id, api.conf)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(res).Json(user)
	})

	router.DELETE("/user/id/:id", func(res http.ResponseWriter, r *http.Request, ps jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		id := ps.ByName("id")
		if jwt.UserId != id && !isAdmin(jwt) {
			log.Println("DEBUG: ", jwt.RealmAccess.Roles, jwt.ResourceAccess)
			http.Error(res, "access denied", http.StatusUnauthorized)
			return
		}
		err := api.eventHandler.DeleteUser(id)
		if err != nil {
			http.Error(res, err.Error(), http.StatusPreconditionFailed)
			return
		}
		response.To(res).Json("ok")
	})

	router.DELETE("/user", func(res http.ResponseWriter, r *http.Request, ps jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		err := api.eventHandler.DeleteUser(jwt.UserId)
		if err != nil {
			http.Error(res, err.Error(), http.StatusPreconditionFailed)
			return
		}
		response.To(res).Json("ok")
	})

	router.GET("/user/id/:id/name", func(res http.ResponseWriter, r *http.Request, ps jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		id := ps.ByName("id")
		user, err := GetUserById(id, api.conf)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(res).Json(user.Name)
	})

	router.GET("/user/name/:name", func(res http.ResponseWriter, r *http.Request, ps jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		name := ps.ByName("name")
		user, err := GetUserByName(name, api.conf)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(res).Json(user)
	})

	router.GET("/user/name/:name/id", func(res http.ResponseWriter, r *http.Request, ps jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		name := ps.ByName("name")
		user, err := GetUserByName(name, api.conf)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(res).Json(user.Id)
	})

	router.GET("/sessions", func(res http.ResponseWriter, r *http.Request, ps jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		token, err := EnsureAccess(api.conf)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		var result interface{}
		err = token.GetJSON(api.conf.KeycloakUrl+"/auth/admin/realms/"+api.conf.KeycloakRealm+"/users/"+jwt.UserId+"/sessions", &result)
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

	router.PUT("/password", func(res http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		passwordRequest := PasswordRequest{}
		err := json.NewDecoder(request.Body).Decode(&passwordRequest)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		token, err := EnsureAccess(api.conf)
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
		resp, err := token.Put(api.conf.KeycloakUrl+"/auth/admin/realms/"+api.conf.KeycloakRealm+"/users/"+jwt.UserId+"/reset-password", "application/json", r)
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

	router.PUT("/info", func(res http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		infoRequest := UserInfoRequest{}
		err := json.NewDecoder(request.Body).Decode(&infoRequest)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		token, err := EnsureAccess(api.conf)
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
		resp, err := token.Put(api.conf.KeycloakUrl+"/auth/admin/realms/"+api.conf.KeycloakRealm+"/users/"+jwt.UserId, "application/json", r)
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

	router.GET("/roles", func(res http.ResponseWriter, r *http.Request, ps jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		roles, err := GetRoles(api.conf)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		response.To(res).Json(roles)
	})

	return
}

func isAdmin(jwt jwt_http_router.Jwt) bool {
	for _, role := range jwt.RealmAccess.Roles {
		if role == "admin" {
			return true
		}
	}
	return false
}

type PasswordRequest struct {
	Password string `json:"password"`
}

type UserInfoRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}
