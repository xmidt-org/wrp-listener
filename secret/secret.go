// SPDX-FileCopyrightText: 2019 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package secretGetter

type constantSecret struct {
	secret string
}

// GetSecret returns the secret given on initialization.
func (c *constantSecret) GetSecret() (string, error) {
	return c.secret, nil
}

// NewConstantSecret returns the secret it is initially given when GetSecret()
// is called.
func NewConstantSecret(secret string) *constantSecret {
	return &constantSecret{
		secret: secret,
	}
}
