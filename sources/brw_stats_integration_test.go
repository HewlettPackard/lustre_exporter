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

const (
	testPagesPerBulkRW string = `pages per bulk r/w     rpcs  %   cum % |  rpcs % cum %
1:                     14    53  53    |  0    0 0
2:                     12    46  100   |  0    0 0`
	testDiscontiguousPages string = `discontiguous pages    rpcs  %   cum % |  rpcs  % cum %
0:                     26    100 100   |  0     0 0
1:                      0      0 100   |  0     0 0`
	testDiskIOsInFlight string = `disk I/Os in flight    ios   %   cum % |  ios % cum %
1:                     26    100 100   |  0   0 0
2:                      0      0 100   |  0   0 0`
	testIOTime string = `I/O time (1/1000s)     ios   %  cum %  |  ios % cum %
1:                     2     50  50    |  0   0 0
2:                     0      0  50    |  0   0 0
4:                     0      0  50    |  0   0 0
8:                     0      0  50    |  0   0 0
16:                    0      0  50    |  0   0 0
32:                    2     50 100    |  0   0 0`
	testDiskIOSize string = `disk I/O size          ios  %  cum %   |  ios % cum %
8:                      4   15  15     |  0   0   0
16:                     0    0  15     |  0   0   0
32:                     1    3  19     |  0   0   0
64:                     1    3  23     |  0   0   0
128:                    2    7  30     |  0   0   0
256:                    1    3  34     |  0   0   0
512:                    1    3  38     |  0   0   0
1K:                     2    7  46     |  0   0   0
2K:                     0    0  46     |  0   0   0
4K:                     0    0  46     |  0   0   0
8K:                    14   53 100     |  0   0   0`
)

func compareBRWMetrics(expectedMetrics []lustreBRWMetric, parsedMetric lustreBRWMetric) error {
	for _, metric := range expectedMetrics {
		if reflect.DeepEqual(metric, parsedMetric) {
			return nil
		}
	}
	return errors.New("Metric not found")
}

func TestBRWStatsIntegration(t *testing.T) {
	numExpectedMetrics := 0
	numParsedMetrics := 0
	brwSections := map[string]string{
		"pages per bulk r/w":  testPagesPerBulkRW,
		"discontiguous pages": testDiscontiguousPages,
		"disk I/Os in flight": testDiskIOsInFlight,
		"I/O time (1/1000s)":  testIOTime,
		"disk I/O size":       testDiskIOSize,
	}
	expectedMetrics := map[string][]lustreBRWMetric{
		"pages per bulk r/w": {
			{"1", "read", "14"},
			{"2", "read", "12"},
			{"1", "write", "0"},
			{"2", "write", "0"},
		},
		"discontiguous pages": {
			{"0", "read", "26"},
			{"1", "read", "0"},
			{"0", "write", "0"},
			{"1", "write", "0"},
		},
		"disk I/Os in flight": {
			{"1", "read", "26"},
			{"2", "read", "0"},
			{"1", "write", "0"},
			{"2", "write", "0"},
		},
		"I/O time (1/1000s)": {
			{"1", "read", "2"},
			{"2", "read", "0"},
			{"4", "read", "0"},
			{"8", "read", "0"},
			{"16", "read", "0"},
			{"32", "read", "2"},
			{"1", "write", "0"},
			{"2", "write", "0"},
			{"4", "write", "0"},
			{"8", "write", "0"},
			{"16", "write", "0"},
			{"32", "write", "0"},
		},
		"disk I/O size": {
			{"8", "read", "4"},
			{"16", "read", "0"},
			{"32", "read", "1"},
			{"64", "read", "1"},
			{"128", "read", "2"},
			{"256", "read", "1"},
			{"512", "read", "1"},
			{"1K", "read", "2"},
			{"2K", "read", "0"},
			{"4K", "read", "0"},
			{"8K", "read", "14"},
			{"8", "write", "0"},
			{"16", "write", "0"},
			{"32", "write", "0"},
			{"64", "write", "0"},
			{"128", "write", "0"},
			{"256", "write", "0"},
			{"512", "write", "0"},
			{"1K", "write", "0"},
			{"2K", "write", "0"},
			{"4K", "write", "0"},
			{"8K", "write", "0"},
		},
	}

	for sectionTitle, sectionText := range brwSections {
		numExpectedMetrics += len(expectedMetrics[sectionTitle])
		metricList, err := splitBRWStats(sectionText)
		if err != nil {
			t.Fatal(err)
		}
		for _, metric := range metricList {
			numParsedMetrics++
			metricFound := compareBRWMetrics(expectedMetrics[sectionTitle], metric)
			if metricFound != nil {
				t.Fatalf("Unexpected result for %s", sectionTitle)
			}
		}
	}

	if numExpectedMetrics != numParsedMetrics {
		t.Fatalf("Retrieved an unexpected number of metrics. Expected: %d, Got: %d", numExpectedMetrics, numParsedMetrics)
	}
}
