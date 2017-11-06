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

package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"


	"net/http/httptest"
	"github.com/apid/apid-core/util"
	"math/rand"
	"net/http"
	"regexp"
	"sync/atomic"
	"testing"
	"time"
)

var _ = BeforeSuite(func() {
})

func TestEvents(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Util Suite")
}

var _ = Describe("APID utils", func() {

	Context("UUID", func() {
		It("should validate UUID", func() {
			data := []string{
				"fc041ed2-62a7-4086-a66e-bbbc9219fad5",
				"invalid-uuid",
			}
			expected := []bool{
				true,
				false,
			}

			for i := range data {
				Ω(util.IsValidUUID(data[i])).Should(Equal(expected[i]))
			}

		})

		It("should generate valid UUID", func() {
			r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
			Ω(r.MatchString(util.GenerateUUID())).Should(BeTrue())
			Ω(util.IsValidUUID(util.GenerateUUID())).Should(BeTrue())
		})
	})

	Context("Forward Proxy Protocol", func() {
		It("Verify Forward proxying to server works", func() {
			var maxIdleConnsPerHost = 10
			var tr *http.Transport
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Fail("Cant come here, as we have not forwarded request from fwdPrxyServer")
			}))
			fwdPrxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Header.Get("foo")).Should(Equal("bar"))
				w.Header().Set("bar", "foo")
			}))
			tr = util.Transport(fwdPrxyServer.URL)
			tr.MaxIdleConnsPerHost =  maxIdleConnsPerHost
			var rspcnt int = 0
			ch := make(chan *http.Response)
			client := &http.Client{Transport: tr}
			for i := 0; i < 2*maxIdleConnsPerHost; i++ {
				go func(client *http.Client) {
					defer GinkgoRecover()
					req, err := http.NewRequest("GET", server.URL, nil)
					Expect(err).Should(Succeed())
					req.Header.Set("foo", "bar")
					resp, err := client.Do(req)
					Expect(err).Should(Succeed())
					Expect(resp.Header.Get("bar")).Should(Equal("foo"))
					resp.Body.Close()
					ch <- resp
				}(client)
			}
			for {
				resp := <-ch
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				if rspcnt >= 2*maxIdleConnsPerHost-1 {
					return
				}
				rspcnt++
			}

		}, 3)
	})

	Context("Long polling utils", func() {
		It("DistributeEvents", func() {
			// make test data
			deliverChan := make(chan interface{})
			addSubscriber := make(chan chan interface{})
			subs := make([]chan interface{}, 50+rand.Intn(50))
			for i := range subs {
				subs[i] = make(chan interface{}, 1)
			}

			// test
			go util.DistributeEvents(deliverChan, addSubscriber)

			for i := range subs {
				go func(j int) {
					addSubscriber <- subs[j]
				}(i)
			}

			n := rand.Int()
			closed := new(int32)
			go func(c *int32) {
				for atomic.LoadInt32(c) == 0 {
					deliverChan <- n
				}
			}(closed)

			for i := range subs {
				Ω(<-subs[i]).Should(Equal(n))
			}
			atomic.StoreInt32(closed, 1)
		}, 2)

		It("Long polling", func(done Done) {
			// make test data
			deliverChan := make(chan interface{})
			addSubscriber := make(chan chan interface{})
			go util.DistributeEvents(deliverChan, addSubscriber)
			n := rand.Int()
			closed := new(int32)
			successHandler := func(e interface{}, w http.ResponseWriter) {
				defer GinkgoRecover()
				Ω(w).Should(BeNil())
				Ω(e).Should(Equal(n))
				atomic.StoreInt32(closed, 1)
				close(done)
			}
			timeoutHandler := func(w http.ResponseWriter) {}

			// Long polling
			go util.LongPolling(nil, time.Minute, addSubscriber, successHandler, timeoutHandler)
			go func(c *int32) {
				for atomic.LoadInt32(c) == 0 {
					deliverChan <- n
				}
			}(closed)
		})

		It("Long polling timeout", func(done Done) {
			// make test data
			deliverChan := make(chan interface{})
			addSubscriber := make(chan chan interface{})
			go util.DistributeEvents(deliverChan, addSubscriber)
			successHandler := func(e interface{}, w http.ResponseWriter) {}
			timeoutHandler := func(w http.ResponseWriter) {
				defer GinkgoRecover()
				Ω(w).Should(BeNil())
				close(done)
			}

			// Long polling
			go util.LongPolling(nil, time.Second, addSubscriber, successHandler, timeoutHandler)
		}, 2)

		It("Debounce", func() {
			// make test data
			data := make(map[int]int)
			inChan := make(chan interface{})
			outChan := make(chan []interface{})
			go util.Debounce(inChan, outChan, time.Second)
			for i := 0; i < 5+rand.Intn(5); i++ {
				n := rand.Int()
				data[n]++
				go func(j int) {
					inChan <- j
				}(n)
			}

			// Debounce
			e := <-outChan
			for _, m := range e {
				num, ok := m.(int)
				Ω(ok).Should(BeTrue())
				data[num]--
			}

			for _, v := range data {
				Ω(v).Should(BeZero())
			}
		}, 2)
	})
})
