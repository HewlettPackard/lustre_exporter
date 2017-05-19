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
	"testing"
)

const testStatsText string = `snapshot_time             1495035142.697801 secs.usecs
create                    2 samples [reqs]
read_bytes                1262 samples [bytes] 1024 1048576 5672395063
write_bytes               432 samples [bytes] 1024 1048576 375920472
statfs                    81813 samples [reqs]
reconnect                 2 samples [reqs]
statfs                    126150 samples [reqs]
ping                      16390 samples [reqs]`

func TestStatsIntegration(t *testing.T) {
	numParsedMetrics := 0
	metricsToTest := map[string]string{
		"read_samples_total":       readSamplesHelp,
		"read_minimum_size_bytes":  readMinimumHelp,
		"read_maximum_size_bytes":  readMaximumHelp,
		"read_bytes_total":         readTotalHelp,
		"write_samples_total":      writeSamplesHelp,
		"write_minimum_size_bytes": writeMinimumHelp,
		"write_maximum_size_bytes": writeMaximumHelp,
		"write_bytes_total":        writeTotalHelp,
	}
	expectedMetrics := []lustreStatsMetric{
		{"read_samples_total", readSamplesHelp, 1262},
		{"read_minimum_size_bytes", readMinimumHelp, 1024},
		{"read_maximum_size_bytes", readMaximumHelp, 1048576},
		{"read_bytes_total", readTotalHelp, 5672395063},
		{"write_samples_total", writeSamplesHelp, 432},
		{"write_minimum_size_bytes", writeMinimumHelp, 1024},
		{"write_maximum_size_bytes", writeMaximumHelp, 1048576},
		{"write_bytes_total", writeTotalHelp, 375920472},
	}

	for promName, promHelp := range metricsToTest {
		parsedMetrics, err := parseReadWriteBytes(testStatsText, promHelp, promName)
		if err != nil {
			t.Fatal(err)
		}
		for _, metric := range parsedMetrics {
			numParsedMetrics++
			metricFound := compareStatsMetrics(expectedMetrics, metric)
			if metricFound != nil {
				t.Fatalf("Metric %s was not found", promName)
			}
		}
	}

	if l := len(expectedMetrics); l != numParsedMetrics {
		t.Fatalf("Retrieved an unexpected number of stats. Expected: %d, Got: %d", l, numParsedMetrics)
	}
}
