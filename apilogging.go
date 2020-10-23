// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package apilogging aids debugging of API interactions by providing a
// reusable logging mechanism.
package apilogging

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httputil"

	"google.golang.org/api/option"
	htransport "google.golang.org/api/transport/http"
)

// LoggerConfig governs the behavior of the api logging.
type LoggerConfig struct {
	// Scopes allows for the override of API scopes used by the underlying HTTP
	// api transport.  If not specified, uses the default cloud-platform scope.
	Scopes []string
	// Logging must contain an instance of a logger.
	Logger *log.Logger
	// LogRequest allows filtration based on the request body bytes.
	LogRequest func(b []byte) bool
	// LogResponse allows filtration based on the response body bytes.
	LogResponse func(b []byte) bool
}

var defaultScopes = []string{"https://www.googleapis.com/auth/cloud-platform"}

// NewLoggingClient gives you the thing.
func NewLoggingClient(ctx context.Context, cfg *LoggerConfig) (*http.Client, error) {
	if cfg == nil {
		return nil, errors.New("must supply a valid loggerconfig")
	}
	if cfg.Logger == nil {
		return nil, errors.New("LoggerConfig must contain a Logger")
	}
	scopes := defaultScopes
	if cfg != nil && cfg.Scopes != nil {
		scopes = cfg.Scopes
	}
	tr, err := htransport.NewTransport(ctx, http.DefaultTransport, option.WithScopes(scopes...))
	if err != nil {
		return nil, err
	}
	cfg.Logger.Print("Starting http logging client")
	return &http.Client{Transport: interceptor{
		rt:  tr,
		cfg: cfg,
	}}, nil

}

type interceptor struct {
	rt  http.RoundTripper
	cfg *LoggerConfig
}

func (i interceptor) RoundTrip(r *http.Request) (*http.Response, error) {

	dumpReq, err := httputil.DumpRequest(r, true)
	if err != nil {
		return nil, err
	}
	if i.cfg.LogRequest == nil || i.cfg.LogRequest(dumpReq) {
		i.cfg.Logger.Printf("REQUEST\n=====\n%s\n=====\n", dumpReq)
	}

	resp, err := i.rt.RoundTrip(r)
	if err != nil {
		return resp, err
	}

	dumpResp, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}
	if i.cfg.LogResponse == nil || i.cfg.LogResponse(dumpResp) {
		i.cfg.Logger.Printf("RESPONSE\n=====\n%s\n=====\n", dumpResp)
	}

	return resp, nil
}
