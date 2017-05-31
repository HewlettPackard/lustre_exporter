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
	"strings"
	"testing"
)

func compareStatsMetrics(expectedMetrics []lustreStatsMetric, parsedMetric lustreStatsMetric) error {
	for _, metric := range expectedMetrics {
		if reflect.DeepEqual(metric, parsedMetric) {
			return nil
		}
	}
	return errors.New("Metric not found")
}

func TestRegexCaptureStrings(t *testing.T) {
	testString := `The lustre_exporter is a collector to be used with Prometheus which captures Lustre metrics.
Lustre is a parrallel filesystem for high-performance computers.
Currently, Lustre is on over 60% of the top supercomputers in the world.`
	// Matching is case-sensitive
	testPattern := "Lustre"
	expected := 3

	matchedStrings := regexCaptureStrings(testPattern, testString)
	if l := len(matchedStrings); l != expected {
		t.Fatalf("Retrieved an unexpected number of regex matches. Expected: %d, Got: %d", expected, l)
	}

	testPattern = "lustre"
	expected = 1

	matchedStrings = regexCaptureStrings(testPattern, testString)
	if l := len(matchedStrings); l != expected {
		t.Fatalf("Retrieved an unexpected number of regex matches. Expected: %d, Got: %d", expected, l)
	}

	// Matching is case-insensitive
	testPattern = "(?i)lustre"
	expected = 4

	matchedStrings = regexCaptureStrings(testPattern, testString)
	if l := len(matchedStrings); l != expected {
		t.Fatalf("Retrieved an unexpected number of regex matches. Expected: %d, Got: %d", expected, l)
	}

	// Match does not exist
	testPattern = "DNE"
	expected = 0

	matchedStrings = regexCaptureStrings(testPattern, testString)
	if l := len(matchedStrings); l != expected {
		t.Fatalf("Retrieved an unexpected number of regex matches. Expected: %d, Got: %d", expected, l)
	}
}

func TestRegexCaptureString(t *testing.T) {
	testString := "Hex Dump: 42 4F 49 4C 45 52 20 55 50"
	testPattern := "[0-9]*\\.[0-9]+|[0-9]+"
	expected := "42"

	matchedString := strings.TrimSpace(regexCaptureString(testPattern, testString))
	if matchedString != expected {
		t.Fatalf("Retrieved an unexpected string. Expected: %s, Got: %s", expected, matchedString)
	}

	testPattern = "DNE"
	expected = ""

	matchedString = strings.TrimSpace(regexCaptureString(testPattern, testString))
	if matchedString != expected {
		t.Fatalf("Retrieved an unexpected string. Expected: %s, Got: %s", expected, matchedString)
	}
}

func TestParseFileElements(t *testing.T) {
	testPath := "/proc/fs/lustre/obdfilter/OST0000/filesfree"
	directoryDepth := 0
	expectedName := "filesfree"
	expectedNodeName := "OST0000"

	name, nodeName, err := parseFileElements(testPath, directoryDepth)
	if err != nil {
		t.Fatal(err)
	}
	if name != expectedName {
		t.Fatalf("Retrieved an unexpected name. Expected: %s, Got: %s", expectedName, name)
	}
	if nodeName != expectedNodeName {
		t.Fatalf("Retrieved an unexpected name. Expected: %s, Got: %s", expectedNodeName, nodeName)
	}

	testPath = "/proc/fs/lustre/ldlm/namespaces/filter-lustrefs-OST0005_UUID/pool/grant_rate"
	directoryDepth = 1
	expectedName = "grant_rate"
	expectedNodeName = "lustrefs-OST0005"

	name, nodeName, err = parseFileElements(testPath, directoryDepth)
	if err != nil {
		t.Fatal(err)
	}
	if name != expectedName {
		t.Fatalf("Retrieved an unexpected name. Expected: %s, Got: %s", expectedName, name)
	}
	if nodeName != expectedNodeName {
		t.Fatalf("Retrieved an unexpected name. Expected: %s, Got: %s", expectedNodeName, nodeName)
	}

	testPath = "/proc/fs/lustre/health_check"
	directoryDepth = 0
	expectedName = "health_check"
	expectedNodeName = "lustre"

	name, nodeName, err = parseFileElements(testPath, directoryDepth)
	if err != nil {
		t.Fatal(err)
	}
	if name != expectedName {
		t.Fatalf("Retrieved an unexpected name. Expected: %s, Got: %s", expectedName, name)
	}
	if nodeName != expectedNodeName {
		t.Fatalf("Retrieved an unexpected name. Expected: %s, Got: %s", expectedNodeName, nodeName)
	}
}
