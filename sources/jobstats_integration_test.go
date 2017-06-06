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
	"errors"
	"reflect"
	"testing"
)

const testJobStats string = `- job_id:          29
  snapshot_time:   1493326943
  read_bytes:      { samples:         212, unit: bytes, min: 1048576, max: 1048576, sum:        15648015 }
  write_bytes:     { samples:         262, unit: bytes, min: 1048576, max: 1048576, sum:       274726912 }
  getattr:         { samples:          10, unit:  reqs }
  setattr:         { samples:           9, unit:  reqs }
  punch:           { samples:           1, unit:  reqs }
  sync:            { samples:           0, unit:  reqs }
  destroy:         { samples:           1, unit:  reqs }
  create:          { samples:          56, unit:  reqs }
  statfs:          { samples:           3, unit:  reqs }
  get_info:        { samples:          98, unit:  reqs }
  set_info:        { samples:          47, unit:  reqs }
  quotactl:        { samples:          10, unit:  reqs }
- job_id:          30
  snapshot_time:   1493326944
  read_bytes:      { samples:         113, unit: bytes, min: 1048576, max: 1048576, sum:       153153145 }
  write_bytes:     { samples:         179, unit: bytes, min: 1048576, max: 1048576, sum:      4534065056 }
  getattr:         { samples:          13, unit:  reqs }
  setattr:         { samples:           8, unit:  reqs }
  punch:           { samples:           0, unit:  reqs }
  sync:            { samples:           0, unit:  reqs }
  destroy:         { samples:           0, unit:  reqs }
  create:          { samples:           0, unit:  reqs }
  statfs:          { samples:           1, unit:  reqs }
  get_info:        { samples:         112, unit:  reqs }
  set_info:        { samples:           1, unit:  reqs }
  quotactl:        { samples:           1, unit:  reqs }
- job_id:          31
  snapshot_time:   1493326945
  read_bytes:      { samples:         289, unit: bytes, min: 1048576, max: 1048576, sum:       486460650 }
  write_bytes:     { samples:         890, unit: bytes, min: 1048576, max: 1048576, sum:     48904865312 }
  getattr:         { samples:           1, unit:  reqs }
  setattr:         { samples:           1, unit:  reqs }
  punch:           { samples:          10, unit:  reqs }
  sync:            { samples:           9, unit:  reqs }
  destroy:         { samples:          19, unit:  reqs }
  create:          { samples:          54, unit:  reqs }
  statfs:          { samples:          16, unit:  reqs }
  get_info:        { samples:          12, unit:  reqs }
  set_info:        { samples:          15, unit:  reqs }
  quotactl:        { samples:          11, unit:  reqs }`

func compareJobStatsMetrics(expectedMetrics []lustreJobsMetric, parsedMetric lustreJobsMetric) error {
	for _, metric := range expectedMetrics {
		if reflect.DeepEqual(metric, parsedMetric) {
			return nil
		}
	}
	return errors.New("Metric not found")
}

