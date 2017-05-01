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
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Help text dedicated to the 'stats' file
	samplesHelp string = "Total number of times the given metric has been collected."
	maximumHelp string = "The maximum value retrieved for the given metric."
	minimumHelp string = "The minimum value retrieved for the given metric."
	totalHelp   string = "The sum of all values collected for the given metric."

	// Help text dedicated to the 'brw_stats' file
	pagesPerBlockRWHelp    string = "Total number of pages per RPC."
	discontiguousPagesHelp string = "Total number of logical discontinuities per RPC."
	ioTimeHelp             string = "Total time in milliseconds the filesystem has spent processing various object sizes."
	diskIOSizeHelp         string = "Total number of operations the filesystem has performed for the given size."
	diskIOsInFlightHelp    string = "Current number of I/O operations that are processing during the snapshot."
)

type lustreProcMetric struct {
	subsystem string
	name      string
	source    string //The parent data source (OST, MDS, MGS, etc)
	path      string //Path to retrieve metric from
	helpText  string
}

type lustreStatsMetric struct {
	title string
	help  string
	value uint64
}

type lustreJobsMetric struct {
	jobID     string
	operation string
	lustreStatsMetric
}

type lustreBRWMetric struct {
	size      string
	operation string
	value     string
}

func init() {
	Factories["procfs"] = newLustreSource
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

func (s *lustreSource) generateOSTMetricTemplates() error {
	metricMap := map[string]map[string]string{
		"obdfilter/*": {
			"blocksize":            "Filesystem block size in bytes",
			"brw_size":             "Block read/write size in bytes",
			"brw_stats":            "A collection of block read/write statistics",
			"degraded":             "Binary indicator as to whether or not the pool is degraded - 0 for not degraded, 1 for degraded",
			"filesfree":            "The number of inodes (objects) available",
			"filestotal":           "The maximum number of inodes (objects) the filesystem can hold",
			"grant_compat_disable": "Binary indicator as to whether clients with OBD_CONNECT_GRANT_PARAM setting will be granted space",
			"grant_precreate":      "Maximum space in bytes that clients can preallocate for objects",
			"job_cleanup_interval": "Interval in seconds between cleanup of tuning statistics",
			"job_stats":            "A collection of read/write statistics listed by jobid",
			"kbytesavail":          "Number of kilobytes readily available in the pool",
			"kbytesfree":           "Number of kilobytes allocated to the pool",
			"kbytestotal":          "Capacity of the pool in kilobytes",
			"lfsck_speed_limit":    "Maximum operations per second LFSCK (Lustre filesystem verification) can run",
			"num_exports":          "Total number of times the pool has been exported",
			"precreate_batch":      "Maximum number of objects that can be included in a single transaction",
			"recovery_time_hard":   "Maximum timeout 'recover_time_soft' can increment to for a single server",
			"recovery_time_soft":   "Duration in seconds for a client to attempt to reconnect after a crash (automatically incremented if servers are still in an error state)",
			"soft_sync_limit":      "Number of RPCs necessary before triggering a sync",
			"stats":                "A collection of statistics specific to Lustre",
			"sync_journal":         "Binary indicator as to whether or not the journal is set for asynchronous commits",
			"tot_dirty":            "Total number of exports that have been marked dirty",
			"tot_granted":          "Total number of exports that have been marked granted",
			"tot_pending":          "Total number of exports that have been marked pending",
		},
		"ldlm/namespaces/filter-*": {
			"lock_count":         "Number of locks",
			"lock_timeouts":      "Number of lock timeouts",
			"contended_locks":    "Number of contended locks",
			"contention_seconds": "Time in seconds during which locks were contended",
			"pool/granted":       "Number of granted locks",
			"pool/grant_rate":    "Lock grant rate",
			"pool/cancel_rate":   "Lock cancel rate",
			"pool/grant_speed":   "Lock grant speed",
		},
	}
	for path := range metricMap {
		for metric, helpText := range metricMap[path] {
			newMetric := newLustreProcMetric(metric, "OST", path, helpText)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
	return nil
}

func (s *lustreSource) generateMDTMetricTemplates() error {
	metricMap := map[string]map[string]string{
		"mdt/*": {
			"num_exports": "Total number of times the pool has been exported",
		},
	}
	for path := range metricMap {
		for metric, helpText := range metricMap[path] {
			newMetric := newLustreProcMetric(metric, "MDT", path, helpText)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
	return nil
}

func (s *lustreSource) generateMGSMetricTemplates() error {
	metricMap := map[string]map[string]string{
		"mgs/MGS/osd/": {
			"blocksize":            "Filesystem block size in bytes",
			"filesfree":            "The number of inodes (objects) available",
			"filestotal":           "The maximum number of inodes (objects) the filesystem can hold",
			"kbytesavail":          "Number of kilobytes readily available in the pool",
			"kbytesfree":           "Number of kilobytes allocated to the pool",
			"kbytestotal":          "Capacity of the pool in kilobytes",
			"quota_iused_estimate": "Returns '1' if a valid address is returned within the pool, referencing whether free space can be allocated",
		},
	}
	for path := range metricMap {
		for metric, helpText := range metricMap[path] {
			newMetric := newLustreProcMetric(metric, "MGS", path, helpText)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
	return nil
}

func (s *lustreSource) generateMDSMetricTemplates() error {
	metricMap := map[string]map[string]string{
		"mds/MDS/osd": {
			"blocksize":            "Filesystem block size in bytes",
			"filesfree":            "The number of inodes (objects) available",
			"filestotal":           "The maximum number of inodes (objects) the filesystem can hold",
			"kbytesavail":          "Number of kilobytes readily available in the pool",
			"kbytesfree":           "Number of kilobytes allocated to the pool",
			"kbytestotal":          "Capacity of the pool in kilobytes",
			"quota_iused_estimate": "Returns '1' if a valid address is returned within the pool, referencing whether free space can be allocated",
		},
	}
	for path := range metricMap {
		for metric, helpText := range metricMap[path] {
			newMetric := newLustreProcMetric(metric, "MDS", path, helpText)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
	return nil
}

func newLustreSource() (LustreSource, error) {
	var l lustreSource
	l.basePath = "/proc/fs/lustre"
	//control which node metrics you pull via flags
	l.generateOSTMetricTemplates()
	l.generateMDTMetricTemplates()
	l.generateMGSMetricTemplates()
	l.generateMDSMetricTemplates()
	return &l, nil
}

func (s *lustreSource) Update(ch chan<- prometheus.Metric) (err error) {
	metricType := "single"
	directoryDepth := 0

	for _, metric := range s.lustreProcMetrics {
		directoryDepth = strings.Count(metric.name, "/")
		paths, err := filepath.Glob(filepath.Join(s.basePath, metric.path, metric.name))
		if err != nil {
			return err
		}
		if paths == nil {
			continue
		}
		for _, path := range paths {
			metricType = "single"
			switch metric.name {
			case "brw_stats":
				err = s.parseBRWStats(metric.source, "brw_stats", path, directoryDepth, metric.helpText, func(nodeType string, brwOperation string, brwSize string, nodeName string, name string, helpText string, value uint64) {
					ch <- s.brwMetric(nodeType, brwOperation, brwSize, nodeName, name, helpText, value)
				})
				if err != nil {
					return err
				}
			case "job_stats":
				err = s.parseJobStats(metric.source, "job_stats", path, directoryDepth, metric.helpText, func(nodeType string, jobid string, jobOperation string, nodeName string, name string, helpText string, value uint64) {
					ch <- s.jobStatsMetric(nodeType, jobid, jobOperation, nodeName, name, helpText, value)
				})
				if err != nil {
					return err
				}
			default:
				if metric.name == "stats" {
					metricType = "stats"
				}
				err = s.parseFile(metric.source, metricType, path, directoryDepth, metric.helpText, func(nodeType string, nodeName string, name string, helpText string, value uint64) {
					ch <- s.constMetric(nodeType, nodeName, name, helpText, value)
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func parseReadWriteBytes(operation string, regexString string, statsFile string) (metricList []lustreStatsMetric, err error) {
	bytesString := regexCaptureString(regexString, statsFile)
	if len(bytesString) < 1 {
		return nil, nil
	}
	r, err := regexp.Compile(" +")
	if err != nil {
		return nil, err
	}
	bytesSplit := r.Split(bytesString, -1)
	// bytesSplit is in the following format:
	// bytesString: {name} {number of samples} 'samples' [{units}] {minimum} {maximum} {sum}
	// bytesSplit:   [0]    [1]                 [2]       [3]       [4]       [5]       [6]
	bytesMap := map[string]map[string]string{
		"_samples_total":      {"help": samplesHelp, "value": bytesSplit[1]},
		"_minimum_size_bytes": {"help": minimumHelp, "value": bytesSplit[4]},
		"_maximum_size_bytes": {"help": maximumHelp, "value": bytesSplit[5]},
		"_total_bytes":        {"help": totalHelp, "value": bytesSplit[6]},
	}
	for name, valueMap := range bytesMap {
		result, err := strconv.ParseUint(valueMap["value"], 10, 64)
		if err != nil {
			return nil, err
		}
		metricList = append(metricList, lustreStatsMetric{title: operation + name, help: valueMap["help"], value: result})
	}

	return metricList, nil
}

func splitBRWStats(statBlock string) (metricList []lustreBRWMetric, err error) {
	if len(statBlock) == 0 || statBlock == "" {
		return nil, nil
	}

	// Skip the first line of text as it doesn't contain any metrics
	for _, line := range strings.Split(statBlock, "\n")[1:] {
		if len(line) > 1 {
			fields := strings.Fields(line)
			// Lines are in the following format:
			// [size] [# read RPCs] [relative read size (%)] [cumulative read size (%)] | [# write RPCs] [relative write size (%)] [cumulative write size (%)]
			// [0]    [1]           [2]                      [3]                       [4] [5]           [6]                       [7]
			size, readRPCs, writeRPCs := fields[0], fields[1], fields[5]
			size = strings.Replace(size, ":", "", -1)
			metricList = append(metricList, lustreBRWMetric{size: size, operation: "read", value: readRPCs})
			metricList = append(metricList, lustreBRWMetric{size: size, operation: "write", value: writeRPCs})
		}
	}
	return metricList, nil
}

func parseStatsFile(path string) (metricList []lustreStatsMetric, err error) {
	operations := map[string]string{"read": "read_bytes .*", "write": "write_bytes .*"}
	statsFileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	statsFile := string(statsFileBytes[:])

	for operation, regexString := range operations {
		statsList, err := parseReadWriteBytes(operation, regexString, statsFile)
		if err != nil {
			return nil, err
		}
		if statsList != nil {
			for _, item := range statsList {
				metricList = append(metricList, item)
			}
		}
	}

	return metricList, nil
}

func getJobStatsByOperation(jobBlock string, jobID string, operation string) (metricList []lustreJobsMetric, err error) {
	opStat := regexCaptureString(operation+"_bytes: .*", jobBlock)
	opNumbers := regexCaptureStrings("[0-9*.[0-9]+|[0-9]+", opStat)

	opMap := map[string]map[string]string{
		"_samples": {"help": samplesHelp, "value": opNumbers[0]},
		"_minimum": {"help": minimumHelp, "value": opNumbers[1]},
		"_maximum": {"help": maximumHelp, "value": opNumbers[2]},
		"_total":   {"help": totalHelp, "value": opNumbers[3]},
	}
	for name, valueMap := range opMap {
		result, err := strconv.ParseUint(strings.TrimSpace(valueMap["value"]), 10, 64)
		if err != nil {
			return nil, err
		}
		l := lustreStatsMetric{
			title: "job_id_" + operation + name,
			help:  valueMap["help"],
			value: result,
		}
		metricList = append(metricList, lustreJobsMetric{jobID, operation, l})
	}

	return metricList, err
}

func getJobStats(jobBlock string, jobID string) (metricList []lustreJobsMetric, err error) {
	operations := []string{"read", "write"}
	for _, operation := range operations {
		statsList, err := getJobStatsByOperation(jobBlock, jobID, operation)
		if err != nil {
			return nil, err
		}
		if statsList != nil {
			for _, item := range statsList {
				metricList = append(metricList, item)
			}
		}
	}

	return metricList, nil
}

func getJobNum(jobBlock string) (jobID string, err error) {
	jobID = regexCaptureString("job_id: .*", jobBlock)
	jobID = regexCaptureString("[0-9]*.[0-9]+|[0-9]+", jobID)
	return strings.Trim(jobID, " "), nil
}

func parseJobStatsText(jobStats string) (metricList []lustreJobsMetric, err error) {
	jobs := regexCaptureStrings("(?ms:job_id:.*?(-|\\z))", jobStats)
	if len(jobs) < 1 {
		return nil, nil
	}

	for _, job := range jobs {
		jobID, err := getJobNum(job)
		if err != nil {
			return nil, err
		}
		jobList, err := getJobStats(job, jobID)
		if err != nil {
			return nil, err
		}
		for _, item := range jobList {
			metricList = append(metricList, item)
		}
	}
	return metricList, nil
}

func (s *lustreSource) parseJobStats(nodeType string, metricType string, path string, directoryDepth int, helpText string, handler func(string, string, string, string, string, string, uint64)) (err error) {
	_, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}
	jobStatsBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	jobStatsFile := string(jobStatsBytes[:])

	metricList, err := parseJobStatsText(jobStatsFile)
	if err != nil {
		return err
	}

	for _, item := range metricList {
		handler(nodeType, item.jobID, item.operation, nodeName, item.lustreStatsMetric.title, item.lustreStatsMetric.help, item.lustreStatsMetric.value)
	}
	return nil
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

func (s *lustreSource) parseBRWStats(nodeType string, metricType string, path string, directoryDepth int, helpText string, handler func(string, string, string, string, string, string, uint64)) (err error) {
	_, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}
	metricBlocks := map[string]string{
		"pages per bulk r/w":  pagesPerBlockRWHelp,
		"discontiguous pages": discontiguousPagesHelp,
		"disk I/Os in flight": diskIOsInFlightHelp,
		"I/O time":            ioTimeHelp,
		"disk I/O size":       diskIOSizeHelp,
	}
	statsFileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	statsFile := string(statsFileBytes[:])
	for title, help := range metricBlocks {
		block := regexCaptureString("(?ms:^"+title+".*?(\n\n|\\z))", statsFile)
		title = strings.Replace(title, " ", "_", -1)
		title = strings.Replace(title, "/", "", -1)
		metricList, err := splitBRWStats(block)
		if err != nil {
			return err
		}
		for _, item := range metricList {
			value, err := strconv.ParseUint(item.value, 10, 64)
			if err != nil {
				return err
			}
			handler(nodeType, item.operation, item.size, nodeName, title, help, value)
		}
	}
	return nil
}

func (s *lustreSource) parseFile(nodeType string, metricType string, path string, directoryDepth int, helpText string, handler func(string, string, string, string, uint64)) (err error) {
	name, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}
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
	case "stats":
		metricList, err := parseStatsFile(path)
		if err != nil {
			return err
		}

		for _, metric := range metricList {
			handler(nodeType, nodeName, metric.title, metric.help, metric.value)
		}
	}
	return nil
}

func (s *lustreSource) constMetric(nodeType string, nodeName string, name string, helpText string, value uint64) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", name),
			helpText,
			[]string{nodeType},
			nil,
		),
		prometheus.CounterValue,
		float64(value),
		nodeName,
	)
}

func (s *lustreSource) brwMetric(nodeType string, brwOperation string, brwSize string, nodeName string, name string, helpText string, value uint64) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", name),
			helpText,
			[]string{nodeType, "operation", "size"},
			nil,
		),
		prometheus.CounterValue,
		float64(value),
		nodeName,
		brwOperation,
		brwSize,
	)
}

func (s *lustreSource) jobStatsMetric(nodeType string, jobid string, jobOperation string, nodeName string, name string, helpText string, value uint64) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", name),
			helpText,
			[]string{nodeType, "jobid", "operation"},
			nil,
		),
		prometheus.CounterValue,
		float64(value),
		nodeName,
		jobid,
		jobOperation,
	)
}
