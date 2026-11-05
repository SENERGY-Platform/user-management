/*
 * Copyright 2026 InfAI (CC SES)
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
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
)

// FlowEngineMock stands in for analytics-flow-engine, which needs an in-cluster Kubernetes API since it
// lost its rancher driver. Like the engine, it registers a posted pipeline in the analytics-pipeline
// registry, which is the state user-management deletes; it deploys no operators.
func FlowEngineMock(ctx context.Context, wg *sync.WaitGroup, pipelineRegistryUrl string) (url string) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		log.Println("DEBUG: FlowEngineMock: got request", request.Method, request.URL.String())
		if request.Method != http.MethodPost || request.URL.Path != "/pipeline" {
			http.Error(writer, "not implemented in FlowEngineMock", http.StatusNotImplemented)
			return
		}
		req, err := http.NewRequestWithContext(request.Context(), http.MethodPost, pipelineRegistryUrl+"/pipeline", request.Body)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", request.Header.Get("Authorization"))
		req.Header.Set("X-UserId", request.Header.Get("X-UserId"))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		writer.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		writer.WriteHeader(resp.StatusCode)
		io.Copy(writer, resp.Body)
	}))

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		server.Close()
	}()
	return server.URL
}
