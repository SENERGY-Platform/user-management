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
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

// LateProxy forwards every request to a target that is set after the proxy started. It lets two
// containers that need each other's address at startup be started one after the other.
type LateProxy struct {
	URL    string
	target atomic.Pointer[url.URL]
}

func NewLateProxy(ctx context.Context, wg *sync.WaitGroup) *LateProxy {
	proxy := &LateProxy{}
	reverse := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(proxy.target.Load())
		},
	}
	server := &httptest.Server{
		Config: &http.Server{Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if proxy.target.Load() == nil {
				http.Error(writer, "LateProxy has no target yet", http.StatusServiceUnavailable)
				return
			}
			reverse.ServeHTTP(writer, request)
		})},
	}
	// reachable from the containers through the docker host, like the other mocks
	server.Listener, _ = net.Listen("tcp", ":")
	server.Start()
	proxy.URL = server.URL

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		server.Close()
	}()
	return proxy
}

func (this *LateProxy) SetTarget(target string) error {
	u, err := url.Parse(target)
	if err != nil {
		return err
	}
	this.target.Store(u)
	return nil
}
