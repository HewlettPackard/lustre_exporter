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
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	numRegexPattern = regexp.MustCompile(`[0-9]*\.[0-9]+|[0-9]+`)
)

type prometheusType func([]string, []string, string, string, uint64) prometheus.Metric

type lustreProcMetric struct {
	filename        string
	promName        string
	source          string //The parent data source (OST, MDS, MGS, etc)
	path            string //Path to retrieve metric from
	helpText        string
	hasMultipleVals bool
	metricFunc      prometheusType
}

type lustreStatsMetric struct {
	title           string
	help            string
	value           uint64
	extraLabel      string
	extraLabelValue string
}

type lustreHelpStruct struct {
	filename        string
	promName        string // Name to be used in Prometheus
	helpText        string
	metricFunc      prometheusType
	hasMultipleVals bool
}

func newLustreProcMetric(filename string, promName string, source string, path string, helpText string, hasMultipleVals bool, metricFunc prometheusType) lustreProcMetric {
	var m lustreProcMetric
	m.filename = filename
	m.promName = promName
	m.source = source
	m.path = path
	m.helpText = helpText
	m.hasMultipleVals = hasMultipleVals
	m.metricFunc = metricFunc

	return m
}

func regexCaptureString(pattern string, textToMatch string) (matchedString string) {
	// Return the first string in a list of matched strings if found
	strings := regexCaptureStrings(pattern, textToMatch)
	if len(strings) < 1 {
		return ""
	}
	return strings[0]
}

func regexCaptureStrings(pattern string, textToMatch string) (matchedStrings []string) {
	re := regexp.MustCompile(pattern)
	matchedStrings = re.FindAllString(textToMatch, -1)
	return matchedStrings
}

func regexCaptureNumbers(textToMatch string) (matchedNumbers []string) {
	matchedNumbers = numRegexPattern.FindAllString(textToMatch, -1)
	return matchedNumbers
}

func parseFileElements(path string, directoryDepth int) (name string, nodeName string, err error) {
	pathElements := strings.Split(path, "/")
	pathLen := len(pathElements)
	if pathLen < 1 {
		return "", "", fmt.Errorf("path did not return at least one element")
	}
	name = pathElements[pathLen-1]
	nodeName = pathElements[pathLen-2-directoryDepth]
	nodeName = strings.TrimPrefix(nodeName, "filter-")
	nodeName = strings.TrimSuffix(nodeName, "_UUID")
	return name, nodeName, nil
}

func convertToBytes(s string) string {
	if len(s) < 1 {
		return s
	}
	numericS := ""
	uppercaseS := strings.ToUpper(s)
	multiplier := float64(1)
	switch finalChar := uppercaseS[len(uppercaseS)-1:]; finalChar {
	case "K":
		numericS = strings.TrimSuffix(uppercaseS, "K")
		multiplier = math.Pow(2, 10)
	case "M":
		numericS = strings.TrimSuffix(uppercaseS, "M")
		multiplier = math.Pow(2, 20)
	case "G":
		numericS = strings.TrimSuffix(uppercaseS, "G")
		multiplier = math.Pow(2, 30)
	default:
		//passthrough integers and IO sizes that we don't expect
		return s
	}
	byteVal, err := strconv.ParseUint(numericS, 10, 64)
	if err != nil {
		return s
	}
	byteVal *= uint64(multiplier)
	return strconv.FormatUint(byteVal, 10)
}
