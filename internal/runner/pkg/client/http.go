// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"crypto/tls"
	"net/http"
	"strings"
	"time"

	"gitea.dev/actionslib/ping/v1/pingv1connect"
	"gitea.dev/actionslib/runner/v1/runnerv1connect"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
)

func getHTTPClient(endpoint string, insecure bool) *http.Client {
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
	if strings.HasPrefix(endpoint, "https://") && insecure {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	return &http.Client{Transport: transport}
}

// New returns a new runner client using connectrpc v2.
func New(endpoint string, insecure bool, uuid, token string) *HTTPClient {
	baseURL := strings.TrimRight(endpoint, "/") + "/api/actions"
	httpClient := getHTTPClient(endpoint, insecure)

	authInterceptor := func(next connect.ClientFunc) connect.ClientFunc {
		return func(ctx context.Context, spec connect.Spec) (connect.ClientStream, error) {
			if info, ok := connect.CallInfoForClientContext(ctx); ok && (uuid != "" || token != "") {
				if uuid != "" {
					info.RequestHeader().Set("X-Runner-UUID", uuid)
				}
				if token != "" {
					info.RequestHeader().Set("X-Runner-Token", token)
				}
			}
			return next(ctx, spec)
		}
	}

	transport := connecthttp.NewTransport(httpClient, baseURL)
	client := connect.NewClient(transport, authInterceptor)

	return &HTTPClient{
		PingServiceClient:   pingv1connect.NewPingServiceClient(client),
		RunnerServiceClient: runnerv1connect.NewRunnerServiceClient(client),
		endpoint:            endpoint,
		insecure:            insecure,
	}
}

func (c *HTTPClient) Address() string {
	return c.endpoint
}

func (c *HTTPClient) Insecure() bool {
	return c.insecure
}

var _ Client = (*HTTPClient)(nil)

// An HTTPClient manages communication with the runner API.
type HTTPClient struct {
	pingv1connect.PingServiceClient
	runnerv1connect.RunnerServiceClient
	endpoint string
	insecure bool
}
