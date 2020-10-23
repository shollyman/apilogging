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

package apilogging

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"google.golang.org/api/oauth2/v1"
	"google.golang.org/api/option"
)

func TestBigQuery(t *testing.T) {

	projectID, ok := os.LookupEnv("GOOGLE_CLOUD_PROJECT")
	if !ok || len(projectID) == 0 {
		t.Fatalf("must set GOOGLE_CLOUD_PROJECT variable to a valid cloud project ID")
	}

	ctx := context.Background()

	// Setup a file logger.
	fname := "integration_log_bigquery.txt"
	f, err := os.Create(fname)
	if err != nil {
		t.Fatalf("os.Create: %v", err)
	}
	defer f.Close()
	l := log.New(f, "", log.LstdFlags)

	hc, err := NewLoggingHTTPClient(ctx, &LoggerConfig{
		Logger: l,
	})
	if err != nil {
		t.Fatalf("NewLoggingClient: %v", err)
	}
	client, err := bigquery.NewClient(ctx, projectID, option.WithHTTPClient(hc))
	if err != nil {
		t.Fatalf("bigquery.NewClient: %v", err)
	}

	// Iterate through datasets in the public dataset project, which should yield multiple
	// datasets and one or more API requests to bigquery.datasets.list
	it := client.DatasetsInProject(ctx, "bigquery-public-data")

	count := 0
	for {
		_, err := it.Next()

		if err == iterator.Done {
			break
		}

		if err != nil {
			t.Errorf("iteration error: %v", err)
			break
		}
		count++
	}

	if count == 0 {
		t.Error("expected to iterate multiple datasets, but got zero")
	}

	// close the API client
	client.Close()
	// close the logger file
	f.Close()
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		t.Fatalf("failed to read log contents: %v", err)
	}
	if len(b) == 0 {
		t.Error("expected non-empty logfile")
	}
	if !strings.Contains(string(b), "GET /bigquery/v2/projects/bigquery-public-data/datasets") {
		t.Error("Expected logfile to contain a datasets.list request and did not")
	}
	if !strings.Contains(string(b), "Content-Type: application/json") {
		t.Error("Expected logfile to contain a content-type response header and did not")
	}
}

func TestOAuthWithMatchers(t *testing.T) {
	ctx := context.Background()

	// Setup logger and intercepting transport
	fname := "integration_log_oauthv2.txt"
	f, err := os.Create(fname)
	if err != nil {
		t.Fatalf("os.Create: %v", err)
	}
	defer f.Close()
	l := log.New(f, "", log.LstdFlags)

	hc, err := NewLoggingHTTPClient(ctx, &LoggerConfig{
		Logger:      l,
		LogRequest:  func(b []byte) bool { return false },
		LogResponse: func(b []byte, reqMatched bool) bool { return !reqMatched },
	})
	if err != nil {
		t.Fatalf("NewLoggingClient: %v", err)
	}

	svc, err := oauth2.NewService(ctx, option.WithHTTPClient(hc))
	if err != nil {
		t.Fatalf("oauth2.NewService: %v", err)
	}

	// Interrogate the oauth2 userinfo info endpoint.
	call := svc.Userinfo.Get()
	resp, err := call.Do()
	if err != nil {
		t.Errorf("call Userinfo.Get: %v", err)
	}

	// We got a response back, but the identity is somehow empty.
	// Treat this as an error.
	if resp.Email == "" {
		t.Errorf("expected email to be populated, was empty")
	}

	// close the logger file
	f.Close()
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		t.Fatalf("failed to read log contents: %v", err)
	}
	if len(b) == 0 {
		t.Error("expected non-empty logfile")
	}
	if strings.Contains(string(b), "GET /oauth") {
		t.Error("Expected logfile to not contain requests, but did")
	}
	if !strings.Contains(string(b), "Content-Type: application/json") {
		t.Error("Expected logfile to contain a content-type response header and did not")
	}

}
