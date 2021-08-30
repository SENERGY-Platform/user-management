package mocks

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

func PermissionsMock(ctx context.Context, wg *sync.WaitGroup) (url string) {
	server := &httptest.Server{
		Config: &http.Server{Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if strings.Contains(request.URL.Path, "jwt/check/") {
				json.NewEncoder(writer).Encode(true)
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
