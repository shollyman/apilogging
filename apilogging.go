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
	// CaptureFullRequest governs whether the body of the request is captured.
	CaptureFullRequest bool
	// CaptureFullResponse governs whether the body of the response is captured.
	CaptureFullResponse bool
	// LogRequest allows filtration based on the request body bytes.
	LogRequest func(b []byte) bool
	// LogResponse allows filtration based on the response body bytes.  Whether the
	// request was matched is also available.
	LogResponse func(b []byte, requestMatched bool) bool
}

var defaultScopes = []string{"https://www.googleapis.com/auth/cloud-platform"}

// NewLoggingHTTPClient provides an instrumented HTTP client, which can be used for constructing
// an appropriate service-specific API client.
func NewLoggingHTTPClient(ctx context.Context, cfg *LoggerConfig) (*http.Client, error) {
	scopes := defaultScopes
	if cfg != nil && cfg.Scopes != nil {
		scopes = cfg.Scopes
	}
	tr, err := htransport.NewTransport(ctx, http.DefaultTransport, option.WithScopes(scopes...))
	if err != nil {
		return nil, err
	}
	interceptor, err := NewInterceptingRoundTripper(cfg, tr)
	if err != nil {
		return nil, err
	}

	return &http.Client{Transport: interceptor}, nil
}

// NewInterceptingRoundTripper sets up a logging http.RoundTripper.
func NewInterceptingRoundTripper(cfg *LoggerConfig, wrapped http.RoundTripper) (http.RoundTripper, error) {
	if cfg == nil {
		return nil, errors.New("must supply a valid loggerconfig")
	}
	if cfg.Logger == nil {
		return nil, errors.New("LoggerConfig must contain a Logger")
	}
	return interceptor{
		rt:  wrapped,
		cfg: cfg,
	}, nil
}

// This library works by using the middleware pattern to wrap a "real" RoundTripper.
// Because that called roundtripper may further modify the request, it is possible
// that the logged Request is not accurate.
type interceptor struct {
	rt  http.RoundTripper
	cfg *LoggerConfig
}

func (i interceptor) RoundTrip(r *http.Request) (*http.Response, error) {

	// Capture and evaluate the outgoing request.
	dumpReq, err := httputil.DumpRequest(r, i.cfg.CaptureFullRequest)
	if err != nil {
		return nil, err
	}
	matchedReq := false
	if i.cfg.LogRequest == nil || i.cfg.LogRequest(dumpReq) {
		matchedReq = true
		i.cfg.Logger.Printf("REQUEST\n=====\n%s\n=====\n", dumpReq)
	}

	// Invoke the real roundtripper
	resp, err := i.rt.RoundTrip(r)
	if err != nil {
		return resp, err
	}

	// Now capture/evaluate the response.
	dumpResp, err := httputil.DumpResponse(resp, i.cfg.CaptureFullResponse)
	if err != nil {
		return nil, err
	}
	if i.cfg.LogResponse == nil || i.cfg.LogResponse(dumpResp, matchedReq) {
		i.cfg.Logger.Printf("RESPONSE\n=====\n%s\n=====\n", dumpResp)
	}
	return resp, nil
}
