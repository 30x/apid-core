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

	"github.com/apid/apid-core/util"
	"regexp"
	"testing"
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
})
