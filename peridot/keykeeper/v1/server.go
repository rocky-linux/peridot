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

package keykeeperv1

import (
	"context"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"google.golang.org/grpc"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	commonpb "peridot.resf.org/common"
	peridotdb "peridot.resf.org/peridot/db"
	keykeeperpb "peridot.resf.org/peridot/keykeeper/pb"
	"peridot.resf.org/peridot/keykeeper/v1/store"
	"peridot.resf.org/peridot/keykeeper/v1/store/awssm"
	"peridot.resf.org/peridot/lookaside"
	"peridot.resf.org/peridot/lookaside/s3"
	"peridot.resf.org/utils"
	"strings"
	"sync"
	"time"
)

const TaskQueue = "keykeeper"

type MapStringLock struct {
	*sync.RWMutex
	m map[string]*sync.Mutex
}

func (m *MapStringLock) ReadLock(key string) {
	m.RLock()
	defer m.RUnlock()
	if m.m[key] == nil {
		m.Lock()
		m.m[key] = &sync.Mutex{}
		m.Unlock()
	}
	m.m[key].Lock()
}

func (m *MapStringLock) ReadUnlock(key string) {
	m.RLock()
	defer m.RUnlock()
	if m.m[key] == nil {
		m.Lock()
		m.m[key] = &sync.Mutex{}
		m.Unlock()
	}
	m.m[key].Unlock()
}

type Server struct {
	keykeeperpb.UnimplementedKeykeeperServiceServer

	log           *logrus.Logger
	db            peridotdb.Access
	storage       lookaside.Storage
	worker        worker.Worker
	temporal      client.Client
	stores        map[string]store.Store
	keys          map[string]*LoadedKey
	keyImportLock *MapStringLock
	defaultStore  string
}

func NewServer(db peridotdb.Access, c client.Client) (*Server, error) {
	storage, err := s3.New(osfs.New("/"))
	if err != nil {
		return nil, err
	}

	sm, err := awssm.New()
	if err != nil {
		return nil, err
	}

	return &Server{
		log:     logrus.New(),
		db:      db,
		storage: storage,
		worker: worker.New(c, TaskQueue, worker.Options{
			DeadlockDetectionTimeout: 15 * time.Minute,
		}),
		temporal: c,
		stores:   map[string]store.Store{"awssm": sm},
		keys:     map[string]*LoadedKey{},
		keyImportLock: &MapStringLock{
			RWMutex: &sync.RWMutex{},
			m:       map[string]*sync.Mutex{},
		},
		defaultStore: "awssm",
	}, nil
}

func (s *Server) interceptor(ctx context.Context, req interface{}, usi *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	n := utils.EndInterceptor

	return n(ctx, req, usi, handler)
}

func (s *Server) Run() {
	// Set timeout to 5 minutes
	// This is used for key generation
	// todo(mustafa): Evaluate if we should move key generation to Temporal
	timeout := 5 * time.Minute
	runtime.DefaultContextTimeout = timeout

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Register Temporal worker
	s.worker.RegisterWorkflow(s.SignArtifactsWorkflow)
	s.worker.RegisterActivity(s.SignArtifactActivity)
	defer s.temporal.Close()

	// Create keykeeper directories (/keykeeper/gpg and /keykeeper/artifacts)
	err := os.MkdirAll("/keykeeper/artifacts", 0755)
	if err != nil {
		s.log.Errorf("Failed to create keykeeper artifacts directory: %v", err)
	}
	err = os.MkdirAll("/keykeeper/gpg", 0755)
	if err != nil {
		s.log.Errorf("Failed to create keykeeper gpg directory: %v", err)
	}

	// Since we launch each server in a container, we're going to overwrite
	// some GPG options on each launch
	// todo(mustafa): Evaluate if this is the best way to do this (even though non-container workloads will never be supported)
	err = os.MkdirAll("/keykeeper/gpg/.gnupg", 0755)
	if err != nil {
		logrus.Fatalf("failed to create /keykeeper/gpg/.gnupg: %v", err)
	}
	err = ioutil.WriteFile("/keykeeper/gpg/.gnupg/gpg.conf", []byte("use-agent\npinentry-mode loopback"), 0644)
	if err != nil {
		logrus.Fatalf("could not create gpg config file: %v", err)
	}
	err = ioutil.WriteFile("/keykeeper/gpg/.gnupg/gpg-agent.conf", []byte("allow-loopback-pinentry"), 0644)
	if err != nil {
		logrus.Fatalf("could not create gpg agent config file: %v", err)
	}
	// Reload gpg-connect-agent
	agentReloadCmd := gpgCmdEnv(exec.Command("gpg-connect-agent"))
	agentReloadCmd.Stdin = strings.NewReader("RELOADAGENT\n")
	logs, err := logCmdRun(agentReloadCmd)
	if err != nil {
		logrus.Fatalf("could not reload gpg-connect-agent: %v\nlogs: %s", err, logs)
	}

	rpmMacros := `%__gpg_sign_cmd %{__gpg} \
    gpg --batch --no-verbose --no-armor --pinentry-mode loopback --passphrase %{_peridot_keykeeper_key} \
    %{?_gpg_digest_algo:--digest-algo %{_gpg_digest_algo}} \
    --no-secmem-warning \
    -u "%{_gpg_name}" -sbo %{__signature_filename} %{__plaintext_filename}`
	err = ioutil.WriteFile("/etc/rpm/macros.gpg", []byte(rpmMacros), 0644)
	if err != nil {
		logrus.Fatalf("could not create rpm macros file: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		err := s.worker.Run(worker.InterruptCh())
		if err != nil {
			logrus.Fatalf("could not run worker: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		res := utils.NewGRPCServer(
			&utils.GRPCOptions{
				Timeout: &timeout,
				ServerOptions: []grpc.ServerOption{
					grpc.UnaryInterceptor(s.interceptor),
				},
			},
			func(r *utils.Register) {
				endpoints := []utils.GrpcEndpointRegister{
					commonpb.RegisterHealthCheckServiceHandlerFromEndpoint,
					keykeeperpb.RegisterKeykeeperServiceHandlerFromEndpoint,
				}

				for _, endpoint := range endpoints {
					err := endpoint(r.Context, r.Mux, r.Endpoint, r.Options)
					if err != nil {
						s.log.Fatalf("could not register handler - %v", err)
					}
				}
			},
			func(r *utils.RegisterServer) {
				commonpb.RegisterHealthCheckServiceServer(r.Server, &utils.HealthServer{})

				keykeeperpb.RegisterKeykeeperServiceServer(r.Server, s)
			},
		)
		defer res.Cancel()
		res.WaitGroup.Wait()
	}()

	wg.Wait()
}
