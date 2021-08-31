package mocks

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
)

func RancherMock(ctx context.Context, wg *sync.WaitGroup) (url string) {
	server := &httptest.Server{
		Config: &http.Server{Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			log.Println("DEBUG: RancherMock: got request", request.Method, request.URL.String())
			if request.Method == http.MethodPost {
				writer.WriteHeader(http.StatusCreated)
				return
			}
			if request.Method == http.MethodDelete {
				writer.WriteHeader(http.StatusNoContent)
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
