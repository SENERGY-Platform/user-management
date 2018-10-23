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
	"errors"
	"log"
	"sync"

	"time"

	"github.com/streadway/amqp"
)

type AmqpConn struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	mux     sync.Mutex

	url       string
	resources []string

	consumer map[string]ConsumerInfo

	reconnectTimeout time.Duration
}

type ConsumerInfo struct {
	Worker   ConsumerFunc
	QName    string
	Resource string
}

type ConsumerFunc func(delivery []byte) error

func InitAmqpConn(url string, resources []string, reconnectTimeout int64) (result *AmqpConn, err error) {
	result = &AmqpConn{}
	err = result.Init(url, resources, time.Duration(reconnectTimeout)*time.Second, map[string]ConsumerInfo{})
	return
}

func (this *AmqpConn) Init(url string, resources []string, reconnectTimeout time.Duration, consumer map[string]ConsumerInfo) (err error) {
	this.url = url
	this.resources = resources
	this.reconnectTimeout = reconnectTimeout
	this.consumer = consumer
	this.conn, err = amqp.Dial(this.url)
	if err != nil {
		return
	}
	err = this.initChannel(resources)
	if err != nil {
		this.Close()
		return
	}
	for name := range this.consumer {
		err = this.consume(name)
		if err != nil {
			this.Close()
			return
		}
	}
	return
}

func (this *AmqpConn) initChannel(resources []string) (err error) {
	log.Println("init channel")
	this.channel, err = this.conn.Channel()
	if err != nil {
		return
	}
	err = this.declareResources()
	if err != nil {
		return
	}
	this.channel.NotifyClose(this.handleError())
	return
}

func (this *AmqpConn) handleError() (ch chan *amqp.Error) {
	ch = make(chan *amqp.Error)
	go func() {
		log.Println("start error handler")
		for err := range ch {
			log.Println("receive amqp close", err)
			if err != nil {
				this.mux.Lock()
				defer this.mux.Unlock()
				this.Close()
				for {
					log.Println("try reconnecting")
					err := this.Init(this.url, this.resources, this.reconnectTimeout, this.consumer)
					if err != nil {
						log.Println("unable to reconnect", err)
						log.Println("try again in ", this.reconnectTimeout.String())
						time.Sleep(this.reconnectTimeout)
					} else {
						return
					}
				}
			}
		}
		log.Println("stop error handler")
	}()
	return ch
}

func (this *AmqpConn) declareResources() (err error) {
	for _, name := range this.resources {
		err = this.declareResource(name)
		if err != nil {
			log.Println("ERROR: while declaring queue", err)
			return err
		}
	}
	return
}

func (this *AmqpConn) declareResource(name string) (err error) {
	log.Println("init exchange ", name)
	err = this.channel.ExchangeDeclare(name, "fanout", true, false, false, false, nil)
	return err
}

func (this *AmqpConn) Close() {
	log.Println("close amqp conn")
	this.channel.Close()
	this.conn.Close()
}

func (this *AmqpConn) Publish(resource string, payload []byte) error {
	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
		Body:         payload,
	}
	return this.UseChannel(func(channel *amqp.Channel) {
		channel.Publish(resource, "", false, false, msg)
	})
}

//locks channel to worker to safely reinitiate it if channel closing error occurs
func (this *AmqpConn) UseChannel(worker func(channel *amqp.Channel)) (err error) {
	this.mux.Lock()
	defer this.mux.Unlock()
	worker(this.channel)
	return
}

func (this *AmqpConn) Consume(qname string, resource string, worker ConsumerFunc) (err error) {
	log.Println("init consumer for ", resource)
	this.mux.Lock()
	defer this.mux.Unlock()
	this.consumer[resource] = ConsumerInfo{QName: qname, Resource: resource, Worker: worker}
	err = this.consume(resource)
	//remove added consumer if consumption fails
	if err != nil {
		delete(this.consumer, resource)
	}
	return
}

func (this *AmqpConn) consume(consumerKey string) (err error) {
	consumerinfo, ok := this.consumer[consumerKey]
	if !ok {
		return errors.New("no consumer info for given resource " + consumerKey)
	}
	log.Printf("use %s queue to consume %s\n", consumerinfo.QName, consumerinfo.Resource)
	q, err := this.channel.QueueDeclare(consumerinfo.QName, true, false, false, false, nil)
	if err != nil {
		return err
	}
	err = this.channel.Qos(1, 0, true)
	if err != nil {
		return err
	}
	err = this.channel.QueueBind(q.Name, "", consumerinfo.Resource, false, nil)
	if err != nil {
		return err
	}
	msgs, err := this.channel.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go this.runworker(q.Name, msgs, consumerinfo.Worker)
	return nil
}

func (this *AmqpConn) runworker(qname string, deliveries <-chan amqp.Delivery, consumerFunc ConsumerFunc) {
	for msg := range deliveries {
		log.Println("amqp receive", qname, string(msg.Body))
		err := consumerFunc(msg.Body)
		if err != nil {
			log.Println("error while processing msg; message consumtion will not be committed", err)
			err = msg.Reject(true)
			log.Println("DEBUG: wait after reject")
			time.Sleep(3 * time.Second)
			if err != nil {
				log.Println("ERROR while rejecting msg", err)
			}
		} else {
			err = msg.Ack(false)
			if err != nil {
				log.Println("ERROR while acknowledging msg", err)
			}
		}
	}
}
