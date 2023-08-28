// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package listener

// Token represents the information needed to authenticate the flow of incoming
// webhook callbacks.
type Token struct {
	alg       string
	principal string
}

// Type returns the type of hash to use for authentication.
func (t Token) Type() string {
	return t.alg
}

// Principal returns the principal (calculated value included with the message)
// to use for authentication.
func (t Token) Principal() string {
	return t.principal
}

// NewToken creates a new token with the given hash type and principal.
func NewToken(alg, principal string) *Token {
	return &Token{
		alg:       alg,
		principal: principal,
	}
}
