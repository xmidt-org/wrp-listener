package hashTokenFactory

import (
	"bytes"
	"context"
	"crypto/hmac"
	"encoding/hex"
	"errors"
	"hash"
	"io/ioutil"
	"net/http"

	"github.com/goph/emperror"

	"github.com/xmidt-org/bascule"
)

type HashTokenFactory struct {
	hashType     string
	newFunc      func() hash.Hash
	secretGetter SecretGetter
}

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

type SecretGetter interface {
	GetSecret() (string, error)
}

func (htf HashTokenFactory) ParseAndValidate(ctx context.Context, req *http.Request, _ bascule.Authorization, value string) (bascule.Token, error) {
	if req.Body == nil {
		return nil, codeError{http.StatusBadRequest, errors.New("Empty request body")}
	}

	msgBytes, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return nil, codeError{http.StatusBadRequest, emperror.Wrap(err, "Could not read request body")}
	}

	// Restore the io.ReadCloser to its original state
	req.Body = ioutil.NopCloser(bytes.NewBuffer(msgBytes))

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

	return bascule.NewToken(htf.hashType, value, bascule.Attributes{}), nil
}

func New(hashType string, newHashFunc func() hash.Hash, secretGetter SecretGetter) (HashTokenFactory, error) {
	if secretGetter == nil {
		return HashTokenFactory{}, errors.New("nil secretGetter")
	}
	return HashTokenFactory{hashType: hashType, newFunc: newHashFunc, secretGetter: secretGetter}, nil
}
