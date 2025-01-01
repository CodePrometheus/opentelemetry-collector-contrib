// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package kongreceiver

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/collector/component"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/scraperinttest"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
)

const kongPort = "8001"

func TestIntegration(t *testing.T) {
	scraperinttest.NewIntegrationTest(
		NewFactory(),
		scraperinttest.WithContainerRequest(
			testcontainers.ContainerRequest{
				Image:        "kong:3.8.0",
				ExposedPorts: []string{kongPort},
				WaitingFor: wait.ForListeningPort(kongPort).
					WithStartupTimeout(2 * time.Minute),
				Env: map[string]string{
					"KONG_DATABASE":           "off",
					"KONG_ADMIN_LISTEN":       "0.0.0.0:8001, 0.0.0.0:8444 ssl",
					"KONG_DECLARATIVE_CONFIG": "/usr/local/kong/declarative/kong.yml",
				},
				Files: []testcontainers.ContainerFile{{
					HostFilePath:      filepath.Join("testdata", "integration", "kong.yml"),
					ContainerFilePath: "/usr/local/kong/declarative/kong.yml",
					FileMode:          700,
				}},
			}),
		scraperinttest.WithCustomConfig(
			func(t *testing.T, cfg component.Config, ci *scraperinttest.ContainerInfo) {
				rCfg := cfg.(*Config)
				rCfg.ControllerConfig.CollectionInterval = 100 * time.Millisecond
				rCfg.Endpoint = fmt.Sprintf("http://%s:%s/metrics", ci.Host(t), ci.MappedPort(t, kongPort))
			}),
		scraperinttest.WithCompareOptions(
			pmetrictest.IgnoreMetricValues(),
			pmetrictest.IgnoreStartTimestamp(),
			pmetrictest.IgnoreTimestamp(),
			pmetrictest.IgnoreMetricDataPointsOrder(),
		),
	).Run(t)
}
