/*
 * Copyright 2021 InfAI (CC SES)
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