func TestJobStatsIntegration(t *testing.T) {
	numParsedMetrics := 0
	metricsToTest := map[string]string{
		"job_read_samples_total":       readSamplesHelp,
		"job_read_minimum_size_bytes":  readMinimumHelp,
		"job_read_maximum_size_bytes":  readMaximumHelp,
		"job_read_bytes_total":         readTotalHelp,
		"job_write_samples_total":      writeSamplesHelp,
		"job_write_minimum_size_bytes": writeMinimumHelp,
		"job_write_maximum_size_bytes": writeMaximumHelp,
		"job_write_bytes_total":        writeTotalHelp,
		"job_stats_total":              jobStatsHelp,
	}
	expectedMetrics := []lustreJobsMetric{
		{"29", lustreStatsMetric{"job_read_samples_total", readSamplesHelp, 212, "", ""}},
		{"29", lustreStatsMetric{"job_read_minimum_size_bytes", readMinimumHelp, 1048576, "", ""}},
		{"29", lustreStatsMetric{"job_read_maximum_size_bytes", readMaximumHelp, 1048576, "", ""}},
		{"29", lustreStatsMetric{"job_read_bytes_total", readTotalHelp, 15648015, "", ""}},
		{"29", lustreStatsMetric{"job_write_samples_total", writeSamplesHelp, 262, "", ""}},
		{"29", lustreStatsMetric{"job_write_minimum_size_bytes", writeMinimumHelp, 1048576, "", ""}},
		{"29", lustreStatsMetric{"job_write_maximum_size_bytes", writeMaximumHelp, 1048576, "", ""}},
		{"29", lustreStatsMetric{"job_write_bytes_total", writeTotalHelp, 274726912, "", ""}},
		{"29", lustreStatsMetric{"job_stats_total", jobStatsHelp, 10, "operation", "getattr"}},
		{"29", lustreStatsMetric{"job_stats_total", jobStatsHelp, 9, "operation", "setattr"}},
		{"29", lustreStatsMetric{"job_stats_total", jobStatsHelp, 1, "operation", "punch"}},
		{"29", lustreStatsMetric{"job_stats_total", jobStatsHelp, 0, "operation", "sync"}},
		{"29", lustreStatsMetric{"job_stats_total", jobStatsHelp, 1, "operation", "destroy"}},
		{"29", lustreStatsMetric{"job_stats_total", jobStatsHelp, 56, "operation", "create"}},
		{"29", lustreStatsMetric{"job_stats_total", jobStatsHelp, 3, "operation", "statfs"}},
		{"29", lustreStatsMetric{"job_stats_total", jobStatsHelp, 98, "operation", "get_info"}},
		{"29", lustreStatsMetric{"job_stats_total", jobStatsHelp, 47, "operation", "set_info"}},
		{"29", lustreStatsMetric{"job_stats_total", jobStatsHelp, 10, "operation", "quotactl"}},
		{"30", lustreStatsMetric{"job_read_samples_total", readSamplesHelp, 113, "", ""}},
		{"30", lustreStatsMetric{"job_read_minimum_size_bytes", readMinimumHelp, 1048576, "", ""}},
		{"30", lustreStatsMetric{"job_read_maximum_size_bytes", readMaximumHelp, 1048576, "", ""}},
		{"30", lustreStatsMetric{"job_read_bytes_total", readTotalHelp, 153153145, "", ""}},
		{"30", lustreStatsMetric{"job_write_samples_total", writeSamplesHelp, 179, "", ""}},
		{"30", lustreStatsMetric{"job_write_minimum_size_bytes", writeMinimumHelp, 1048576, "", ""}},
		{"30", lustreStatsMetric{"job_write_maximum_size_bytes", writeMaximumHelp, 1048576, "", ""}},
		{"30", lustreStatsMetric{"job_write_bytes_total", writeTotalHelp, 4534065056, "", ""}},
		{"30", lustreStatsMetric{"job_stats_total", jobStatsHelp, 13, "operation", "getattr"}},
		{"30", lustreStatsMetric{"job_stats_total", jobStatsHelp, 8, "operation", "setattr"}},
		{"30", lustreStatsMetric{"job_stats_total", jobStatsHelp, 0, "operation", "punch"}},
		{"30", lustreStatsMetric{"job_stats_total", jobStatsHelp, 0, "operation", "sync"}},
		{"30", lustreStatsMetric{"job_stats_total", jobStatsHelp, 0, "operation", "destroy"}},
		{"30", lustreStatsMetric{"job_stats_total", jobStatsHelp, 0, "operation", "create"}},
		{"30", lustreStatsMetric{"job_stats_total", jobStatsHelp, 1, "operation", "statfs"}},
		{"30", lustreStatsMetric{"job_stats_total", jobStatsHelp, 112, "operation", "get_info"}},
		{"30", lustreStatsMetric{"job_stats_total", jobStatsHelp, 1, "operation", "set_info"}},
		{"30", lustreStatsMetric{"job_stats_total", jobStatsHelp, 1, "operation", "quotactl"}},
		{"31", lustreStatsMetric{"job_read_samples_total", readSamplesHelp, 289, "", ""}},
		{"31", lustreStatsMetric{"job_read_minimum_size_bytes", readMinimumHelp, 1048576, "", ""}},
		{"31", lustreStatsMetric{"job_read_maximum_size_bytes", readMaximumHelp, 1048576, "", ""}},
		{"31", lustreStatsMetric{"job_read_bytes_total", readTotalHelp, 486460650, "", ""}},
		{"31", lustreStatsMetric{"job_write_samples_total", writeSamplesHelp, 890, "", ""}},
		{"31", lustreStatsMetric{"job_write_minimum_size_bytes", writeMinimumHelp, 1048576, "", ""}},
		{"31", lustreStatsMetric{"job_write_maximum_size_bytes", writeMaximumHelp, 1048576, "", ""}},
		{"31", lustreStatsMetric{"job_write_bytes_total", writeTotalHelp, 48904865312, "", ""}},
		{"31", lustreStatsMetric{"job_stats_total", jobStatsHelp, 1, "operation", "getattr"}},
		{"31", lustreStatsMetric{"job_stats_total", jobStatsHelp, 1, "operation", "setattr"}},
		{"31", lustreStatsMetric{"job_stats_total", jobStatsHelp, 10, "operation", "punch"}},
		{"31", lustreStatsMetric{"job_stats_total", jobStatsHelp, 9, "operation", "sync"}},
		{"31", lustreStatsMetric{"job_stats_total", jobStatsHelp, 19, "operation", "destroy"}},
		{"31", lustreStatsMetric{"job_stats_total", jobStatsHelp, 54, "operation", "create"}},
		{"31", lustreStatsMetric{"job_stats_total", jobStatsHelp, 16, "operation", "statfs"}},
		{"31", lustreStatsMetric{"job_stats_total", jobStatsHelp, 12, "operation", "get_info"}},
		{"31", lustreStatsMetric{"job_stats_total", jobStatsHelp, 15, "operation", "set_info"}},
		{"31", lustreStatsMetric{"job_stats_total", jobStatsHelp, 11, "operation", "quotactl"}},
	}

	for promName, promHelp := range metricsToTest {
		isJobStats := false
		if promName == "job_stats_total" {
			isJobStats = true
		}
		parsedMetrics, err := parseJobStatsText(testJobStats, promName, promHelp, isJobStats)
		if err != nil {
			t.Fatal(err)
		}
		for _, metric := range parsedMetrics {
			numParsedMetrics++
			metricFound := compareJobStatsMetrics(expectedMetrics, metric)
			if metricFound != nil {
				t.Fatalf("Metric %s was not found", promName)
			}
		}
	}
}
