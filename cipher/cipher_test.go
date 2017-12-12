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

package cipher_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/apid/apid-core/cipher"
	"testing"
)

func TestEvents(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cipher Suite")
}

var _ = Describe("APID Cipher", func() {

	Context("AES", func() {

		Context("AES/ECB/PKCS7Padding Encrypt/Decrypt", func() {
			type testData struct {
				key        []byte
				plaintext  []byte
				ciphertext []byte
			}

			data := []testData{
				{
					// 128-bit
					[]byte{2, 122, 212, 83, 150, 164, 180, 4, 148, 242, 65, 189, 3, 188, 76, 247},
					[]byte("aUWQKgAwmaR0p2kY"),
					// 32-byte after padding
					[]byte{218, 53, 247, 87, 119, 80, 231, 16, 125, 11, 214, 101, 246, 202, 178, 163, 202, 102,
						146, 245, 79, 215, 74, 228, 17, 83, 213, 134, 105, 203, 31, 14},
				},
				{
					// 192-bit
					[]byte{2, 122, 212, 83, 150, 164, 180, 4, 148, 242, 65, 189, 3, 188, 76, 247,
						2, 122, 212, 83, 150, 164, 180, 4},
					[]byte("a"),
					// 16-byte after padding
					[]byte{225, 2, 177, 65, 152, 88, 116, 43, 71, 215, 84, 240, 221, 175, 11, 131},
				},
				{
					// 256-bit
					[]byte{2, 122, 212, 83, 150, 164, 180, 4, 148, 242, 65, 189, 3, 188, 76, 247,
						2, 122, 212, 83, 150, 164, 180, 4, 148, 242, 65, 189, 3, 188, 76, 247},
					[]byte(""),
					// 16-byte after padding
					[]byte{88, 192, 164, 235, 153, 89, 14, 134, 224, 122, 31, 36, 238, 117, 121, 117},
				},
			}
			It("Encrypt", func() {
				for i := 0; i < len(data); i++ {
					c, err := cipher.CreateAesCipher(data[i].key)
					Expect(err).Should(Succeed())
					Expect(c.Encrypt(data[i].plaintext, cipher.ModeEcb, cipher.PaddingPKCS5)).Should(Equal(data[i].ciphertext))
					Expect(c.Encrypt(data[i].plaintext, cipher.ModeEcb, cipher.PaddingPKCS7)).Should(Equal(data[i].ciphertext))
				}
			})

			It("Decrypt", func() {
				for i := 0; i < len(data); i++ {
					c, err := cipher.CreateAesCipher(data[i].key)
					Expect(err).Should(Succeed())
					Expect(c.Encrypt(data[i].plaintext, cipher.ModeEcb, cipher.PaddingPKCS5)).Should(Equal(data[i].ciphertext))
					Expect(c.Encrypt(data[i].plaintext, cipher.ModeEcb, cipher.PaddingPKCS7)).Should(Equal(data[i].ciphertext))
				}
			})
		})

		It("SetKey", func() {
			key := make([]byte, 16)
			plaintext := []byte("aUWQKgAwmaR0p2kY")
			ciphertext := []byte{218, 53, 247, 87, 119, 80, 231, 16, 125, 11, 214, 101, 246, 202, 178, 163, 202, 102, 146, 245, 79, 215, 74, 228, 17, 83, 213, 134, 105, 203, 31, 14}
			c, err := cipher.CreateAesCipher(key)
			Expect(err).Should(Succeed())
			key = []byte{2, 122, 212, 83, 150, 164, 180, 4, 148, 242, 65, 189, 3, 188, 76, 247}
			Expect(c.SetKey(key)).Should(Succeed())
			Expect(c.Encrypt(plaintext, cipher.ModeEcb, cipher.PaddingPKCS5)).Should(Equal(ciphertext))
			Expect(c.Decrypt(ciphertext, cipher.ModeEcb, cipher.PaddingPKCS7)).Should(Equal(plaintext))
		})

		It("Invalid Parameters", func() {
			_, err := cipher.CreateAesCipher(make([]byte, 15))
			Expect(err).ToNot(Succeed())
			_, err = cipher.CreateAesCipher(nil)
			Expect(err).ToNot(Succeed())
			key := make([]byte, 16)
			c, err := cipher.CreateAesCipher(key)
			Expect(err).Should(Succeed())
			_, err = c.Encrypt([]byte{1, 2, 3}, cipher.Mode("unsupported"), cipher.PaddingPKCS7)
			Expect(err).ToNot(Succeed())
			_, err = c.Encrypt([]byte{1, 2, 3}, cipher.ModeEcb, cipher.Padding("unsupported"))
			Expect(err).ToNot(Succeed())
			_, err = c.Decrypt([]byte{88, 192, 164, 235, 153, 89, 14, 134, 224, 122, 31, 36, 238, 117, 121, 117},
				cipher.Mode("unsupported"), cipher.PaddingPKCS7)
			Expect(err).ToNot(Succeed())
			_, err = c.Decrypt([]byte{88, 192, 164, 235, 153, 89, 14, 134, 224, 122, 31, 36, 238, 117, 121, 117},
				cipher.ModeEcb, cipher.Padding("unsupported"))
			Expect(err).ToNot(Succeed())
		})
	})
})
