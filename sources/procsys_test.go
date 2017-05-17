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

func TestReadStatsFile(t *testing.T) {
	numParsedMetrics := 0
	testLNETStatsText := "0 16 0 1911487 1898918 0 0 498100008 543996712 0 0"
	expectedResults := []lustreStatsMetric{
		{"allocated", lnetAllocatedHelp, 0, "", ""},
		{"maximum", lnetMaximumHelp, 16, "", ""},
		{"errors", lnetErrorsHelp, 0, "", ""},
		{"send_count", lnetSendCountHelp, 1911487, "", ""},
		{"receive_count", lnetReceiveCountHelp, 1898918, "", ""},
		{"route_count", lnetRouteCountHelp, 0, "", ""},
		{"drop_count", lnetDropCountHelp, 0, "", ""},
		{"send_length", lnetSendLengthHelp, 498100008, "", ""},
		{"receive_length", lnetReceiveLengthHelp, 543996712, "", ""},
		{"route_length", lnetRouteLengthHelp, 0, "", ""},
		{"drop_length", lnetDropLengthHelp, 0, "", ""},
	}

	for _, result := range expectedResults {
		metric, err := parseSysStatsFile(result.help, result.title, testLNETStatsText)
		if err != nil {
			t.Fatal(err)
		}
		metricFound := compareStatsMetrics(expectedResults, metric)
		if metricFound != nil {
			t.Fatalf("Metric %s was not found", metric.title)
		}
		numParsedMetrics++
	}

	if l := len(expectedResults); l != numParsedMetrics {
		t.Fatalf("Retrieved an unexpected number of stats. Expected: %d, Got: %d", l, numParsedMetrics)
	}
}
