# wrp-listener

wrp-listener is a library that provides a webhook registerer and a validation 
function to be used for authentication.

[![Build Status](https://github.com/xmidt-org/wrp-listener/actions/workflows/ci.yml/badge.svg)](https://github.com/xmidt-org/wrp-listener/actions/workflows/ci.yml)
[![codecov.io](http://codecov.io/github/xmidt-org/wrp-listener/coverage.svg?branch=main)](http://codecov.io/github/xmidt-org/wrp-listener?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/xmidt-org/wrp-listener)](https://goreportcard.com/report/github.com/xmidt-org/wrp-listener)
[![Apache V2 License](http://img.shields.io/badge/license-Apache%20V2-blue.svg)](https://github.com/xmidt-org/wrp-listener/blob/main/LICENSE)
[![GitHub Release](https://img.shields.io/github/release/xmidt-org/wrp-listener.svg)](https://github.com/xmidt-org/wrp-listener/releases)
[![GoDoc](https://pkg.go.dev/badge/github.com/xmidt-org/wrp-listener)](https://pkg.go.dev/github.com/xmidt-org/wrp-listener)

## Summary

`wrp-listener`` provides a package to help a consumer register to a webhook and 
authenticate messages received.  Registering to a webhook can be done directly 
or set up to run at an interval.

## Details

The below code snippet gets you registered to the webhook and events flowing to you.

```golang
	l, err := listener.New("https://example.com",
		&webhook.Registration{
			Config: webhook.DeliveryConfig{
				ReceiverURL: receiverURL,
				ContentType: contentType,
			},
			Events:   []string{events},
			Duration: webhook.CustomDuration(5 * time.Minute),
		},		
		listener.AcceptSHA1(),
		listener.Interval(1 * time.Minute),
		listener.AcceptedSecrets(sharedSecrets...),
	)
	if err != nil {
		panic(err)
	}

    err = l.Register(sharedSecrets[0])
	if err != nil {
		panic(err)
	}
```

Authorization that the information from the webhook likstener provider is also
pretty simple.

```golang
func (el *eventListener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token, err := el.l.Tokenize(r)
    if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = el.l.Authorize(r, token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// ... do more stuff ...

	w.WriteHeader(http.StatusOK)
}
```

The full example found in [cmd/bearerListener/main.go](https://github.com/xmidt-org/wrp-listener/blob/main/cmd/bearerListener/main.go) is a working command line example that shows how to use the library from end to end.

Additional examples can be found in the `example_test.go` file.

Functional tests are found in `functional_test.go`

## Code of Conduct

This project and everyone participating in it are governed by the [XMiDT Code Of Conduct](https://xmidt.io/code_of_conduct/). 
By participating, you agree to this Code.

## Contributing

Refer to [CONTRIBUTING.md](CONTRIBUTING.md).
