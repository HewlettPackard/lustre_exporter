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

const (
	// Help text dedicated to the 'lnet' metrics
	lnetAllocatedHelp     string = "Number of messages currently allocated"
	lnetMaximumHelp       string = "Maximum number of outstanding messages"
	lnetErrorsHelp        string = "Total number of errors"
	lnetSendCountHelp     string = "Total number of messages that have been sent"
	lnetReceiveCountHelp  string = "Total number of messages that have been received"
	lnetRouteCountHelp    string = "Total number of messages that have been routed"
	lnetDropCountHelp     string = "Total number of messages that have been dropped"
	lnetSendLengthHelp    string = "Total number of bytes sent"
	lnetReceiveLengthHelp string = "Total number of bytes received"
	lnetRouteLengthHelp   string = "Total number of bytes for routed messages"
	lnetDropLengthHelp    string = "Total number of bytes that have been dropped"
	//repeated strings replaced by constants
	single string = "single"
	stats  string = "stats"
)

// LnetEnabled specified whether LNET metrics should be collected
var LnetEnabled string

func init() {
	Factories["procsys"] = newLustreProcSysSource
}

type lustreProcsysSource struct {
	lustreProcMetrics []lustreProcMetric
	basePath          string
}

func (s *lustreProcsysSource) generateLNETTemplates(filter string) {
	metricMap := map[string][]lustreHelpStruct{
		"lnet": {
			{"catastrophe", "catastrophe_enabled", "Returns 1 if currently in catastrophe mode", s.gaugeMetric, false, extended},
			{"console_backoff", "console_backoff_enabled", "Returns non-zero number if console_backoff is enabled", s.gaugeMetric, false, extended},
			{"console_max_delay_centisecs", "console_max_delay_centiseconds", "Minimum time in centiseconds before the console logs a message", s.gaugeMetric, false, extended},
			{"console_min_delay_centisecs", "console_min_delay_centiseconds", "Maximum time in centiseconds before the console logs a message", s.gaugeMetric, false, extended},
			{"console_ratelimit", "console_ratelimit_enabled", "Returns 1 if the console message rate limiting is enabled", s.gaugeMetric, false, extended},
			{"debug_mb", "debug_megabytes", "Maximum buffer size in megabytes for the LNET debug messages", s.gaugeMetric, false, extended},
			{"fail_err", "fail_error_total", "Number of errors that have been thrown", s.counterMetric, false, core},
			{"fail_val", "fail_maximum", "Maximum number of times to fail", s.gaugeMetric, false, core},
			{"lnet_memused", "lnet_memory_used_bytes", "Number of bytes allocated by LNET", s.gaugeMetric, false, core},
			{"panic_on_lbug", "panic_on_lbug_enabled", "Returns 1 if panic_on_lbug is enabled", s.gaugeMetric, false, extended},
			{"stats", "allocated", lnetAllocatedHelp, s.gaugeMetric, false, core},
			{"stats", "maximum", lnetMaximumHelp, s.gaugeMetric, false, core},
			{"stats", "errors_total", lnetErrorsHelp, s.counterMetric, false, core},
			{"stats", "send_count_total", lnetSendCountHelp, s.counterMetric, false, core},
			{"stats", "receive_count_total", lnetReceiveCountHelp, s.counterMetric, false, core},
			{"stats", "route_count_total", lnetRouteCountHelp, s.counterMetric, false, core},
			{"stats", "drop_count_total", lnetDropCountHelp, s.counterMetric, false, core},
			{"stats", "send_bytes_total", lnetSendLengthHelp, s.counterMetric, false, core},
			{"stats", "receive_bytes_total", lnetReceiveLengthHelp, s.counterMetric, false, core},
			{"stats", "route_bytes_total", lnetRouteLengthHelp, s.counterMetric, false, core},
			{"stats", "drop_bytes_total", lnetDropLengthHelp, s.counterMetric, false, core},
			{"watchdog_ratelimit", "watchdog_ratelimit_enabled", "Returns 1 if the watchdog rate limiter is enabled", s.gaugeMetric, false, extended},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			if filter == extended || item.priorityLevel == core {
				newMetric := newLustreProcMetric(item.filename, item.promName, "lnet", path, item.helpText, item.hasMultipleVals, item.metricFunc)
				s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
			}
		}
	}
}

