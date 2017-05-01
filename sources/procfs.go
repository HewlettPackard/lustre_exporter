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

func parseReadWriteBytes(operation string, regexString string, statsFile string) (metricMap map[string]map[string]string, err error) {
	bytesRegex, err := regexp.Compile(regexString)
	if err != nil {
		return nil, err
	}

	bytesString := bytesRegex.FindString(statsFile)
	if len(bytesString) == 0 {
		return nil, nil
	}

	r, err := regexp.Compile(" +")
	if err != nil {
		return nil, err
	}

	metricMap = make(map[string]map[string]string)
	bytesSplit := r.Split(bytesString, -1)
	// bytesSplit is in the following format:
	// bytesString: {name} {number of samples} 'samples' [{units}] {minimum} {maximum} {sum}
	// bytesSplit:   [0]    [1]                 [2]       [3]       [4]       [5]       [6]
	metricMap[operation+"_samples_total"] = map[string]string{"help": samplesHelp, "value": bytesSplit[1]}
	metricMap[operation+"_minimum_size_bytes"] = map[string]string{"help": minimumHelp, "value": bytesSplit[4]}
	metricMap[operation+"_maximum_size_bytes"] = map[string]string{"help": maximumHelp, "value": bytesSplit[5]}
	metricMap[operation+"_total_bytes"] = map[string]string{"help": totalHelp, "value": bytesSplit[6]}

	return metricMap, nil
}

func splitBRWStats(title string, statBlock string) (metricMap map[string]map[string]string, err error) {
	title = strings.Replace(title, " ", "_", -1)
	title = strings.Replace(title, "/", "", -1)
	metricMap = make(map[string]map[string]string)

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
			metricMap[title+"_"+size+"_read"] = map[string]string{"value": readRPCs, "size": size, "operation": "read", "name": title}
			metricMap[title+"_"+size+"_write"] = map[string]string{"value": writeRPCs, "size": size, "operation": "write", "name": title}
		}
	}
	return metricMap, nil
}

func parseStatsFile(path string) (metricMap map[string]map[string]string, err error) {
	metricMap = make(map[string]map[string]string)
	statsFileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	statsFile := string(statsFileBytes[:])

	readStatsMap, err := parseReadWriteBytes("read", "read_bytes .*", statsFile)
	if err != nil {
		return nil, err
	}
	if readStatsMap != nil {
		for key, value := range readStatsMap {
			metricMap[key] = value
		}
	}

	writeStatsMap, err := parseReadWriteBytes("write", "write_bytes .*", statsFile)
	if err != nil {
		return nil, err
	}
	if writeStatsMap != nil {
		for key, value := range writeStatsMap {
			metricMap[key] = value
		}
	}

	return metricMap, nil
}

func getJobStatsByOperation(jobBlock string, jobID string, operation string) (metricMap map[string]map[string]string, err error) {
	opRegex, err := regexp.Compile(operation + "_bytes: .*")
	if err != nil {
		return nil, err
	}
	numbersRegex, err := regexp.Compile("[0-9]*.[0-9]+|[0-9]+")
	if err != nil {
		return nil, err
	}

	opStat := opRegex.FindString(jobBlock)
	if len(opStat) == 0 {
		return nil, nil
	}
	opNumbers := numbersRegex.FindAllString(opStat, -1)
	if len(opNumbers) == 0 {
		return nil, nil
	}

	metricMap = make(map[string]map[string]string)

	metricMap["job_id_"+operation+"_samples"] = map[string]string{"help": samplesHelp, "value": strings.TrimSpace(opNumbers[0]), "jobID": jobID, "operation": operation, "name": "job_id_" + operation + "_samples"}
	metricMap["job_id_"+operation+"_minimum"] = map[string]string{"help": minimumHelp, "value": strings.TrimSpace(opNumbers[1]), "jobID": jobID, "operation": operation, "name": "job_id_" + operation + "_minimum"}
	metricMap["job_id_"+operation+"_maximum"] = map[string]string{"help": maximumHelp, "value": strings.TrimSpace(opNumbers[2]), "jobID": jobID, "operation": operation, "name": "job_id_" + operation + "_maximum"}
	metricMap["job_id_"+operation+"_total"] = map[string]string{"help": totalHelp, "value": strings.TrimSpace(opNumbers[3]), "jobID": jobID, "operation": operation, "name": "job_id_" + operation + "_total"}

	return metricMap, err
}

