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
	"github.com/swaggo/http-swagger"
	"github.com/swaggo/swag"
	"io"
	"log"
	"net/http"
	"strings"
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
	logg := util.NewLogger(corsHandler)
	go func() { log.Println(http.ListenAndServe(":"+conf.ServerPort, logg)) }()
	return
}

// @title         User Management API
// @version       v0.0.5
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath  /
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func (api *api) getRoutes() (router *httprouter.Router) {
	router = httprouter.New()
	api.getUserByID(router)
	api.deleteUserByID(router)
	api.deleteUser(router)
	api.getUsernameByID(router)
	api.getUsers(router)
	api.getSessions(router)
	api.putPassword(router)
	api.putInfo(router)
	if api.conf.EnableSwaggerUi {
		router.GET("/swagger/:any", func(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
			httpSwagger.WrapHandler(res, req)
		})
	}
	router.GET("/doc", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		doc, err := swag.ReadDoc()
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		doc = strings.Replace(doc, `"host": "",`, "", 1)
		_, _ = writer.Write([]byte(doc))
	})
	return
}

// getUserByID godoc
// @Summary      get user by ID
// @Description  get user by providing a user ID
// @Tags         user
// @Security Bearer
// @Param        id path string true "user ID"
// @Produce      json
// @Success      200 {object} ctrl.User
// @Failure      500
// @Router       /user/id/{id} [get]
func (api *api) getUserByID(router *httprouter.Router) {
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
}

// deleteUserByID godoc
// @Summary      delete user by ID
// @Description  delete user by providing a user ID
// @Tags         user
// @Security Bearer
// @Param        id path string true "user ID"
// @Produce      json
// @Success      200 {object} string
// @Failure      400
// @Failure      403
// @Failure      412
// @Failure      500
// @Router       /user/id/{id} [delete]
func (api *api) deleteUserByID(router *httprouter.Router) {
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
}

// deleteUser godoc
// @Summary      delete user
// @Description  delete user by parsing provided jwt token
// @Tags         user
// @Security Bearer
// @Produce      json
// @Success      200 {object} string
// @Failure      401
// @Failure      412
// @Failure      500
// @Router       /user [delete]
func (api *api) deleteUser(router *httprouter.Router) {
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
}

// getUsernameByID godoc
// @Summary      get username
// @Description  get username by providing a user ID
// @Tags         user
// @Security Bearer
// @Param        id path string true "user ID"
// @Produce      json
// @Success      200 {object} string
// @Failure      500
// @Router       /user/id/{id}/name [get]
func (api *api) getUsernameByID(router *httprouter.Router) {
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
}

// getUsers godoc
// @Summary      get users
// @Description  parses provided jwt and lists all users if admin or only lists users from groups the calling user is a member of
// @Tags         user
// @Security Bearer
// @Param        excludeCaller query bool false "if true exclude calling user from result"
// @Produce      json
// @Success      200 {array} ctrl.User
// @Failure      400
// @Failure      500
// @Router       /user-list [get]
func (api *api) getUsers(router *httprouter.Router) {
	router.GET("/user-list", func(res http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		token, err := GetParsedToken(r)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		excludeID := ""
		if excludeCaller := r.URL.Query().Get("excludeCaller"); excludeCaller == "true" {
			excludeID = token.GetUserId()
		}
		var users []ctrl.User
		if token.IsAdmin() {
			users, err = ctrl.GetUsers(excludeID, api.conf)
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			groups, err := ctrl.GetUsersGroups(token.GetUserId(), api.conf)
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			users, err = ctrl.GetGroupMembersCombined(groups, excludeID, api.conf)
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(res).Encode(users)
	})
}

// getSessions godoc
// @Summary      get user's sessions
// @Description  get user's sessions by parsing provided jwt token
// @Tags         user
// @Security Bearer
// @Produce      json
// @Success      200 {array} object
// @Failure      400
// @Failure      500
// @Router       /sessions [get]
func (api *api) getSessions(router *httprouter.Router) {
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
}

// putPassword godoc
// @Summary      set password for user
// @Description  set password for user, parses jwt for ID
// @Tags         user
// @Security Bearer
// @Accept       json
// @Produce      json
// @Param        message body PasswordRequest true "user password"
// @Success      200 {object} object
// @Failure      400
// @Failure      500
// @Router       /password [put]
func (api *api) putPassword(router *httprouter.Router) {
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
}

// putInfo godoc
// @Summary      set user info
// @Description  set user's details, parses jwt for ID
// @Tags         user
// @Security Bearer
// @Accept       json
// @Produce      json
// @Param        message body UserInfoRequest true "user details"
// @Success      200 {object} object
// @Failure      400
// @Failure      500
// @Router       /info [put]
func (api *api) putInfo(router *httprouter.Router) {
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
}

type PasswordRequest struct {
	Password string `json:"password"`
}

type UserInfoRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}
