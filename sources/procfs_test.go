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

func TestGetJobNum(t *testing.T) {
	testString := "job_id: 1234"
	expected := "1234"

	jobID, err := getJobNum(testString)
	if err != nil {
		t.Fatal(err)
	}
	if jobID != expected {
		t.Fatalf("Retrieved an unexpected Job ID. Expected: %s, Got: %s", expected, jobID)
	}

	testString = "job_id: ABCD"
	expected = ""

	jobID, err = getJobNum(testString)
	if err != nil {
		t.Fatal(err)
	}
	if jobID != expected {
		t.Fatalf("Retrieved an unexpected Job ID. Expected: %s, Got: %s", expected, jobID)
	}
}

func TestGetJobStats(t *testing.T) {
	testJobBlock := `- job_id:          29
  snapshot_time:   1493326943
  read_bytes:      { samples:           0, unit: bytes, min:       0, max:       0, sum:               0 }
  write_bytes:     { samples:         262, unit: bytes, min: 1048576, max: 1048576, sum:       274726912 }
  getattr:         { samples:           0, unit:  reqs }
  setattr:         { samples:           0, unit:  reqs }
  punch:           { samples:           0, unit:  reqs }
  sync:            { samples:           0, unit:  reqs }
  destroy:         { samples:           0, unit:  reqs }
  create:          { samples:           0, unit:  reqs }
  statfs:          { samples:           0, unit:  reqs }
  get_info:        { samples:           0, unit:  reqs }
  set_info:        { samples:           0, unit:  reqs }
  quotactl:        { samples:           0, unit:  reqs }`
	testJobID := "29"
	testPromName := "job_write_bytes_total"
	testHelpText := writeTotalHelp
	expected := uint64(274726912)

	metricList, err := getJobStatsByOperation(testJobBlock, testJobID, testPromName, testHelpText)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(metricList); l != 1 {
		t.Fatalf("Retrieved an unexpected number of items. Expected: %d, Got: %d", 1, l)
	}
	if metricList[0].value != expected {
		t.Fatalf("Retrieved an unexpected value. Expected: %d, Got: %d", expected, metricList[0].value)
	}
	if metricList[0].help != writeTotalHelp {
		t.Fatal("Retrieved an unexpected help text.")
	}
	if metricList[0].title != testPromName {
		t.Fatalf("Retrieved an unexpected name. Expected: %s, Got: %s", testPromName, metricList[0].title)
	}

	testPromName = "dne"
	testHelpText = "Help for DNE"
	expected = 0

	metricList, err = getJobStatsByOperation(testJobBlock, testJobID, testPromName, testHelpText)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(metricList); l != 0 {
		t.Fatalf("Retrieved an unexpected number of items. Expected: %d, Got: %d", 0, l)
	}

	testJobBlock = "- job_id:           29"
	testPromName = "job_write_bytes_total"
	testHelpText = writeTotalHelp

	_, err = getJobStatsByOperation(testJobBlock, testJobID, testPromName, testHelpText)
	if err != nil {
		t.Fatal(err)
	}
}