func getJobStats(jobBlock string, jobID string) (metricMap map[string]map[string]string, err error) {
	metricMap = make(map[string]map[string]string)
	readStatsMap, err := getJobStatsByOperation(jobBlock, jobID, "read")
	if err != nil {
		return nil, err
	}
	if readStatsMap != nil {
		for key, value := range readStatsMap {
			metricMap[key] = value
		}
	}

	writeStatsMap, err := getJobStatsByOperation(jobBlock, jobID, "write")
	if err != nil {
		return nil, err
	}
	if writeStatsMap != nil {
		for key, value := range writeStatsMap {
			metricMap[key] = value
		}
	}

	return metricMap, nil
}

func getJobNum(jobBlock string) (jobID string, err error) {
	numbersRegex, err := regexp.Compile("[0-9]*.[0-9]+|[0-9]+")
	if err != nil {
		return "", err
	}
	jobIDRegex, err := regexp.Compile("job_id: .*")
	if err != nil {
		return "", err
	}

	jobID = jobIDRegex.FindString(jobBlock)
	if len(jobID) == 0 {
		return "", nil
	}
	jobID = numbersRegex.FindString(jobID)
	if len(jobID) == 0 {
		return "", nil
	}

	return strings.Trim(jobID, " "), nil
}

func parseJobStatsText(jobStats string) (metricMap map[string]map[string]string, err error) {
	metricMap = make(map[string]map[string]string)
	pattern := "(?ms:job_id:.*?(-|\\z))"
	jobRegex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	jobs := jobRegex.FindAllString(jobStats, -1)
	if len(jobs) == 0 {
		return nil, nil
	}

	for _, job := range jobs {
		jobID, err := getJobNum(job)
		if err != nil {
			return nil, err
		}
		jobMap, err := getJobStats(job, jobID)
		if err != nil {
			return nil, err
		}
		for key, value := range jobMap {
			metricMap[key+"_"+jobID] = value
		}
	}
	return metricMap, nil
}

func (s *lustreSource) parseJobStats(nodeType string, metricType string, path string, directoryDepth int, helpText string, handler func(string, string, string, string, string, string, uint64)) (err error) {
	metricMap := make(map[string]map[string]string)
	_, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}
	jobStatsBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	jobStatsFile := string(jobStatsBytes[:])

	metricMap, err = parseJobStatsText(jobStatsFile)
	if err != nil {
		return err
	}

	for _, value := range metricMap {
		result, err := strconv.ParseUint(value["value"], 10, 64)
		if err != nil {
			return err
		}
		handler(nodeType, value["jobID"], value["operation"], nodeName, value["name"], value["help"], result)
	}
	return nil
}

func extractStatsBlock(title string, statsFile string) (block string) {
	// The following expressions match the specified block in the text or the end of the string,
	// whichever comes first.
	pattern := "(?ms:^" + title + ".*?(\n\n|\\z))"
	re := regexp.MustCompile(pattern)
	block = re.FindString(statsFile)
	return block
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
		block := extractStatsBlock(title, statsFile)
		mapSubset, err := splitBRWStats(title, block)
		if err != nil {
			return err
		}
		for _, metricMap := range mapSubset {
			value, err := strconv.ParseUint(metricMap["value"], 10, 64)
			if err != nil {
				return err
			}
			handler(nodeType, metricMap["operation"], metricMap["size"], nodeName, metricMap["name"], help, value)
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
		metricMap, err := parseStatsFile(path)
		if err != nil {
			return err
		}

		for key, statMap := range metricMap {
			value, err := strconv.ParseUint(statMap["value"], 10, 64)
			if err != nil {
				return err
			}
			handler(nodeType, nodeName, key, statMap["help"], value)
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
