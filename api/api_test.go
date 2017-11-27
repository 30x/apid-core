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

package api_test

import (
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"net/url"
)

var _ = Describe("API Service", func() {

	It("should return vars from /exp/vars with request counter", func() {

		uri, err := url.Parse(testServer.URL)
		Expect(err).NotTo(HaveOccurred())
		uri.Path = "/exp/vars"

		resp, err := http.Get(uri.String())
		Expect(err).ShouldNot(HaveOccurred())
		defer resp.Body.Close()
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))

		body, err := ioutil.ReadAll(resp.Body)
		var m map[string]interface{}
		err = json.Unmarshal(body, &m)
		Expect(err).ShouldNot(HaveOccurred())

		requests := m["requests"].(map[string]interface{})
		Expect(requests["/exp/vars"]).Should(Equal(float64(1)))
	})
})
