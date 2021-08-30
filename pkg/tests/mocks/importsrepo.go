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

func ImportsRepoMock(ctx context.Context, wg *sync.WaitGroup) (url string) {
	server := &httptest.Server{
		Config: &http.Server{Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			log.Println("DEBUG: ImportsRepoMock: got request", request.URL.String())
			if strings.Contains(request.URL.Path, "import-types/") {
				parts := strings.Split(request.URL.Path, "/")
				id := parts[len(parts)-1]
				json.NewEncoder(writer).Encode(map[string]interface{}{
					"id":    id,
					"name":  id,
					"image": "docker.io/library/hello-world",
				})
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
