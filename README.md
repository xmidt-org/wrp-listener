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

Wrp-listener provides packages to help a consumer register to a webhook and 
authenticate messages received.  Registering to a webhook can be done directly 
or set up to run at an interval.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Details](#details)
- [Contributing](#contributing)

## Code of Conduct

This project and everyone participating in it are governed by the [XMiDT Code Of Conduct](https://xmidt.io/code_of_conduct/). 
By participating, you agree to this Code.

## Details

The below code snippet gets you registered to the webhook and events flowing to you.

```golang
	r := webhook.Registration{
		Config: webhook.DeliveryConfig{
			ReceiverURL: receiverURL,
			ContentType: contentType,
		},
		Events:   []string{events},
		Duration: webhook.CustomDuration(5 * time.Minute),
	}

	l, _ := listener.New(&r, "https://example.com",
		listener.AuthBearer(os.Getenv("WEBHOOK_BEARER_TOKEN")),
		listener.AcceptSHA1(),
		listener.Logger(logger),
		listener.Interval(1 * time.Minute),
		listener.AcceptedSecrets(sharedSecrets...),
    
    _ = l.Register(sharedSecrets[0])
```

Authorization is also pretty simple.

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

The example found in [cmd/bearerListener/main.go](https://github.com/xmidt-org/wrp-listener/blob/main/cmd/bearerListener/main.go) is a working command line example that shows how to use the library from end to end.

Additional examples can be found in the `example_test.go` file.

Functional tests are found in `functional_test.go`

## Contributing

Refer to [CONTRIBUTING.md](CONTRIBUTING.md).
