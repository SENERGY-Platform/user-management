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

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"syscall"
)

func main() {
	configLocation := flag.String("config", "config.json", "configuration file")
	flag.Parse()

	conf, err := Load(*configLocation)
	if err != nil {
		log.Fatal("ERROR: unable to load config", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg, err := StartApi(ctx, conf)
	if err != nil {
		cancel()
		if wg != nil {
			wg.Wait()
		}
		log.Fatal(err)
	}

	var shutdownTime time.Time
	go func() {
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
		sig := <-shutdown
		log.Println("received shutdown signal", sig)
		shutdownTime = time.Now()
		cancel()
	}()

	wg.Wait()
	log.Println("Shutdown complete, took", time.Since(shutdownTime))
}
