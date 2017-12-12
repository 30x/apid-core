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
package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"sync"
)

type Mode string
type Padding string

const (
	ModeEcb = Mode("ECB")
)

const (
	PaddingPKCS7 = Padding("PKCS7Padding")
	PaddingPKCS5 = Padding("PKCS5Padding")
)

var (
	supportedModes = map[Mode]bool{
		ModeEcb: true,
	}
	supportedPadding = map[Padding]bool{
		PaddingPKCS7: true,
		PaddingPKCS5: true,
	}
)

func (p Padding) padding(input []byte) (output []byte) {
	switch p {
	case PaddingPKCS5:
		fallthrough
	case PaddingPKCS7:
		//PKCS7Padding
		numPad := 16 - (len(input) % 16)
		output = make([]byte, len(input)+numPad)
		for i := copy(output, []byte(input)); i < len(output); i++ {
			output[i] = byte(numPad)
		}
	}
	return
}

func (p Padding) strip(input []byte) (output []byte) {
	switch p {
	case PaddingPKCS5:
		fallthrough
	case PaddingPKCS7:
		//remove PKCS7Padding
		numPad := int(input[len(input)-1])
		output = input[:(len(input) - numPad)]
	}
	return
}

// Create a new AesCipher object with the specified encryption mode and padding algorithm.
func CreateAesCipher(key []byte) (*AesCipher, error) {
	a := &AesCipher{
		mutex: &sync.RWMutex{},
	}
	if err := a.SetKey(key); err != nil {
		return nil, err
	}
	return a, nil
}

// An object to perform AES encryption/decryption.
type AesCipher struct {
	key   []byte
	block cipher.Block
	mutex *sync.RWMutex
}

// Set/Change the AES key, accepted key's bit-size is 128/192/256.
func (a *AesCipher) SetKey(key []byte) (err error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.key = key
	a.block, err = aes.NewCipher(key)
	return err
}

// Encrypt the plaintext. Padding is performed before encryption, so the
// plaintext input can have any length.
func (a *AesCipher) Encrypt(plaintext []byte, mode Mode, padding Padding) (ciphertext []byte, err error) {
	// check mode
	if !supportedModes[mode] {
		return nil, &ErrModeUnsupported{
			mode: mode,
		}
	}
	// check padding
	if !supportedPadding[padding] {
		return nil, &ErrPaddingUnsupported{
			padding: padding,
		}
	}
	// padding
	text := padding.padding(plaintext)
	ciphertext = text

	// encrypt
	a.mutex.RLock()
	block := a.block
	a.mutex.RUnlock()

	switch mode {
	case ModeEcb:
		size := block.BlockSize()
		for len(text) > 0 {
			block.Encrypt(text, text)
			text = text[size:]
		}
	}
	return
}

// Decrypt the ciphertext. Padding is removed after decryption, so the
// plaintext output can have different length from input ciphertext.
func (a *AesCipher) Decrypt(ciphertext []byte, mode Mode, padding Padding) (plaintext []byte, err error) {
	// check mode
	if !supportedModes[mode] {
		return nil, &ErrModeUnsupported{
			mode: mode,
		}
	}
	// check padding
	if !supportedPadding[padding] {
		return nil, &ErrPaddingUnsupported{
			padding: padding,
		}
	}

	// encrypt
	a.mutex.RLock()
	block := a.block
	a.mutex.RUnlock()

	switch mode {
	case ModeEcb:
		plaintext = make([]byte, len(ciphertext))
		buffer := plaintext
		size := block.BlockSize()
		for len(ciphertext) > 0 {
			block.Decrypt(buffer, ciphertext)
			ciphertext = ciphertext[size:]
			buffer = buffer[size:]
		}
	}

	// strip padding
	plaintext = padding.strip(plaintext)
	return
}

type ErrModeUnsupported struct {
	mode Mode
}

func (e *ErrModeUnsupported) Error() string {
	return fmt.Sprintf("mode unsupported: %v", e.mode)
}

type ErrPaddingUnsupported struct {
	padding Padding
}

func (e *ErrPaddingUnsupported) Error() string {
	return fmt.Sprintf("padding unsupported: %v", e.padding)
}
