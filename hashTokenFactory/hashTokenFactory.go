/**
 * Copyright 2019 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package hashTokenFactory

import (
	"bytes"
	"context"
	"crypto/hmac"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"net/http"

	"github.com/goph/emperror"

	"github.com/xmidt-org/bascule"
)

// H is the struct for the hashTokenFactory that can be used to validate that a
// hash given matches the hash of the request body using the secret from the
// SecretGetter.
type H struct {
	hashType     string
	newFunc      func() hash.Hash
	secretGetter SecretGetter
}

// codeError is an error that also returns the status code that should be given
// in the response.
type codeError struct {
	code int
	err  error
}

func (c codeError) Error() string {
	return c.err.Error()
}

func (c codeError) StatusCode() int {
	return c.code
}

// SecretGetter gets the secret to use when hashing.  If getting the secret is
// unsuccessful, an error can be returned.
type SecretGetter interface {
	GetSecret() (string, error)
}

// ParseAndValidate takes the hash given and validates that it matches the body
// hashed with the expected secret.
func (htf H) ParseAndValidate(ctx context.Context, req *http.Request, _ bascule.Authorization, value string) (bascule.Token, error) {
	if req.Body == nil {
		return nil, codeError{http.StatusBadRequest, errors.New("Empty request body")}
	}

	msgBytes, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return nil, codeError{http.StatusBadRequest, emperror.Wrap(err, "Could not read request body")}
	}

	// Restore the io.ReadCloser to its original state
	req.Body = io.NopCloser(bytes.NewBuffer(msgBytes))

	secretGiven, err := hex.DecodeString(value)
	if err != nil {
		return nil, codeError{http.StatusBadRequest, emperror.Wrap(err, "Could not decode signature")}
	}

	secret, err := htf.secretGetter.GetSecret()
	if err != nil {
		return nil, codeError{http.StatusInternalServerError, emperror.Wrap(err, "Could not get secret")}
	}
	h := hmac.New(htf.newFunc, []byte(secret))
	h.Write(msgBytes)
	sig := h.Sum(nil)
	if !hmac.Equal(sig, secretGiven) {
		return nil, codeError{http.StatusForbidden, emperror.With(errors.New("Invalid secret"), "secretGiven", secretGiven, "hashCalculated", sig, "body", msgBytes)}
	}

	return bascule.NewToken(htf.hashType, value, bascule.NewAttributes(map[string]interface{}{})), nil
}

// New returns the hash token factory to be used to validate a request.
func New(hashType string, newHashFunc func() hash.Hash, secretGetter SecretGetter) (H, error) {
	if secretGetter == nil {
		return H{}, errors.New("nil secretGetter")
	}
	return H{hashType: hashType, newFunc: newHashFunc, secretGetter: secretGetter}, nil
}
