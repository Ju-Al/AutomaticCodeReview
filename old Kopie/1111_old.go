// Copyright 2018 The Go Cloud Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package driver defines interfaces to be implemented for providers of the
// secrets package.
package driver // import "gocloud.dev/internal/secrets/driver"

import "context"

// Crypter holds the key information to encrypt a plain text message into a
// cipher message, as well as decrypt a cipher message into a plain text message
// encrypted by the same key.
type Crypter interface {

	// Decrypt decrypts the ciphertext and returns the plaintext or an error.
	Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error)

	// Encrypt encrypts the plaintext and returns the cipher message.
	Encrypt(ctx context.Context, plaintext []byte) ([]byte, error)
}
// CrypterOptions controls Crypter behaviors.
// It is provided for future extensibility.
type CrypterOptions struct{}
