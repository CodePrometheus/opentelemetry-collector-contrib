// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package kongreceiver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/prom2json"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kongreceiver/internal/mocks"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kongreceiver/internal/models"
)

func TestScraperStart(t *testing.T) {
	clientConfigNoCA := confighttp.NewDefaultClientConfig()
	clientConfigNoCA.Endpoint = defaultEndpoint
	clientConfigNoCA.TLSSetting = configtls.ClientConfig{
		Config: configtls.Config{
			CAFile: "/non/existent",
		},
	}

	clientConfig := confighttp.NewDefaultClientConfig()
	clientConfig.TLSSetting = configtls.ClientConfig{}
	clientConfig.Endpoint = defaultEndpoint

	testcases := []struct {
		desc        string
		scraper     *kongScraper
		expectError bool
	}{
		{
			desc: "Bad Config",
			scraper: &kongScraper{
				cfg: &Config{
					ClientConfig: clientConfigNoCA,
				},
				settings: componenttest.NewNopTelemetrySettings(),
			},
			expectError: true,
		},
		{
			desc: "Valid Config",
			scraper: &kongScraper{
				cfg: &Config{
					ClientConfig: clientConfig,
				},
				settings: componenttest.NewNopTelemetrySettings(),
			},
			expectError: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.scraper.start(context.Background(), componenttest.NewNopHost())
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestScraperScrape(t *testing.T) {
	testCases := []struct {
		desc               string
		setupMockClient    func(t *testing.T) client
		expectedMetricFile string
		expectedErr        error
	}{
		{
			desc: "Nil client",
			setupMockClient: func(t *testing.T) client {
				return nil
			},
			expectedMetricFile: filepath.Join("testdata", "scraper", "no_metrics.yaml"),
			expectedErr:        errors.New("client not initialized"),
		},
		{
			desc: "successful scrape",
			setupMockClient: func(t *testing.T) client {
				mockClient := mocks.MockClient{}
				mockData := bytes.NewReader(loadAPIResponseData(t, "apiresponses", "kong-metric.txt"))
				mfChan := make(chan *dto.MetricFamily, 1024)
				err := prom2json.ParseReader(mockData, mfChan)
				require.NoError(t, err)
				var family []*prom2json.Family
				for mf := range mfChan {
					family = append(family, prom2json.NewFamily(mf))
				}
				result, err := json.Marshal(family)
				require.NoError(t, err)
				var props []*models.Properties
				err = json.Unmarshal(result, &props)
				require.NoError(t, err)
				mockClient.On("Metrics", mock.Anything).Return(props, nil)
				return &mockClient
			},
			expectedMetricFile: filepath.Join("testdata", "scraper", "expected.yaml"),
			expectedErr:        nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			scraper := newKongScraper(receivertest.NewNopSettings(), createDefaultConfig().(*Config))
			scraper.client = tc.setupMockClient(t)
			actualMetrics, err := scraper.scrape(context.Background())

			if tc.expectedErr == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr.Error())
			}

			expectedMetrics, err := golden.ReadMetrics(tc.expectedMetricFile)
			require.NoError(t, err)

			require.NoError(t, pmetrictest.CompareMetrics(expectedMetrics, actualMetrics,
				pmetrictest.IgnoreMetricDataPointsOrder(),
				pmetrictest.IgnoreResourceMetricsOrder(),
				pmetrictest.IgnoreStartTimestamp(),
				pmetrictest.IgnoreTimestamp()),
			)
		})
	}
}
