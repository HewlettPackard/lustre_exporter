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
	// string mappings for 'health_check' values
	healthCheckHealthy   string = "1"
	healthCheckUnhealthy string = "0"
)

var (
	// HealthStatusEnabled specifies whether to collect Health metrics
	HealthStatusEnabled string
)

func init() {
	Factories["sysfs"] = newLustreSysSource
}

type lustreSysSource struct {
	lustreProcMetrics []lustreProcMetric
	basePath          string
}

func (s *lustreSysSource) generateHealthStatusTemplates(filter string) {
	metricMap := map[string][]lustreHelpStruct{
		"": {
			{"health_check", "health_check", "Current health status for the indicated instance: " + healthCheckHealthy + " refers to 'healthy', " + healthCheckUnhealthy + " refers to 'unhealthy'", s.gaugeMetric, false, core},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			if filter == extended || item.priorityLevel == core {
				newMetric := newLustreProcMetric(item.filename, item.promName, "health", path, item.helpText, item.hasMultipleVals, item.metricFunc)
				s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
			}
		}
	}
}

func newLustreSysSource() LustreSource {
	var l lustreSysSource
	l.basePath = filepath.Join(SysLocation, "fs/lustre")
	if HealthStatusEnabled != disabled {
		l.generateHealthStatusTemplates(HealthStatusEnabled)
	}
	return &l
}

func (s *lustreSysSource) Update(ch chan<- prometheus.Metric) (err error) {
	var directoryDepth int

	for _, metric := range s.lustreProcMetrics {
		directoryDepth = strings.Count(metric.filename, "/")
		paths, err := filepath.Glob(filepath.Join(s.basePath, metric.path, metric.filename))
		if err != nil {
			return err
		}
		if paths == nil {
			continue
		}
		for _, path := range paths {
			switch metric.filename {
			case "health_check":
				err = s.parseTextFile(metric.source, "health_check", path, directoryDepth, metric.helpText, metric.promName, func(nodeType string, nodeName string, name string, helpText string, value float64) {
					ch <- metric.metricFunc([]string{"component", "target"}, []string{nodeType, nodeName}, name, helpText, value)
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *lustreSysSource) parseTextFile(nodeType string, metricType string, path string, directoryDepth int, helpText string, promName string, handler func(string, string, string, string, float64)) (err error) {
	filename, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	fileString := string(fileBytes[:])
	switch filename {
	case "health_check":
		if strings.TrimSpace(fileString) == "healthy" {
			value, err := strconv.ParseFloat(strings.TrimSpace(healthCheckHealthy), 64)
			if err != nil {
				return err
			}
			handler(nodeType, nodeName, promName, helpText, value)
		} else {
			value, err := strconv.ParseFloat(strings.TrimSpace(healthCheckUnhealthy), 64)
			if err != nil {
				return err
			}
			handler(nodeType, nodeName, promName, helpText, value)
		}
	}
	return nil
}

func (s *lustreSysSource) gaugeMetric(labels []string, labelValues []string, name string, helpText string, value float64) prometheus.Metric {
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
