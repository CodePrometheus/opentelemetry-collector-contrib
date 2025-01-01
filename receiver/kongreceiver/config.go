// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package kongreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kongreceiver"

import (
	"fmt"
	"net/url"

	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/receiver/scraperhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kongreceiver/internal/metadata"
)

const defaultEndpoint = "http://localhost:8081/metrics"

// Config defines the configuration for the various elements of the receiver agent.
type Config struct {
	scraperhelper.ControllerConfig `mapstructure:",squash"`
	confighttp.ClientConfig        `mapstructure:",squash"`
	MetricsBuilderConfig           metadata.MetricsBuilderConfig `mapstructure:",squash"`
}

// Validate validates the configuration by checking for missing or invalid fields
func (cfg *Config) Validate() error {
	if _, err := url.Parse(cfg.Endpoint); err != nil {
		return fmt.Errorf("\"endpoint\" must be in the form of <scheme>://<hostname>:<port>: %w", err)
	}

	return nil
}
