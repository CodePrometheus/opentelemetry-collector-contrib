// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package kongreceiver

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configtls"
	"go.uber.org/zap"
)

func TestNewClient(t *testing.T) {
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

	testCase := []struct {
		desc        string
		cfg         *Config
		host        component.Host
		settings    component.TelemetrySettings
		logger      *zap.Logger
		expectError error
	}{
		{
			desc: "Invalid HTTP config",
			cfg: &Config{
				ClientConfig: clientConfigNoCA,
			},
			host:        componenttest.NewNopHost(),
			settings:    componenttest.NewNopTelemetrySettings(),
			logger:      zap.NewNop(),
			expectError: errors.New("failed to create HTTP Client"),
		},
		{
			desc: "Valid Configuration",
			cfg: &Config{
				ClientConfig: clientConfig,
			},
			host:        componenttest.NewNopHost(),
			settings:    componenttest.NewNopTelemetrySettings(),
			logger:      zap.NewNop(),
			expectError: nil,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.desc, func(t *testing.T) {
			ac, err := newClient(context.Background(), tc.cfg, tc.host, tc.settings, tc.logger)
			if tc.expectError != nil {
				require.Nil(t, ac)
				require.ErrorContains(t, err, tc.expectError.Error())
			} else {
				require.NoError(t, err)

				actualClient, ok := ac.(*kongClient)
				require.True(t, ok)

				require.Equal(t, tc.cfg.Endpoint, actualClient.hostEndpoint)
				require.Equal(t, tc.logger, actualClient.logger)
				require.NotNil(t, actualClient.client)
			}
		})
	}
}

func loadAPIResponseData(t *testing.T, folder, fileName string) []byte {
	t.Helper()
	fullPath := filepath.Join("testdata", folder, fileName)

	data, err := os.ReadFile(fullPath)
	require.NoError(t, err)

	return data
}
