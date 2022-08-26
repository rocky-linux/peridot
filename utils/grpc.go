// Copyright (c) All respective contributors to the Peridot Project. All rights reserved.
// Copyright (c) 2021-2022 Rocky Enterprise Software Foundation, Inc. All rights reserved.
// Copyright (c) 2021-2022 Ctrl IQ, Inc. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
// this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
// this list of conditions and the following disclaimer in the documentation
// and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors
// may be used to endorse or promote products derived from this software without
// specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package utils

import (
	"context"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	_ "github.com/lib/pq"
)

type GrpcEndpointRegister func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error

// HeaderMatcher is the default header matcher for gRPC gateway
func HeaderMatcher(headerName string) (string, bool) {
	switch headerName {
	case
		"Authorization",
		"Cookie",
		// The following headers are tracing headers
		"Grpc-Metadata-X-Request-Id",
		"X-Request-Id",
		"X-B3-Traceid",
		"X-B3-Spanid",
		"X-B3-Parentspanid",
		"X-B3-Sampled",
		"X-B3-Flags",
		"X-Ot-Span-Context",
		"X-Cloud-Trace-Context",
		"Traceparent",
		"Grpc-Trace-Bin":
		return headerName, true
	}

	return headerName, false
}

// DefaultServeMuxOption is the default serve mux chain
func DefaultServeMuxOption() []runtime.ServeMuxOption {
	return []runtime.ServeMuxOption{
		runtime.WithOutgoingHeaderMatcher(func(header string) (string, bool) {
			switch header {
			case
				"location",
				"set-cookie":
				return header, true
			}

			return header, false
		}),
		runtime.WithForwardResponseOption(func(ctx context.Context, w http.ResponseWriter, msg proto.Message) error {
			if w.Header().Get("location") != "" {
				w.WriteHeader(302)
			}

			return nil
		}),
		runtime.WithIncomingHeaderMatcher(HeaderMatcher),
	}
}

type GRPCOptions struct {
	DialOptions       []grpc.DialOption
	MuxOptions        []runtime.ServeMuxOption
	ServerOptions     []grpc.ServerOption
	Interceptor       grpc.UnaryServerInterceptor
	ServerInterceptor grpc.StreamServerInterceptor
	DisableREST       bool
	DisableGRPC       bool
	Timeout           *time.Duration
}

type Register struct {
	Context  context.Context
	Mux      *runtime.ServeMux
	Router   chi.Router
	Endpoint string
	Options  []grpc.DialOption
	Server   *grpc.Server
}

type RegisterServer struct {
	Server *grpc.Server
}

type GRPCServerRes struct {
	Cancel    context.CancelFunc
	WaitGroup *sync.WaitGroup
}

type EmptyF func()

// NewGRPCServer initializes a new gRPC server with
// our defaults and other common actions
func NewGRPCServer(goptions *GRPCOptions, endpoint func(*Register), serve func(*RegisterServer)) *GRPCServerRes {
	options := goptions
	var defInterceptors []grpc.ServerOption
	if options == nil {
		options = &GRPCOptions{
			ServerOptions: defInterceptors,
		}
	}

	// get grpc port from viper
	grpcPort := viper.GetString("grpc.port")
	grpcEndpoint := ":" + grpcPort

	var lis net.Listener
	var err error
	if !options.DisableGRPC {
		// create new listener for grpc endpoint
		lis, err = net.Listen("tcp", grpcEndpoint)
		if err != nil {
			logrus.Fatalf("failed to listen: %v", err)
		}
	}

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	// use DialOptions if not nil
	if options.DialOptions != nil {
		opts = options.DialOptions
	}
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1000*1024*1024), grpc.MaxCallSendMsgSize(1000*1024*1024)))

	serverOpts := options.ServerOptions
	// If the server already declares a unary interceptor, let's chain
	// and make grpc_prometheus run first
	if options.Interceptor != nil {
		serverOpts = append(serverOpts, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_prometheus.UnaryServerInterceptor,
			options.Interceptor,
		)))
	} else {
		// Else, only declare prometheus interceptor
		serverOpts = append(serverOpts, grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor))
	}

	// If the server already declares a stream interceptor, let's chain
	// and make grpc_prometheus run first
	if options.ServerInterceptor != nil {
		serverOpts = append(serverOpts, grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_prometheus.StreamServerInterceptor,
			options.ServerInterceptor,
		)))
	} else {
		// Else, only declare prometheus interceptor
		serverOpts = append(serverOpts, grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor))
	}
	serv := grpc.NewServer(serverOpts...)

	// background context since this is the "main" app
	ctx, cancel := context.WithCancel(context.TODO())

	// new common router
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	timeout := 190 * time.Second
	if options.Timeout != nil {
		timeout = *options.Timeout
	}
	r.Use(middleware.Timeout(timeout))

	// combine options (default and additional)
	var muxOptions []runtime.ServeMuxOption
	muxOptions = append(muxOptions, DefaultServeMuxOption()...)
	if options.MuxOptions != nil {
		muxOptions = append(muxOptions, options.MuxOptions...)
	}

	mux := runtime.NewServeMux(muxOptions...)

	register := &Register{
		Context:  ctx,
		Endpoint: grpcEndpoint,
		Mux:      mux,
		Options:  opts,
		Router:   r,
		Server:   serv,
	}

	endpoint(register)

	r.Mount("/", mux)

	var wg sync.WaitGroup

	if !options.DisableREST {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			logrus.Infof("starting http server on port %s", viper.GetString("api.port"))

			err := http.ListenAndServe(":"+viper.GetString("api.port"), r)

			if err != nil {
				logrus.Fatalf("could not start server - %s", err)
			}
			wg.Done()
		}(&wg)
	}

	if !options.DisableGRPC {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			logrus.Infof("starting grpc server on port %s", viper.GetString("grpc.port"))
			registerServer := &RegisterServer{
				Server: serv,
			}

			if serve != nil {
				serve(registerServer)
			}
			grpc_prometheus.Register(serv)

			err = serv.Serve(lis)
			if err != nil {
				logrus.Fatalf("failed to serve: %v", err)
			}
			wg.Done()
		}(&wg)
	}

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		promMux := http.NewServeMux()
		promMux.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":7332", promMux)
		if err != nil {
			logrus.Fatalf("could not start prometheus server - %s", err)
		}
	}(&wg)

	return &GRPCServerRes{
		Cancel:    cancel,
		WaitGroup: &wg,
	}
}
