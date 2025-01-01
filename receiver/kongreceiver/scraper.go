// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package kongreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kongreceiver"

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/scraper/scrapererror"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kongreceiver/internal/metadata"
)

var metricsFailedFetch = "Failed to fetch kong metrics"

type kongScraper struct {
	client client

	settings component.TelemetrySettings
	cfg      *Config
	mb       *metadata.MetricsBuilder
}

func newKongScraper(
	settings receiver.Settings,
	cfg *Config,
) *kongScraper {
	return &kongScraper{
		settings: settings.TelemetrySettings,
		cfg:      cfg,
		mb:       metadata.NewMetricsBuilder(cfg.MetricsBuilderConfig, settings),
	}
}

func (k *kongScraper) start(ctx context.Context, host component.Host) error {
	httpClient, err := newClient(ctx, k.cfg, host, k.settings, k.settings.Logger)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	k.client = httpClient

	return nil
}

func (k *kongScraper) scrape(ctx context.Context) (pmetric.Metrics, error) {
	if k.client == nil {
		return pmetric.NewMetrics(), errors.New("client not initialized")
	}

	now := pcommon.NewTimestampFromTime(time.Now())
	var scraperErrors scrapererror.ScrapeErrors

	metrics, err := k.client.Metrics(ctx)
	if err != nil {
		k.settings.Logger.Error(metricsFailedFetch, zap.Error(err))
		scraperErrors.AddPartial(1, fmt.Errorf("%s %w", metricsFailedFetch, err))
	}
	k.processMetrics(metrics, now)

	return k.mb.Emit(), scraperErrors.Combine()
}
