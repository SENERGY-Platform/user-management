package mocks

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

type PermissionsCheck struct {
	Resource string         `json:"resource,omitempty"`
	CheckIds CheckIdsObject `json:"check_ids,omitempty"`
}

type CheckIdsObject struct {
	Ids    []string `json:"ids,omitempty"`
	Rights string   `json:"rights,omitempty"`
}

func PermissionsMock(ctx context.Context, wg *sync.WaitGroup) (url string) {
	server := &httptest.Server{
		Config: &http.Server{Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			log.Println("DEBUG: PermissionsMock: got request", request.URL.String())
			if strings.Contains(request.URL.Path, "jwt/check/") {
				json.NewEncoder(writer).Encode(true)
				return
			}
			if strings.Contains(request.URL.Path, "query") {
				check := PermissionsCheck{}
				json.NewDecoder(request.Body).Decode(&check)
				result := map[string]bool{}
				for _, id := range check.CheckIds.Ids {
					result[id] = true
				}
				json.NewEncoder(writer).Encode(result)
				return
			}
		})},
	}
	server.Listener, _ = net.Listen("tcp", ":")
	server.Start()
	wg.Add(1)
	go func() {
		<-ctx.Done()
		wg.Done()
	}()
	return server.URL
}
