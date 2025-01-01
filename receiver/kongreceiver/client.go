// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package kongreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kongreceiver"

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/prom2json"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kongreceiver/internal/models"
)

type client interface {
	Metrics(ctx context.Context) ([]*models.Properties, error)
}

type kongClient struct {
	client       *http.Client
	hostEndpoint string
	hostName     string
	logger       *zap.Logger
}

func newClient(ctx context.Context, cfg *Config, host component.Host, settings component.TelemetrySettings, logger *zap.Logger) (client, error) {
	httpClient, err := cfg.ToClient(ctx, host, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP Client: %w", err)
	}

	hostName, err := getHostname()
	if err != nil {
		return nil, err
	}
	return &kongClient{
		client:       httpClient,
		hostName:     hostName,
		hostEndpoint: cfg.Endpoint,
		logger:       logger,
	}, nil
}

func (k *kongClient) Metrics(ctx context.Context) ([]*models.Properties, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, k.hostEndpoint, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create get request for path %s: %w", k.hostEndpoint, err)
	}

	resp, err := k.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make http request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			k.logger.Warn("failed to close response body", zap.Error(closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		k.logger.Debug("Kong doGet non-200", zap.Error(err), zap.Int("status_code", resp.StatusCode))

		// Attempt to extract the error payload
		var payloadData []byte
		payloadData, err = io.ReadAll(resp.Body)
		if err != nil {
			k.logger.Debug("failed to read payload error message", zap.Error(err))
		} else {
			k.logger.Debug("Kong API Error", zap.ByteString("api_error", payloadData))
		}

		return nil, fmt.Errorf("non 200 code returned %d", resp.StatusCode)
	}

	mfChan := make(chan *dto.MetricFamily, 1024)
	if err = prom2json.ParseReader(resp.Body, mfChan); err != nil {
		k.logger.Debug("prom2json failed to parse metric", zap.Error(err))
		return nil, err
	}
	var family []*prom2json.Family
	for mf := range mfChan {
		family = append(family, prom2json.NewFamily(mf))
	}
	result, err := json.Marshal(family)
	if err != nil {
		return nil, err
	}

	var props []*models.Properties
	err = json.Unmarshal(result, &props)
	if err != nil {
		return nil, err
	}
	return props, nil
}

var osHostname = os.Hostname

func getHostname() (string, error) {
	host, err := osHostname()
	if err != nil {
		return "", err
	}
	return host, nil
}
