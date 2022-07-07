# XML-RPC Client for Go

This is an implementation of client-side part of XML-RPC protocol in Go.

![GitHub Workflow Status](https://img.shields.io/github/workflow/status/alexejk/go-xmlrpc/Build)
[![codecov](https://codecov.io/gh/alexejk/go-xmlrpc/branch/master/graph/badge.svg)](https://codecov.io/gh/alexejk/go-xmlrpc)
[![Go Report Card](https://goreportcard.com/badge/alexejk.io/go-xmlrpc)](https://goreportcard.com/report/alexejk.io/go-xmlrpc)

[![GoDoc](https://godoc.org/alexejk.io/go-xmlrpc?status.svg)](https://godoc.org/alexejk.io/go-xmlrpc)
![GitHub](https://img.shields.io/github/license/alexejk/go-xmlrpc)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/alexejk/go-xmlrpc)


## Usage

Add dependency to your project:

```shell
go get -u alexejk.io/go-xmlrpc
```

Use it by creating an `*xmlrpc.Client` and firing RPC method calls with `Call()`.

```go
package main

import(
    "fmt"

    "alexejk.io/go-xmlrpc"
)

func main() {
    client, _ := xmlrpc.NewClient("https://bugzilla.mozilla.org/xmlrpc.cgi")

    result := &struct {
        BugzillaVersion struct {
            Version string
        }
    }{}

    _ = client.Call("Bugzilla.version", nil, result)
    fmt.Printf("Version: %s\n", result.BugzillaVersion.Version)
}
```

Customization is supported by passing a list of `Option` to the `NewClient` function. 
For instance:

 - To customize any aspect of `http.Client` used to perform requests, use `HttpClient` option, otherwise `http.DefaultClient` will be used
 - To pass custom headers, make use of `Headers` option.

### Argument encoding

Arguments to the remote RPC method are passed on as a `*struct`. This struct is encoded into XML-RPC types based on following rules:

* Order of fields in struct type matters - fields are taken in the order they are defined on the **type**.
* Numbers are to be specified as `int` (encoded as `<int>`) or `float64` (encoded as `<double>`)
* Both pointer and value references are accepted (pointers are followed to actual values)

### Response decoding

Response is decoded following similar rules to argument encoding.

* Order of fields is important.
* Outer struct should contain exported field for each response parameter.
* Structs may contain pointers - they will be initialized if required.

## Building

To build this project, simply run `make all`. 
If you prefer building in Docker instead - `make build-in-docker` is your friend.
