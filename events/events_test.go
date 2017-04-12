package events_test

import (
	"sync"
	"sync/atomic"

	"github.com/30x/apid-core"
	"github.com/30x/apid-core/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"strconv"
)

var _ = Describe("Events Service", func() {

	It("should ignore event with no listeners", func() {
		em := events.CreateService()
		defer em.Close()
		em.Emit("no listeners", &test_event{"test"})
	})

	It("should publish an event to a listener", func(done Done) {
		em := events.CreateService()

		h := test_handler{
			"handler",
			func(event apid.Event) {
				defer GinkgoRecover()

				em.Close()
				close(done)
			},
		}

		em.Listen("selector", &h)
		em.Emit("selector", &test_event{"test"})
	})

	It("should publish an event to a listener func", func(done Done) {
		em := events.CreateService()

		h := func(event apid.Event) {
			defer GinkgoRecover()

			em.Close()
			close(done)
		}

		em.ListenFunc("selector", h)
		em.Emit("selector", &test_event{"test"})
	})

	It("should publish multiple events to a listener", func(done Done) {
		em := events.CreateService()

		count := int32(0)
		h := test_handler{
			"handler",
			func(event apid.Event) {
				defer GinkgoRecover()

				c := atomic.AddInt32(&count, 1)
				if c > 1 {
					em.Close()
					close(done)
				}
			},
		}

		em.Listen("selector", &h)
		em.Emit("selector", &test_event{"test1"})
		em.Emit("selector", &test_event{"test2"})
	})

	It("EmitWithCallback should call the callback when done with delivery", func(done Done) {
		em := events.CreateService()

		delivered := func(event apid.Event) {
			defer GinkgoRecover()
			close(done)
		}

		em.EmitWithCallback("selector", &test_event{"test1"}, delivered)
	})

	It("should publish only one event to a listenOnce", func(done Done) {
		em := events.CreateService()

		count := 0
		h := func(event apid.Event) {
			defer GinkgoRecover()
			count++
		}

		delivered := func(event apid.Event) {
			defer GinkgoRecover()
			Expect(count).To(Equal(1))
			em.Close()
			close(done)
		}

		em.ListenOnceFunc("selector", h)
		em.Emit("selector", &test_event{"test1"})
		em.EmitWithCallback("selector", &test_event{"test2"}, delivered)
	})

	It("should publish an event to multiple listeners", func(done Done) {
		em := events.CreateService()
		defer em.Close()

		mut := sync.Mutex{}
		hitH1 := false
		hitH2 := false
		h1 := test_handler{
			"handler 1",
			func(event apid.Event) {
				defer GinkgoRecover()
				mut.Lock()
				defer mut.Unlock()
				hitH1 = true
				if hitH1 && hitH2 {
					em.Close()
					close(done)
				}
			},
		}
		h2 := test_handler{
			"handler 2",
			func(event apid.Event) {
				defer GinkgoRecover()

				mut.Lock()
				defer mut.Unlock()
				hitH2 = true
				if hitH1 && hitH2 {
					em.Close()
					close(done)
				}
			},
		}

		em.Listen("selector", &h1)
		em.Listen("selector", &h2)
		em.Emit("selector", &test_event{"test"})
	})

	It("should publish an event delivered event", func(done Done) {
		em := events.CreateService()
		testEvent := &test_event{"test"}
		var testSelector apid.EventSelector = "selector"

		dummy := func(event apid.Event) {}
		em.ListenFunc(testSelector, dummy)

		h := test_handler{
			"event delivered handler",
			func(event apid.Event) {
				defer GinkgoRecover()

				e, ok := event.(apid.EventDeliveryEvent)

				Expect(ok).To(BeTrue())
				Expect(e.Event).To(Equal(testEvent))
				Expect(e.Selector).To(Equal(testSelector))

				em.Close()
				close(done)
			},
		}

		em.Listen(apid.EventDeliveredSelector, &h)
		em.Emit(testSelector, testEvent)
	})

	It("should be able to remove a listener", func(done Done) {
		em := events.CreateService()

		event1 := &test_event{"test1"}
		event2 := &test_event{"test2"}
		event3 := &test_event{"test3"}

		dummy := func(event apid.Event) {}
		em.ListenFunc("selector", dummy)

		h := test_handler{
			"handler",
			func(event apid.Event) {
				defer GinkgoRecover()

				Expect(event).NotTo(Equal(event2))
				if event == event3 {
					em.Close()
					close(done)
				}
			},
		}
		em.Listen("selector", &h)

		// need to drive test like this because of async delivery
		td := test_handler{
			"test driver",
			func(event apid.Event) {
				defer GinkgoRecover()

				e := event.(apid.EventDeliveryEvent)
				if e.Event == event1 {
					em.StopListening("selector", &h)
					em.Emit("selector", event2)
				} else if e.Event == event2 {
					em.Listen("selector", &h)
					em.Emit("selector", event3)
				}
			},
		}
		em.Listen(apid.EventDeliveredSelector, &td)

		em.Emit("selector", event1)
	})

	It("should deliver events according selector", func(done Done) {
		em := events.CreateService()

		e1 := &test_event{"test1"}
		e2 := &test_event{"test2"}

		count := int32(0)

		h1 := test_handler{
			"handler1",
			func(event apid.Event) {
				defer GinkgoRecover()

				c := atomic.AddInt32(&count, 1)
				Expect(event).Should(Equal(e1))
				if c == 2 {
					em.Close()
					close(done)
				}
			},
		}

		h2 := test_handler{
			"handler2",
			func(event apid.Event) {
				defer GinkgoRecover()

				c := atomic.AddInt32(&count, 1)
				Expect(event).Should(Equal(e2))
				if c == 2 {
					em.Close()
					close(done)
				}
			},
		}

		em.Listen("selector1", &h1)
		em.Listen("selector2", &h2)

		em.Emit("selector2", e2)
		em.Emit("selector1", e1)
	})

	It("should publish PluginsInitialized event", func(done Done) {
		xData := make(map[string]interface{})
		xData["schemaVersion"] = "1.2.3"
		p := func(s apid.Services) (pd apid.PluginData, err error) {
			pd = apid.PluginData{
				Name:      "test plugin",
				Version:   "1.0.0",
				ExtraData: xData,
			}
			return
		}
		apid.RegisterPlugin(p)

		h := func(event apid.Event) {
			defer GinkgoRecover()

			if pie, ok := event.(apid.PluginsInitializedEvent); ok {

				apid.Events().Close()

				Expect(len(pie.Plugins)).Should(Equal(1))
				p := pie.Plugins[0]
				Expect(p.Name).To(Equal("test plugin"))
				Expect(p.Version).To(Equal("1.0.0"))
				Expect(p.ExtraData["schemaVersion"]).To(Equal("1.2.3"))

				close(done)
			}
		}
		apid.Events().ListenFunc(apid.SystemEventsSelector, h)

		apid.InitializePlugins("")
	})

	It("shutdown event should be emitted and listened successfully", func(done Done) {
		h := func(event apid.Event) {
			defer GinkgoRecover()

			if pie, ok := event.(apid.ShutdownEvent); ok {
				apid.Events().Close()
				Expect(pie.Description).Should(Equal("apid is going to shutdown"))
				fmt.Println(pie.Description)
				close(done)
			} else {
				Fail("Received wrong event")
			}
		}
		apid.Events().ListenFunc(apid.ShutdownEventSelector, h)
		shutdownHandler := func(event apid.Event) {}
		apid.Events().EmitWithCallback(apid.ShutdownEventSelector, apid.ShutdownEvent{"apid is going to shutdown"}, shutdownHandler)
	})

	It("handlers registered by plugins should execute before apid shutdown", func(done Done) {
		pluginNum := 10
		count := int32(0)

		// create and register plugins, listen to shutdown event
		for i:=0; i<pluginNum; i++ {
			apid.RegisterPlugin(createDummyPlugin(i))
			h := func(event apid.Event) {
				if pie, ok := event.(apid.ShutdownEvent); ok {
					Expect(pie.Description).Should(Equal("apid is going to shutdown"))
					atomic.AddInt32(&count, 1)
				} else {
					Fail("Received wrong event")
				}
			}
			apid.Events().ListenFunc(apid.ShutdownEventSelector, h)
		}


		apid.InitializePlugins("")

		apid.ShutdownPluginsAndWait()

		defer GinkgoRecover()
		apid.Events().Close()

		// handlers of all registered plugins have executed
		Expect(count).Should(Equal(int32(pluginNum)))

		close(done)
	})

	It("should be able to read apid version from PluginsInitialized event", func(done Done) {
		xData := make(map[string]interface{})
		xData["schemaVersion"] = "1.2.3"
		p := func(s apid.Services) (pd apid.PluginData, err error) {
			pd = apid.PluginData{
				Name:      "test plugin",
				Version:   "1.0.0",
				ExtraData: xData,
			}
			return
		}
		apid.RegisterPlugin(p)

		apidVersion := "dummy_version"

		h := func(event apid.Event) {
			defer GinkgoRecover()

			if pie, ok := event.(apid.PluginsInitializedEvent); ok {

				apid.Events().Close()
				Expect(pie.ApidVersion).To(Equal(apidVersion))
				close(done)
			}
		}
		apid.Events().ListenFunc(apid.SystemEventsSelector, h)

		apid.InitializePlugins(apidVersion)
	})
})

func createDummyPlugin(id int) apid.PluginInitFunc{
	xData := make(map[string]interface{})
	xData["schemaVersion"] = "1.2.3"
	p := func(s apid.Services) (pd apid.PluginData, err error) {
		pd = apid.PluginData{
			Name: "test plugin " + strconv.Itoa(id),
			Version: "1.0.0",
			ExtraData: xData,
		}
		return
	}
	return p

}

type test_handler struct {
	description string
	f           func(event apid.Event)
}

func (t *test_handler) String() string {
	return t.description
}

func (t *test_handler) Handle(event apid.Event) {
	t.f(event)
}

type test_event struct {
	description string
}
