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
	source    string //The node type (OSS, MDS, MGS)
	path      string //Path to retreive metric from
	helpText  string
}

func init() {
	Factories["procfs"] = NewLustreSource
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

func (s *lustreSource) generateOSSMetricTemplates() error {
	metricMap := map[string]map[string]string{
		"obdfilter/*": map[string]string{
			"blocksize":            "Filesystem block size in bytes",
			"brw_size":             "Block read/write size in bytes",
			"brw_stats":            "A collection of block read/write statistics",
			"degraded":             "Binary indicator as to whether or not the pool is degraded - 0 for not degraded, 1 for degraded",
			"filesfree":            "The number of inodes (objects) available",
			"filestotal":           "The maximum number of inodes (objects) the filesystem can hold",
			"grant_compat_disable": "Binary indicator as to whether clients with OBD_CONNECT_GRANT_PARAM setting will be granted space",
			"grant_precreate":      "Maximum space in bytes that clients can preallocate for objects",
			"job_cleanup_interval": "Interval in seconds between cleanup of tuning statistics",
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
	}
	for path, _ := range metricMap {
		for metric, helpText := range metricMap[path] {
			newMetric := newLustreProcMetric(metric, "OSS", path, helpText)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
	return nil
}

func (s *lustreSource) generateMGSMetricTemplates() error {
	metricMap := map[string]map[string]string{
		"mgs/MGS/osd/": map[string]string{
			"blocksize":            "Filesystem block size in bytes",
			"filesfree":            "The number of inodes (objects) available",
			"filestotal":           "The maximum number of inodes (objects) the filesystem can hold",
			"kbytesavail":          "Number of kilobytes readily available in the pool",
			"kbytesfree":           "Number of kilobytes allocated to the pool",
			"kbytestotal":          "Capacity of the pool in kilobytes",
			"quota_iused_estimate": "Returns '1' if a valid address is returned within the pool, referencing whether free space can be allocated",
		},
	}
	for path, _ := range metricMap {
		for metric, helpText := range metricMap[path] {
			newMetric := newLustreProcMetric(metric, "MGS", path, helpText)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
	return nil
}

func (s *lustreSource) generateMDSMetricTemplates() error {
	metricMap := map[string]map[string]string{
		"mds/MDS/osd": map[string]string{
			"blocksize":            "Filesystem block size in bytes",
			"filesfree":            "The number of inodes (objects) available",
			"filestotal":           "The maximum number of inodes (objects) the filesystem can hold",
			"kbytesavail":          "Number of kilobytes readily available in the pool",
			"kbytesfree":           "Number of kilobytes allocated to the pool",
			"kbytestotal":          "Capacity of the pool in kilobytes",
			"quota_iused_estimate": "Returns '1' if a valid address is returned within the pool, referencing whether free space can be allocated",
		},
	}
	for path, _ := range metricMap {
		for metric, helpText := range metricMap[path] {
			newMetric := newLustreProcMetric(metric, "MDS", path, helpText)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
	return nil
}

func NewLustreSource() (LustreSource, error) {
	var l lustreSource
	l.basePath = "/proc/fs/lustre"
	//control which node metrics you pull via flags
	l.generateOSSMetricTemplates()
	l.generateMGSMetricTemplates()
	l.generateMDSMetricTemplates()
	return &l, nil
}

func (s *lustreSource) Update(ch chan<- prometheus.Metric) (err error) {
	metricType := "single"

	for _, metric := range s.lustreProcMetrics {
		paths, err := filepath.Glob(filepath.Join(s.basePath, metric.path, metric.name))
		if err != nil {
			return err
		}
		if paths == nil {
			continue
		}
		for _, path := range paths {
			switch metric.name {
			case "brw_stats":
				err = s.parseBRWStats(metric.source, "brw_stats", path, metric.helpText, func(nodeType string, brwOperation string, brwSize string, nodeName string, name string, helpText string, value uint64) {
					ch <- s.brwMetric(nodeType, brwOperation, brwSize, nodeName, name, helpText, value)
				})
				if err != nil {
					return err
				}
			default:
				if metric.name == "stats" {
					metricType = "stats"
				}
				err = s.parseFile(metric.source, metricType, path, metric.helpText, func(nodeType string, nodeName string, name string, helpText string, value uint64) {
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

func parseReadWriteBytes(regexString string, statsFile string) (metricMap map[string]map[string]uint64, err error) {
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

	bytesSplit := r.Split(bytesString, -1)
	// bytesSplit is in the following format:
	// bytesString: {name} {number of samples} 'samples' [{units}] {minimum} {maximum} {sum}
	// bytesSplit:   [0]    [1]                 [2]       [3]       [4]       [5]       [6]
	samples, err := strconv.ParseUint(bytesSplit[1], 10, 64)
	if err != nil {
		return nil, err
	}
	minimum, err := strconv.ParseUint(bytesSplit[4], 10, 64)
	if err != nil {
		return nil, err
	}
	maximum, err := strconv.ParseUint(bytesSplit[5], 10, 64)
	if err != nil {
		return nil, err
	}
	total, err := strconv.ParseUint(bytesSplit[6], 10, 64)
	if err != nil {
		return nil, err
	}

	metricMap = make(map[string]map[string]uint64)

	metricMap["samples_total"] = map[string]uint64{samplesHelp: samples}
	metricMap["minimum_size_bytes"] = map[string]uint64{minimumHelp: minimum}
	metricMap["maximum_size_bytes"] = map[string]uint64{maximumHelp: maximum}
	metricMap["total_bytes"] = map[string]uint64{totalHelp: total}

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

func parseStatsFile(path string) (metricMap map[string]map[string]map[string]uint64, err error) {
	statsFileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	statsFile := string(statsFileBytes[:])

	readStatsMap, err := parseReadWriteBytes("read_bytes .*", statsFile)
	if err != nil {
		return nil, err
	}

	writeStatsMap, err := parseReadWriteBytes("write_bytes .*", statsFile)
	if err != nil {
		return nil, err
	}

	metricMap = make(map[string]map[string]map[string]uint64)
	metricMap["read"] = readStatsMap
	metricMap["write"] = writeStatsMap

	return metricMap, nil
}

func extractStatsBlock(title string, statsFile string) (block string) {
	// The following expressions match the specified block in the text or the end of the string,
	// whichever comes first.
	pattern := "(?ms:^" + title + ".*?(\n\n|\\z))"
	re := regexp.MustCompile(pattern)
	block = re.FindString(statsFile)
	return block
}

func parseFileElements(path string) (name string, nodeName string, err error) {
	pathElements := strings.Split(path, "/")
	pathLen := len(pathElements)
	if pathLen < 1 {
		return "", "", fmt.Errorf("path did not return at least one element")
	}
	name = pathElements[pathLen-1]
	nodeName = pathElements[pathLen-2]
	return name, nodeName, nil
}

func (s *lustreSource) parseBRWStats(nodeType string, metricType string, path string, helpText string, handler func(string, string, string, string, string, string, uint64)) (err error) {
	_, nodeName, err := parseFileElements(path)
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

func (s *lustreSource) parseFile(nodeType string, metricType string, path string, helpText string, handler func(string, string, string, string, uint64)) (err error) {
	name, nodeName, err := parseFileElements(path)
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

		for statType, statMap := range metricMap {
			for key, metricMap := range statMap {
				metricName := statType + "_" + key
				for detailedHelp, value := range metricMap {
					handler(nodeType, nodeName, metricName, detailedHelp, value)
				}
			}
		}
	}
	return nil
}

func (s *lustreSource) constMetric(nodeType string, nodeName string, name string, helpText string, value uint64) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "lustre", name),
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
			prometheus.BuildFQName(Namespace, "lustre", name),
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
