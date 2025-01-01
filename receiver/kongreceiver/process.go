// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package kongreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kongreceiver"

import (
	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kongreceiver/internal/metadata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kongreceiver/internal/models"
)

func (k *kongScraper) processMetrics(props []*models.Properties, now pcommon.Timestamp) {
	for _, prop := range props {
		if prop.Name == "kong_bandwidth_bytes" {
			for _, metric := range prop.Metrics {
				switch metric.Labels["direction"] {
				case "egress":
					_ = k.mb.RecordKongBandwidthBytesDataPoint(now, metric.Value, metadata.AttributeDirectionEgress)
				case "ingress":
					_ = k.mb.RecordKongBandwidthBytesDataPoint(now, metric.Value, metadata.AttributeDirectionIngress)
				}
			}
		}
		if prop.Name == "kong_nginx_connections_total" {
			for _, metric := range prop.Metrics {
				switch metric.Labels["state"] {
				case "accepted":
					_ = k.mb.RecordKongNginxConnectionsTotalDataPoint(now, metric.Value, metadata.AttributeStateAccepted)
				case "active":
					_ = k.mb.RecordKongNginxConnectionsTotalDataPoint(now, metric.Value, metadata.AttributeStateActive)
				case "handled":
					_ = k.mb.RecordKongNginxConnectionsTotalDataPoint(now, metric.Value, metadata.AttributeStateHandled)
				case "reading":
					_ = k.mb.RecordKongNginxConnectionsTotalDataPoint(now, metric.Value, metadata.AttributeStateReading)
				case "total":
					_ = k.mb.RecordKongNginxConnectionsTotalDataPoint(now, metric.Value, metadata.AttributeStateTotal)
				case "waiting":
					_ = k.mb.RecordKongNginxConnectionsTotalDataPoint(now, metric.Value, metadata.AttributeStateWaiting)
				case "writing":
					_ = k.mb.RecordKongNginxConnectionsTotalDataPoint(now, metric.Value, metadata.AttributeStateWriting)
				}
			}
		}
	}
}
