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

package events

import (
	"sync"
	"reflect"

	"github.com/30x/apid-core"
)

// events published to a given channel are processed entirely in order, though delivery to listeners is async

type eventManager struct {
	sync.Mutex
	dispatchers map[apid.EventSelector]*dispatcher
}

func (em *eventManager) Emit(selector apid.EventSelector, event apid.Event) chan apid.Event {

	log.Debugf("emit selector: '%s' event %v: %v", selector, &event, event)

	responseChannel := make(chan apid.Event, 1)
	em.EmitWithCallback(selector, event, func(event apid.Event) {
		responseChannel <- event
	})
	return responseChannel
}

func (em *eventManager) EmitWithCallback(selector apid.EventSelector, event apid.Event, callback apid.EventHandlerFunc) {
	log.Debugf("emit with callback selector: '%s' event %v: %v", selector, &event, event)

	handler := &funcWrapper{em, nil}
	handler.HandlerFunc = func(e apid.Event) {
		if ede, ok := e.(apid.EventDeliveryEvent); ok {
			if reflect.DeepEqual(ede.Event, event) {
				em.StopListening(apid.EventDeliveredSelector, handler)
				callback(e)
			}
		}
	}

	em.Listen(apid.EventDeliveredSelector, handler)

	em.Lock()
	dispatch := em.dispatchers[selector]
	em.Unlock()

	if !dispatch.Send(event) {
		em.sendDelivered(selector, event, 0) // in case of no dispatcher
	}
}

func (em *eventManager) HasListeners(selector apid.EventSelector) bool {
	em.Lock()
	dispatch := em.dispatchers[selector]
	em.Unlock()
	return dispatch.HasHandlers()
}

func (em *eventManager) Listen(selector apid.EventSelector, handler apid.EventHandler) {
	em.Lock()
	defer em.Unlock()
	log.Debugf("listen: '%s' handler: %v", selector, handler)
	if em.dispatchers == nil {
		em.dispatchers = make(map[apid.EventSelector]*dispatcher)
	}
	if em.dispatchers[selector] == nil {
		d := &dispatcher{sync.Mutex{}, em, selector, nil, nil}
		em.dispatchers[selector] = d
	}
	em.dispatchers[selector].Add(handler)
}

func (em *eventManager) StopListening(selector apid.EventSelector, handler apid.EventHandler) {
	em.Lock()
	defer em.Unlock()
	log.Debugf("stop listening: '%s' handler: %v", selector, handler)
	if em.dispatchers == nil {
		return
	}
	em.dispatchers[selector].Remove(handler)
}

func (em *eventManager) ListenFunc(selector apid.EventSelector, handlerFunc apid.EventHandlerFunc) {
	log.Debugf("listenFunc: '%s' handler: %v", selector, handlerFunc)
	handler := &funcWrapper{em, handlerFunc}
	em.Listen(selector, handler)
}

func (em *eventManager) ListenOnceFunc(selector apid.EventSelector, handlerFunc apid.EventHandlerFunc) {
	log.Debugf("listenOnceFunc: '%s' handler: %v", selector, handlerFunc)
	handler := &funcWrapper{em, nil}
	handler.HandlerFunc = func(event apid.Event) {
		em.StopListening(selector, handler)
		handlerFunc(event)
	}
	em.Listen(selector, handler)
}

func (em *eventManager) Close() {
	em.Lock()
	dispatchers := em.dispatchers
	em.dispatchers = nil
	em.Unlock()
	log.Debugf("Closing %d dispatchers", len(dispatchers))
	for _, dispatcher := range dispatchers {
		dispatcher.Close()
	}
}

func (em *eventManager) sendDelivered(selector apid.EventSelector, event apid.Event, count int) {
	if selector != apid.EventDeliveredSelector {
		ede := apid.EventDeliveryEvent{
			Description: "event complete",
			Selector:    selector,
			Event:       event,
			Count:       count,
		}
		em.Lock()
		defer em.Unlock()
		em.dispatchers[apid.EventDeliveredSelector].Send(ede)
	}
}

type dispatcher struct {
	sync.Mutex
	em       *eventManager
	selector apid.EventSelector
	channel  chan apid.Event
	handlers []apid.EventHandler
}

func (d *dispatcher) Add(h apid.EventHandler) {
	if d == nil {
		return
	}
	d.Lock()
	defer d.Unlock()
	if d.handlers == nil {
		d.handlers = []apid.EventHandler{h}
		d.channel = make(chan apid.Event, config.GetInt(configChannelBufferSize))
		d.startDelivery()
		return
	}
	cp := make([]apid.EventHandler, len(d.handlers)+1)
	copy(cp, d.handlers)
	cp[len(d.handlers)] = h
	d.handlers = cp
}

func (d *dispatcher) Remove(h apid.EventHandler) {
	if d == nil {
		return
	}
	d.Lock()
	defer d.Unlock()
	for i := len(d.handlers) - 1; i >= 0; i-- {
		ih := d.handlers[i]
		if h == ih {
			d.handlers = append(d.handlers[:i], d.handlers[i+1:]...)
			return
		}
	}
}

func (d *dispatcher) Close() {
	if d == nil {
		return
	}
	close(d.channel)
}

func (d *dispatcher) Send(e apid.Event) bool {
	if d == nil {
		return false
	}
	defer func() {
		if err := recover(); err != nil {
			log.Warnf("Send %v failed: %v", e, err)
		}
	}()
	d.channel <- e
	return true
}

func (d *dispatcher) HasHandlers() bool {
	if d == nil {
		return false
	}
	d.Lock()
	defer d.Unlock()
	return d != nil && len(d.handlers) > 0
}

func (d *dispatcher) startDelivery() {
	if d == nil {
		return
	}
	go func() {
		for {
			select {
			case event := <-d.channel:
				if event != nil {
					d.Lock()
					handlers := d.handlers
					d.Unlock()
					log.Debugf("delivering %v to %v", &event, handlers)
					if len(handlers) > 0 {
						var wg sync.WaitGroup
						for _, h := range handlers {
							handler := h
							wg.Add(1)
							go func() {
								defer wg.Done()
								handler.Handle(event) // todo: recover on error?
							}()
						}
						log.Debugf("waiting for handlers")
						wg.Wait()
					}
					d.em.sendDelivered(d.selector, event, len(handlers))
					log.Debugf("event %v delivered", &event)
				}
			}

		}
	}()
}

type funcWrapper struct {
	*eventManager
	HandlerFunc apid.EventHandlerFunc
}

func (r *funcWrapper) Handle(e apid.Event) {
	r.HandlerFunc(e)
}
