// (C) Copyright 2017 Hewlett Packard Enterprise Development LP
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sources

import (
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	Factories["procsys"] = newLustreProcSysSource
}

type lustreProcsysSource struct {
	lustreProcMetrics []lustreProcMetric
	basePath          string
}

func (s *lustreProcsysSource) generateLNETTemplates() error {
	metricMap := map[string][]lustreHelpStruct{
		"lnet": {
			{"catastrophe", "catastrophe_enabled", "Returns 1 if currently in catastrophe mode", s.gaugeMetric},
			{"console_backoff", "console_backoff_enabled", "Returns non-zero number if console_backoff is enabled", s.gaugeMetric},
			{"console_max_delay_centisecs", "console_max_delay_centiseconds", "Minimum time in centiseconds before the console logs a message", s.gaugeMetric},
			{"console_min_delay_centisecs", "console_min_delay_centiseconds", "Maximum time in centiseconds before the console logs a message", s.gaugeMetric},
			{"console_ratelimit", "console_ratelimit_enabled", "Returns 1 if the console message rate limiting is enabled", s.gaugeMetric},
			{"debug_mb", "debug_megabytes", "Maximum buffer size in megabytes for the LNET debug messages", s.gaugeMetric},
			{"fail_err", "fail_error_total", "Number of errors that have been thrown", s.counterMetric},
			{"fail_val", "fail_maximum", "Maximum number of times to fail", s.gaugeMetric},
			{"lnet_memused", "lnet_memory_used_bytes", "Number of bytes allocated by LNET", s.gaugeMetric},
			{"panic_on_lbug", "panic_on_lbug_enabled", "Returns 1 if panic_on_lbug is enabled", s.gaugeMetric},
			{"watchdog_ratelimit", "watchdog_ratelimit_enabled", "Returns 1 if the watchdog rate limiter is enabled", s.gaugeMetric},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			newMetric := newLustreProcMetric(item.filename, item.promName, "lnet", path, item.helpText, item.metricFunc)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
	return nil
}

func newLustreProcSysSource() (LustreSource, error) {
	var l lustreProcsysSource
	l.basePath = "/proc/sys"
	l.generateLNETTemplates()
	return &l, nil
}

func (s *lustreProcsysSource) Update(ch chan<- prometheus.Metric) (err error) {
	metricType := "single"

	for _, metric := range s.lustreProcMetrics {
		paths, err := filepath.Glob(filepath.Join(s.basePath, metric.path, metric.filename))
		if err != nil {
			return err
		}
		if paths == nil {
			continue
		}
		for _, path := range paths {
			err = s.parseFile(metric.source, metricType, path, metric.helpText, metric.promName, func(nodeType string, nodeName string, name string, helpText string, value uint64) {
				ch <- metric.metricFunc([]string{nodeType}, []string{nodeName}, name, helpText, value)
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *lustreProcsysSource) parseFile(nodeType string, metricType string, path string, helpText string, promName string, handler func(string, string, string, string, uint64)) (err error) {
	_, nodeName, err := parseFileElements(path, 0)
	if err != nil {
		return err
	}
	value, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	convertedValue, err := strconv.ParseUint(strings.TrimSpace(string(value)), 10, 64)
	if err != nil {
		return err
	}
	handler(nodeType, nodeName, promName, helpText, convertedValue)
	return nil
}

func (s *lustreProcsysSource) counterMetric(labels []string, labelValues []string, name string, helpText string, value uint64) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", name),
			helpText,
			labels,
			nil,
		),
		prometheus.CounterValue,
		float64(value),
		labelValues...,
	)
}

func (s *lustreProcsysSource) gaugeMetric(labels []string, labelValues []string, name string, helpText string, value uint64) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", name),
			helpText,
			labels,
			nil,
		),
		prometheus.GaugeValue,
		float64(value),
		labelValues...,
	)
}
