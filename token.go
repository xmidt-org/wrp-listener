// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package listener

// Token represents the information needed to authenticate the flow of incoming
// webhook callbacks.
type Token interface {
	Type() string
	Principal() string
}

type token struct {
	alg       string
	principal string
}

// Type returns the type of hash to use for authentication.
func (t token) Type() string {
	return t.alg
}

// Principal returns the principal (calculated value included with the message)
// to use for authentication.
func (t token) Principal() string {
	return t.principal
}

// newToken creates a new token with the given hash type and principal.
func newToken(alg, principal string) *token {
	return &token{
		alg:       alg,
		principal: principal,
	}
}
