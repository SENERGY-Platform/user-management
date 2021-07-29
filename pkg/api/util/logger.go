/*
 * Copyright 2019 InfAI (CC SES)
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

package util

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func NewLogger(handler http.Handler, logLevel string) *LoggerMiddleWare {
	return &LoggerMiddleWare{handler: handler, logLevel: logLevel}
}

type LoggerMiddleWare struct {
	handler  http.Handler
	logLevel string //DEBUG | CALL | NONE
}

func (this *LoggerMiddleWare) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	this.log(r)
	if this.handler != nil {
		this.handler.ServeHTTP(w, r)
	} else {
		http.Error(w, "Forbidden", 403)
	}
}

func (this *LoggerMiddleWare) log(request *http.Request) {
	if this.logLevel != "NONE" {
		method := request.Method
		path := request.URL

		if this.logLevel == "CALL" {
			log.Printf("%v [%v] %v \n", request.RemoteAddr, method, path)
		}

		if this.logLevel == "DEBUG" {
			//read on request.Body would empty it -> create new ReadCloser for request.Body while reading
			var buf bytes.Buffer
			temp := io.TeeReader(request.Body, &buf)
			b, err := ioutil.ReadAll(temp)
			if err != nil {
				log.Println("ERROR: read error in debuglog:", err)
			}
			request.Body = ioutil.NopCloser(bytes.NewReader(buf.Bytes()))

			client := request.RemoteAddr
			log.Printf("%v [%v] %v %v", client, method, path, string(b))

		}

	}
}
