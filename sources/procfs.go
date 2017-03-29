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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type lustreProcMetric struct {
	subsystem string
	name      string
	source    string //The node type (OSS, MDS, MGS)
	path      string //Path to retreive metric from
	helpText  string
}

func init() {
	Factories["procfs"] = NewLustreSource
}

type lustreSource struct {
	lustreProcMetrics []lustreProcMetric
	basePath          string
}

func newLustreProcMetric(name string, source string, path string, helpText string) lustreProcMetric {
	var m lustreProcMetric
	m.name = name
	m.source = source
	m.path = path
	m.helpText = helpText

	return m
}

func (s *lustreSource) generateOSSMetricTemplates() error {
	metricMap := map[string]map[string]string{
		"obdfilter/*": map[string]string{
			"blocksize":            "Filesystem block size in bytes",
			"brw_size":             "Block read/write size in bytes",
			"degraded":             "Binary indicator as to whether or not the pool is degraded - 0 for not degraded, 1 for degraded",
			"filesfree":            "The number of inodes (objects) available",
			"filestotal":           "The maximum number of inodes (objects) the filesystem can hold",
			"grant_compat_disable": "Binary indicator as to whether clients with OBD_CONNECT_GRANT_PARAM setting will be granted space",
			"grant_precreate":      "Maximum space in bytes that clients can preallocate for objects",
			"job_cleanup_interval": "Interval in seconds between cleanup of tuning statistics",
			"kbytesavail":          "Number of kilobytes readily available in the pool",
			"kbytesfree":           "Number of kilobytes allocated to the pool",
			"kbytestotal":          "Capacity of the pool in kilobytes",
			"lfsck_speed_limit":    "Maximum operations per second LFSCK (Lustre filesystem verification) can run",
			"num_exports":          "Total number of times the pool has been exported",
			"precreate_batch":      "Maximum number of objects that can be included in a single transaction",
			"recovery_time_hard":   "Maximum timeout 'recover_time_soft' can increment to for a single server",
			"recovery_time_soft":   "Duration in seconds for a client to attempt to reconnect after a crash (automatically incremented if servers are still in an error state)",
			"soft_sync_limit":      "Number of RPCs necessary before triggering a sync",
			"sync_journal":         "Binary indicator as to whether or not the journal is set for asynchronous commits",
			"tot_dirty":            "Total number of exports that have been marked dirty",
			"tot_granted":          "Total number of exports that have been marked granted",
			"tot_pending":          "Total number of exports that have been marked pending",
		},
	}
	for path, _ := range metricMap {
		for metric, helpText := range metricMap[path] {
			newMetric := newLustreProcMetric(metric, "OSS", path, helpText)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
	return nil
}

func (s *lustreSource) generateMGSMetricTemplates() error {
	metricMap := map[string]map[string]string{
		"mgs/MGS/osd/": map[string]string{
			"blocksize":            "Filesystem block size in bytes",
			"filesfree":            "The number of inodes (objects) available",
			"filestotal":           "The maximum number of inodes (objects) the filesystem can hold",
			"kbytesavail":          "Number of kilobytes readily available in the pool",
			"kbytesfree":           "Number of kilobytes allocated to the pool",
			"kbytestotal":          "Capacity of the pool in kilobytes",
			"quota_iused_estimate": "Returns '1' if a valid address is returned within the pool, referencing whether free space can be allocated",
		},
	}
	for path, _ := range metricMap {
		for metric, helpText := range metricMap[path] {
			newMetric := newLustreProcMetric(metric, "MGS", path, helpText)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
	return nil
}

func NewLustreSource() (LustreSource, error) {
	var l lustreSource
	l.basePath = "/proc/fs/lustre"
	//control which node metrics you pull via flags
	l.generateOSSMetricTemplates()
	l.generateMGSMetricTemplates()
	return &l, nil
}

func (s *lustreSource) Update(ch chan<- prometheus.Metric) (err error) {
	for _, metric := range s.lustreProcMetrics {
		paths, err := filepath.Glob(filepath.Join(s.basePath, metric.path, metric.name))
		if err != nil {
			return err
		}
		if paths == nil {
			continue
		}
		for _, path := range paths {

			err = s.parseFile(metric.source, "single", path, metric.helpText, func(nodeType string, nodeName string, name string, helpText string, value uint64) {
				ch <- s.constMetric(nodeType, nodeName, name, helpText, value)
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *lustreSource) parseFile(nodeType string, metricType string, path string, helpText string, handler func(string, string, string, string, uint64)) (err error) {
	pathElements := strings.Split(path, "/")
	pathLen := len(pathElements)
	if pathLen < 1 {
		return fmt.Errorf("path did not return at least one element")
	}
	name := pathElements[pathLen-1]
	nodeName := pathElements[pathLen-2]
	switch metricType {
	case "single":
		value, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		convertedValue, err := strconv.ParseUint(strings.TrimSpace(string(value)), 10, 64)
		if err != nil {
			return err
		}
		handler(nodeType, nodeName, name, helpText, convertedValue)
	}
	return nil
}

func (s *lustreSource) constMetric(nodeType string, nodeName string, name string, helpText string, value uint64) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "lustre", name),
			helpText,
			[]string{nodeType},
			nil,
		),
		prometheus.CounterValue,
		float64(value),
		nodeName,
	)
}
