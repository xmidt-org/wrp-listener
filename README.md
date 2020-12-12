# wrp-listener

wrp-listener is a library that provides a webhook registerer and a validation 
function to be used for authentication.

[![Build Status](https://github.com/xmidt-org/wrp-listener/workflows/CI/badge.svg)](https://github.com/xmidt-org/wrp-listener/actions)
[![codecov.io](http://codecov.io/github/xmidt-org/wrp-listener/coverage.svg?branch=main)](http://codecov.io/github/xmidt-org/wrp-listener?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/xmidt-org/wrp-listener)](https://goreportcard.com/report/github.com/xmidt-org/wrp-listener)
[![Apache V2 License](http://img.shields.io/badge/license-Apache%20V2-blue.svg)](https://github.com/xmidt-org/wrp-listener/blob/main/LICENSE)
[![GitHub release](https://img.shields.io/github/release/xmidt-org/wrp-listener.svg)](CHANGELOG.md)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/xmidt-org/wrp-listener)](https://pkg.go.dev/github.com/xmidt-org/wrp-listener)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=xmidt-org_wrp-listener&metric=alert_status)](https://sonarcloud.io/dashboard?id=xmidt-org_wrp-listener)

## Summary

Wrp-listener provides packages to help a consumer register to a webhook and 
authenticate messages received.  Registering to a webhook can be done directly 
or set up to run at an interval.  Message authentication is set up to work with 
the [bascule](https://github.com/xmidt-org/bascule) library.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Details](#details)
- [Contributing](#contributing)

## Code of Conduct

This project and everyone participating in it are governed by the [XMiDT Code Of Conduct](https://xmidt.io/code_of_conduct/). 
By participating, you agree to this Code.

## Details

### Authentication

Authentication is done using the bascule library: the token factory provided 
in the `hashTokenFactory` package can be given to the bascule Constructor 
middleware in order to verify that the hashed body given with a request is 
valid and created with the expected secret.

### Registering

Registration happens through the `webhookClient` package, and can be set up for 
manual registration or registration at an interval.  If the consumer of this 
package decides when to register, an error is returned if registering a webhook 
is not successful.  With registering at an interval, a logger can be provided.  
Then, if an error occurs, the registerer will log it and then try again at the 
next interval.

## Contributing

Refer to [CONTRIBUTING.md](CONTRIBUTING.md).
