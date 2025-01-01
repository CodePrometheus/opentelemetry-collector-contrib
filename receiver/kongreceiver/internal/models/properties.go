// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package models // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kongreceiver/internal/models"

type Properties struct {
	Name    string       `json:"name"`
	Help    string       `json:"help"`
	Type    string       `json:"type"`
	Metrics []MetricData `json:"metrics"`
}

type MetricData struct {
	Labels map[string]string `json:"labels"`
	Value  string            `json:"value"`
}
