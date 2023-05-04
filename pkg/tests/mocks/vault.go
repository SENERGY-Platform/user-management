package mocks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"runtime/debug"
	"sync"
)

func MockVault(ctx context.Context) (addr string, err error) {
	router, err := getVaultRouter()
	if err != nil {
		return "", err
	}
	server := &httptest.Server{
		Config: &http.Server{Handler: router},
	}

	server.Listener, _ = net.Listen("tcp", ":")
	server.Start()

	go func() {
		<-ctx.Done()
		server.Close()
	}()
	return server.URL, nil
}

func getVaultRouter() (router *httprouter.Router, err error) {
	defer func() {
		if r := recover(); r != nil && err == nil {
			log.Printf("%s: %s", r, debug.Stack())
			err = errors.New(fmt.Sprint("Recovered Error: ", r))
		}
	}()
	router = httprouter.New()

	token := map[string]interface{}{
		"request_id":     "000",
		"lease_id":       "123",
		"lease_duration": 0,
		"renewable":      true,
		"auth": map[string]interface{}{
			"client_token":   "token",
			"accessor":       "accessor",
			"entity_id":      "123",
			"renewable":      true,
			"orphan":         true,
			"lease_duration": 3600,
		},
	}
	data := make(map[string]map[string]interface{})
	mux := sync.Mutex{}

	router.HandlerFunc(http.MethodPost, "/v1/:engine/jwt/login", func(writer http.ResponseWriter, request *http.Request) {
		vars, ok := request.Context().Value(httprouter.ParamsKey).(httprouter.Params)
		if !ok {
			debug.PrintStack()
			log.Fatal("nope")
		}
		if vars.ByName("engine") != "auth" {
			http.Error(writer, "unexpected cal to "+request.RequestURI, 404)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		token["request_id"] = uuid.NewString()
		token["lease_id"] = uuid.NewString()
		err = json.NewEncoder(writer).Encode(token)
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
	})

	router.HandlerFunc(http.MethodPut, "/v1/:engine/token/renew-self", func(writer http.ResponseWriter, request *http.Request) {
		vars, ok := request.Context().Value(httprouter.ParamsKey).(httprouter.Params)
		if !ok {
			debug.PrintStack()
			log.Fatal("nope")
		}
		if vars.ByName("engine") != "auth" {
			http.Error(writer, "unexpected cal to "+request.RequestURI, 404)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		requestData := map[string]interface{}{}
		_ = json.NewDecoder(request.Body).Decode(&requestData)
		token["request_id"] = uuid.NewString()
		token["lease_id"] = uuid.NewString()
		err = json.NewEncoder(writer).Encode(token)
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
	})

	router.HandlerFunc(http.MethodGet, "/v1/:engine/data/:key", func(writer http.ResponseWriter, request *http.Request) {
		mux.Lock()
		defer mux.Unlock()
		vars, ok := request.Context().Value(httprouter.ParamsKey).(httprouter.Params)
		if !ok {
			debug.PrintStack()
			log.Fatal("nope")
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"data": data[vars.ByName("engine")][vars.ByName("key")],
			},
		})
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
	})

	router.HandlerFunc(http.MethodDelete, "/v1/:engine/metadata/:key", func(writer http.ResponseWriter, request *http.Request) {
		mux.Lock()
		defer mux.Unlock()
		vars, ok := request.Context().Value(httprouter.ParamsKey).(httprouter.Params)
		if !ok {
			debug.PrintStack()
			log.Fatal("nope")
		}
		if data[vars.ByName("engine")] == nil {
			data[vars.ByName("engine")] = make(map[string]interface{})
		}
		delete(data[vars.ByName("engine")], vars.ByName("key"))
	})

	router.HandlerFunc(http.MethodGet, "/v1/:engine/metadata", func(writer http.ResponseWriter, request *http.Request) {
		mux.Lock()
		defer mux.Unlock()
		vars, ok := request.Context().Value(httprouter.ParamsKey).(httprouter.Params)
		if !ok {
			debug.PrintStack()
			log.Fatal("nope")
		}
		keys := []string{}
		for k, _ := range data[vars.ByName("engine")] {
			keys = append(keys, k)
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(map[string]interface{}{
			"data": map[string]interface{}{"keys": keys},
		})
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
	})

	router.HandlerFunc(http.MethodPut, "/v1/:engine/data/:key", func(writer http.ResponseWriter, request *http.Request) {
		mux.Lock()
		defer mux.Unlock()
		vars, ok := request.Context().Value(httprouter.ParamsKey).(httprouter.Params)
		if !ok {
			debug.PrintStack()
			log.Fatal("nope")
		}
		defer request.Body.Close()
		requestData := map[string]interface{}{}
		_ = json.NewDecoder(request.Body).Decode(&requestData)
		if data[vars.ByName("engine")] == nil {
			data[vars.ByName("engine")] = make(map[string]interface{})
		}
		data[vars.ByName("engine")][vars.ByName("key")] = requestData["data"]
	})

	return
}
