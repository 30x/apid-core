// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"github.com/google/uuid"
	"net/http"
	"net/url"
	"time"
)

const ConfigfwdProxyPortURL = "configcompletefwdp"

func IsValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

func GenerateUUID() string {
	return uuid.New().String()
}

// Helper method that initializes the roundtripper based on the configuration parameters.
func Transport(pURL string) *http.Transport {
	var tr http.Transport
	if pURL != "" {
		paURL, err := url.Parse(pURL)
		if err != nil {
			panic("Error parsing proxy URL")
		}
		tr = http.Transport{
			Proxy: http.ProxyURL(paURL),
		}
	}
	return &tr
}

// distributeEvents() receives elements from deliverChan, and send them to subscribers
// Sending a `chan interface{}` to addSubscriber adds a new subscriber.
// It closes the subscriber channel after sending the element.
// `go DistributeEvents(deliverChan, addSubscriber)` should be called during API initialization.
// Any subscriber sent to `addSubscriber` should be buffered chan.
func DistributeEvents(deliverChan <-chan interface{}, addSubscriber chan chan interface{}) {
	subscribers := make([]chan interface{}, 0)
	for {
		select {
		case element, ok := <-deliverChan:
			if !ok {
				return
			}
			for _, subscriber := range subscribers {
				go func(sub chan interface{}) {
					sub <- element
					close(sub)
				}(subscriber)
			}
			subscribers = make([]chan interface{}, 0)
		case sub, ok := <-addSubscriber:
			if !ok {
				return
			}
			subscribers = append(subscribers, sub)
		}
	}
}

// LongPolling() subscribes to `addSubscriber`, and do long-polling until anything is delivered.
// It calls `successHandler` if receives a notification.
// It calls `timeoutHandler` if there's a timeout.
// `go DistributeEvents(deliverChan, addSubscriber)` must have been called during API initialization.
func LongPolling(w http.ResponseWriter, timeout time.Duration, addSubscriber chan chan interface{}, successHandler func(interface{}, http.ResponseWriter), timeoutHandler func(http.ResponseWriter)) {
	notifyChan := make(chan interface{}, 1)
	addSubscriber <- notifyChan
	select {
	case n := <-notifyChan:
		successHandler(n, w)
	case <-time.After(timeout):
		timeoutHandler(w)
	}
}

// Debounce() packs all elements received from channel `inChan` within the specified time window to one slice,
// and send it to channel `outChan` periodically. If nothing is received in the time window, nothing will be sent to `outChan`.
func Debounce(inChan chan interface{}, outChan chan []interface{}, window time.Duration) {
	send := func(toSend []interface{}) {
		if toSend != nil {
			outChan <- toSend
		}
	}
	var toSend []interface{} = nil
	for {
		select {
		case incoming, ok := <-inChan:
			if ok {
				toSend = append(toSend, incoming)
			} else {
				send(toSend)
				close(outChan)
				return
			}
		case <-time.After(window):
			send(toSend)
			toSend = nil
		}
	}
}

// Contains return whether the target string is found in the slice.
func Contains(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}
