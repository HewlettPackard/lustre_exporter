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
	readSamplesHelp  string = "Total number of reads that have been recorded for the given job."
	readMaximumHelp  string = "The maximum read size in bytes for the given job."
	readMinimumHelp  string = "The minimum read size in bytes for the given job."
	readTotalHelp    string = "The total number of bytes that have been read for the given job."
	writeSamplesHelp string = "Total number of writes that have been recorded for the given job."
	writeMaximumHelp string = "The maximum write size in bytes for the given job."
	writeMinimumHelp string = "The minimum write size in bytes for the given job."
	writeTotalHelp   string = "The total number of bytes that have been written for the given job."

	// Help text dedicated to the 'brw_stats' file
	pagesPerBlockRWHelp    string = "Total number of pages per RPC."
	discontiguousPagesHelp string = "Total number of logical discontinuities per RPC."
	ioTimeHelp             string = "Total time in milliseconds the filesystem has spent processing various object sizes."
	diskIOSizeHelp         string = "Total number of operations the filesystem has performed for the given size."
	diskIOsInFlightHelp    string = "Current number of I/O operations that are processing during the snapshot."
)

type lustreProcMetric struct {
	subsystem string
	filename  string
	promName  string
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
	jobID string
	lustreStatsMetric
}

type lustreBRWMetric struct {
	size      string
	operation string
	value     string
}

type lustreHelpStruct struct {
	filename string
	promName string // Name to be used in Prometheus
	helpText string
}

type multistatParsingStruct struct {
	index   int
	pattern string
}

func init() {
	Factories["procfs"] = newLustreSource
}

type lustreSource struct {
	lustreProcMetrics []lustreProcMetric
	basePath          string
}

func newLustreProcMetric(filename string, promName string, source string, path string, helpText string) lustreProcMetric {
	var m lustreProcMetric
	m.filename = filename
	m.promName = promName
	m.source = source
	m.path = path
	m.helpText = helpText

	return m
}

