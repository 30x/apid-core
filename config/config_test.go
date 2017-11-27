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

package config_test

import (
	"github.com/apid/apid-core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"time"
)

var _ = Describe("Config Service", func() {

	Context("lowercase env var", func() {

		It("no env var", func() {
			Expect(apid.Config().Get("test")).To(Equal("test"))
		})

		It("as interface", func() {
			os.Setenv("apid_test", "TEST")
			Expect(apid.Config().Get("test")).To(Equal("TEST"))
		})

		It("as bool", func() {
			os.Setenv("apid_test", "true")
			Expect(apid.Config().GetBool("test")).To(BeTrue())
		})

		It("as float", func() {
			os.Setenv("apid_test", "64.1")
			Expect(apid.Config().GetFloat64("test")).To(Equal(64.1))
		})

		It("as int", func() {
			os.Setenv("apid_test", "64")
			Expect(apid.Config().GetInt("test")).To(Equal(64))
		})

		It("as string", func() {
			os.Setenv("apid_test", "TEST")
			Expect(apid.Config().GetString("test")).To(Equal("TEST"))
		})

		It("as duration", func() {
			os.Setenv("apid_test", "300ms")
			Expect(apid.Config().GetDuration("test")).To(Equal(300 * time.Millisecond))
		})

		It("as IsSet", func() {
			os.Setenv("apid_test", "300ms")
			Expect(apid.Config().IsSet("test")).To(BeTrue())
		})
	})
})