func newLustreProcSysSource() LustreSource {
	var l lustreProcsysSource
	l.basePath = filepath.Join(ProcLocation, "sys")
	if LnetEnabled != disabled {
		l.generateLNETTemplates(LnetEnabled)
	}
	return &l
}

func (s *lustreProcsysSource) Update(ch chan<- prometheus.Metric) (err error) {
	var metricType string

	for _, metric := range s.lustreProcMetrics {
		paths, err := filepath.Glob(filepath.Join(s.basePath, metric.path, metric.filename))
		if err != nil {
			return err
		}
		if paths == nil {
			continue
		}
		for _, path := range paths {
			metricType = single
			if metric.filename == stats {
				metricType = stats
			}
			err = s.parseFile(metric.source, metricType, path, metric.helpText, metric.promName, func(nodeType string, nodeName string, name string, helpText string, value float64) {
				ch <- metric.metricFunc([]string{"component", "target"}, []string{nodeType, nodeName}, name, helpText, value)
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func parseSysStatsFile(helpText string, promName string, statsFile string) (metric lustreStatsMetric, err error) {
	// statsMap contains the index mapping for the provided statistic
	statsMap := map[string]int{
		lnetAllocatedHelp:     0,
		lnetMaximumHelp:       1,
		lnetErrorsHelp:        2,
		lnetSendCountHelp:     3,
		lnetReceiveCountHelp:  4,
		lnetRouteCountHelp:    5,
		lnetDropCountHelp:     6,
		lnetSendLengthHelp:    7,
		lnetReceiveLengthHelp: 8,
		lnetRouteLengthHelp:   9,
		lnetDropLengthHelp:    10,
	}
	statsResults := regexCaptureNumbers(statsFile)
	if len(statsResults) < 1 {
		return metric, nil
	}
	index := statsMap[helpText]
	value, err := strconv.ParseFloat(statsResults[index], 64)
	if err != nil {
		return metric, err
	}
	metric = lustreStatsMetric{
		title: promName,
		help:  helpText,
		value: value,
	}
	return metric, nil
}

func (s *lustreProcsysSource) parseFile(nodeType string, metricType string, path string, helpText string, promName string, handler func(string, string, string, string, float64)) (err error) {
	_, nodeName, err := parseFileElements(path, 0)
	if err != nil {
		return err
	}
	switch metricType {
	case single:
		value, err := ioutil.ReadFile(filepath.Clean(path))
		if err != nil {
			return err
		}
		convertedValue, err := strconv.ParseFloat(strings.TrimSpace(string(value)), 64)
		if err != nil {
			return err
		}
		handler(nodeType, nodeName, promName, helpText, convertedValue)
	case stats:
		statsFileBytes, err := ioutil.ReadFile(filepath.Clean(path))
		if err != nil {
			return err
		}
		statsFile := string(statsFileBytes[:])
		metric, err := parseSysStatsFile(helpText, promName, statsFile)
		if err != nil {
			return err
		}
		handler(nodeType, nodeName, metric.title, helpText, metric.value)
	}
	return nil
}

func (s *lustreProcsysSource) counterMetric(labels []string, labelValues []string, name string, helpText string, value float64) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", name),
			helpText,
			labels,
			nil,
		),
		prometheus.CounterValue,
		value,
		labelValues...,
	)
}

func (s *lustreProcsysSource) gaugeMetric(labels []string, labelValues []string, name string, helpText string, value float64) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", name),
			helpText,
			labels,
			nil,
		),
		prometheus.GaugeValue,
		value,
		labelValues...,
	)
}