func (s *lustreSource) generateOSTMetricTemplates() error {
	metricMap := map[string][]lustreHelpStruct{
		"obdfilter/*": {
			{"blocksize", "blocksize", "Filesystem block size in bytes"},
			{"brw_size", "brw_size", "Block read/write size in bytes"},
			{"brw_stats", "pages_per_bulk_rw", pagesPerBlockRWHelp},
			{"brw_stats", "discontiguous_pages", discontiguousPagesHelp},
			{"brw_stats", "disk_ios_in_flight", diskIOsInFlightHelp},
			{"brw_stats", "io_time", ioTimeHelp},
			{"brw_stats", "disk_io_size", diskIOSizeHelp},
			{"degraded", "degraded", "Binary indicator as to whether or not the pool is degraded - 0 for not degraded, 1 for degraded"},
			{"filesfree", "filesfree", "The number of inodes (objects) available"},
			{"filestotal", "filestotal", "The maximum number of inodes (objects) the filesystem can hold"},
			{"grant_compat_disable", "grant_compat_disable", "Binary indicator as to whether clients with OBD_CONNECT_GRANT_PARAM setting will be granted space"},
			{"grant_precreate", "grant_precreate", "Maximum space in bytes that clients can preallocate for objects"},
			{"job_cleanup_interval", "job_cleanup_interval", "Interval in seconds between cleanup of tuning statistics"},
			{"job_stats", "job_id_read_samples", readSamplesHelp},
			{"job_stats", "job_id_read_minimum", readMinimumHelp},
			{"job_stats", "job_id_read_maximum", readMaximumHelp},
			{"job_stats", "job_id_read_total", readTotalHelp},
			{"job_stats", "job_id_write_samples", writeSamplesHelp},
			{"job_stats", "job_id_write_minimum", writeMinimumHelp},
			{"job_stats", "job_id_write_maximum", writeMaximumHelp},
			{"job_stats", "job_id_write_total", writeTotalHelp},
			{"kbytesavail", "kbytesavail", "Number of kilobytes readily available in the pool"},
			{"kbytesfree", "kbytesfree", "Number of kilobytes allocated to the pool"},
			{"kbytestotal", "kbytestotal", "Capacity of the pool in kilobytes"},
			{"lfsck_speed_limit", "lfsck_speed_limit", "Maximum operations per second LFSCK (Lustre filesystem verification) can run"},
			{"num_exports", "num_exports", "Total number of times the pool has been exported"},
			{"precreate_batch", "precreate_batch", "Maximum number of objects that can be included in a single transaction"},
			{"recovery_time_hard", "recovery_time_hard", "Maximum timeout 'recover_time_soft' can increment to for a single server"},
			{"recovery_time_soft", "recovery_time_soft", "Duration in seconds for a client to attempt to reconnect after a crash (automatically incremented if servers are still in an error state)"},
			{"soft_sync_limit", "soft_sync_limit", "Number of RPCs necessary before triggering a sync"},
			{"stats", "stats", "A collection of statistics specific to Lustre"},
			{"sync_journal", "sync_journal", "Binary indicator as to whether or not the journal is set for asynchronous commits"},
			{"tot_dirty", "tot_dirty", "Total number of exports that have been marked dirty"},
			{"tot_granted", "tot_granted", "Total number of exports that have been marked granted"},
			{"tot_pending", "tot_pending", "Total number of exports that have been marked pending"},
		},
		"ldlm/namespaces/filter-*": {
			{"lock_count", "lock_count", "Number of locks"},
			{"lock_timeouts", "lock_timeouts", "Number of lock timeouts"},
			{"contended_locks", "contended_locks", "Number of contended locks"},
			{"contention_seconds", "contention_seconds", "Time in seconds during which locks were contended"},
			{"pool/granted", "granted", "Number of granted locks"},
			{"pool/grant_rate", "grant_rate", "Lock grant rate"},
			{"pool/cancel_rate", "cancel_rate", "Lock cancel rate"},
			{"pool/grant_speed", "grant_speed", "Lock grant speed"},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			newMetric := newLustreProcMetric(item.filename, item.promName, "OST", path, item.helpText)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
	return nil
}

func (s *lustreSource) generateMDTMetricTemplates() error {
	metricMap := map[string][]lustreHelpStruct{
		"mdt/*": {
			{"num_exports", "num_exports", "Total number of times the pool has been exported"},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			newMetric := newLustreProcMetric(item.filename, item.promName, "MDT", path, item.helpText)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
	return nil
}

func (s *lustreSource) generateMGSMetricTemplates() error {
	metricMap := map[string][]lustreHelpStruct{
		"mgs/MGS/osd/": {
			{"blocksize", "blocksize", "Filesystem block size in bytes"},
			{"filesfree", "filesfree", "The number of inodes (objects) available"},
			{"filestotal", "filestotal", "The maximum number of inodes (objects) the filesystem can hold"},
			{"kbytesavail", "kbytesavail", "Number of kilobytes readily available in the pool"},
			{"kbytesfree", "kbytesfree", "Number of kilobytes allocated to the pool"},
			{"kbytestotal", "kbytestotal", "Capacity of the pool in kilobytes"},
			{"quota_iused_estimate", "quota_iused_estimate", "Returns '1' if a valid address is returned within the pool, referencing whether free space can be allocated"},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			newMetric := newLustreProcMetric(item.filename, item.promName, "MGS", path, item.helpText)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
	return nil
}

func (s *lustreSource) generateMDSMetricTemplates() error {
	metricMap := map[string][]lustreHelpStruct{
		"mds/MDS/osd": {
			{"blocksize", "blocksize", "Filesystem block size in bytes"},
			{"filesfree", "filesfree", "The number of inodes (objects) available"},
			{"filestotal", "filestotal", "The maximum number of inodes (objects) the filesystem can hold"},
			{"kbytesavail", "kbytesavail", "Number of kilobytes readily available in the pool"},
			{"kbytesfree", "kbytesfree", "Number of kilobytes allocated to the pool"},
			{"kbytestotal", "kbytestotal", "Capacity of the pool in kilobytes"},
			{"quota_iused_estimate", "quota_iused_estimate", "Returns '1' if a valid address is returned within the pool, referencing whether free space can be allocated"},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			newMetric := newLustreProcMetric(item.filename, item.promName, "MDS", path, item.helpText)
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
		directoryDepth = strings.Count(metric.filename, "/")
		paths, err := filepath.Glob(filepath.Join(s.basePath, metric.path, metric.filename))
		if err != nil {
			return err
		}
		if paths == nil {
			continue
		}
		for _, path := range paths {
			metricType = "single"
			switch metric.filename {
			case "brw_stats":
				err = s.parseBRWStats(metric.source, "brw_stats", path, directoryDepth, metric.helpText, metric.promName, func(nodeType string, brwOperation string, brwSize string, nodeName string, name string, helpText string, value uint64) {
					ch <- s.brwMetric(nodeType, brwOperation, brwSize, nodeName, name, helpText, value)
				})
				if err != nil {
					return err
				}
			case "job_stats":
				err = s.parseJobStats(metric.source, "job_stats", path, directoryDepth, metric.helpText, metric.promName, func(nodeType string, jobid string, nodeName string, name string, helpText string, value uint64) {
					ch <- s.jobStatsMetric(nodeType, jobid, nodeName, name, helpText, value)
				})
				if err != nil {
					return err
				}
			default:
				if metric.filename == "stats" {
					metricType = "stats"
				}
				err = s.parseFile(metric.source, metricType, path, directoryDepth, metric.helpText, metric.promName, func(nodeType string, nodeName string, name string, helpText string, value uint64) {
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
		"_samples_total":      {"help": readSamplesHelp, "value": bytesSplit[1]},
		"_minimum_size_bytes": {"help": readMinimumHelp, "value": bytesSplit[4]},
		"_maximum_size_bytes": {"help": readMaximumHelp, "value": bytesSplit[5]},
		"_total_bytes":        {"help": readTotalHelp, "value": bytesSplit[6]},
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

func getJobStatsByOperation(jobBlock string, jobID string, promName string, helpText string) (metricList []lustreJobsMetric, err error) {
	// opMap matches the given helpText value with the placement of the numeric fields within each metric line.
	// For example, the number of samples is the first number in the line and has a helpText of readSamplesHelp,
	// hence the 'index' value of 0. 'pattern' is the regex capture pattern for the desired line.
	opMap := map[string]multistatParsingStruct{
		readSamplesHelp:  {index: 0, pattern: "read_bytes"},
		readMinimumHelp:  {index: 1, pattern: "read_bytes"},
		readMaximumHelp:  {index: 2, pattern: "read_bytes"},
		readTotalHelp:    {index: 3, pattern: "read_bytes"},
		writeSamplesHelp: {index: 0, pattern: "write_bytes"},
		writeMinimumHelp: {index: 1, pattern: "write_bytes"},
		writeMaximumHelp: {index: 2, pattern: "write_bytes"},
		writeTotalHelp:   {index: 3, pattern: "write_bytes"},
	}
	pattern := opMap[helpText].pattern
	opStat := regexCaptureString(pattern+": .*", jobBlock)
	opNumbers := regexCaptureStrings("[0-9]*.[0-9]+|[0-9]+", opStat)
	result, err := strconv.ParseUint(strings.TrimSpace(opNumbers[opMap[helpText].index]), 10, 64)
	if err != nil {
		return nil, err
	}
	l := lustreStatsMetric{
		title: promName,
		help:  helpText,
		value: result,
	}
	metricList = append(metricList, lustreJobsMetric{jobID, l})

	return metricList, err
}

func getJobStats(jobBlock string, jobID string, promName string, helpText string) (metricList []lustreJobsMetric, err error) {
	statsList, err := getJobStatsByOperation(jobBlock, jobID, promName, helpText)
	if err != nil {
		return nil, err
	}
	if statsList != nil {
		for _, item := range statsList {
			metricList = append(metricList, item)
		}
	}

	return metricList, nil
}

func getJobNum(jobBlock string) (jobID string, err error) {
	jobID = regexCaptureString("job_id: .*", jobBlock)
	jobID = regexCaptureString("[0-9]*.[0-9]+|[0-9]+", jobID)
	return strings.Trim(jobID, " "), nil
}

func parseJobStatsText(jobStats string, promName string, helpText string) (metricList []lustreJobsMetric, err error) {
	jobs := regexCaptureStrings("(?ms:job_id:.*?(-|\\z))", jobStats)
	if len(jobs) < 1 {
		return nil, nil
	}

	for _, job := range jobs {
		jobID, err := getJobNum(job)
		if err != nil {
			return nil, err
		}
		jobList, err := getJobStats(job, jobID, promName, helpText)
		if err != nil {
			return nil, err
		}
		for _, item := range jobList {
			metricList = append(metricList, item)
		}
	}
	return metricList, nil
}

func (s *lustreSource) parseJobStats(nodeType string, metricType string, path string, directoryDepth int, helpText string, promName string, handler func(string, string, string, string, string, uint64)) (err error) {
	_, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}
	jobStatsBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	jobStatsFile := string(jobStatsBytes[:])

	metricList, err := parseJobStatsText(jobStatsFile, promName, helpText)
	if err != nil {
		return err
	}

	for _, item := range metricList {
		handler(nodeType, item.jobID, nodeName, item.lustreStatsMetric.title, item.lustreStatsMetric.help, item.lustreStatsMetric.value)
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

func (s *lustreSource) parseBRWStats(nodeType string, metricType string, path string, directoryDepth int, helpText string, promName string, handler func(string, string, string, string, string, string, uint64)) (err error) {
	_, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}
	metricBlocks := map[string]string{
		pagesPerBlockRWHelp:    "pages per bulk r/w",
		discontiguousPagesHelp: "discontiguous pages",
		diskIOsInFlightHelp:    "disk I/Os in flight",
		ioTimeHelp:             "I/O time",
		diskIOSizeHelp:         "disk I/O size",
	}
	statsFileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	statsFile := string(statsFileBytes[:])
	block := regexCaptureString("(?ms:^"+metricBlocks[helpText]+".*?(\n\n|\\z))", statsFile)
	metricList, err := splitBRWStats(block)
	if err != nil {
		return err
	}
	for _, item := range metricList {
		value, err := strconv.ParseUint(item.value, 10, 64)
		if err != nil {
			return err
		}
		handler(nodeType, item.operation, item.size, nodeName, promName, helpText, value)
	}
	return nil
}

func (s *lustreSource) parseFile(nodeType string, metricType string, path string, directoryDepth int, helpText string, promName string, handler func(string, string, string, string, uint64)) (err error) {
	_, nodeName, err := parseFileElements(path, directoryDepth)
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
		handler(nodeType, nodeName, promName, helpText, convertedValue)
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

func (s *lustreSource) jobStatsMetric(nodeType string, jobid string, nodeName string, name string, helpText string, value uint64) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", name),
			helpText,
			[]string{nodeType, "jobid"},
			nil,
		),
		prometheus.CounterValue,
		float64(value),
		nodeName,
		jobid,
	)
}
