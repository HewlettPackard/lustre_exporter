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
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Help text dedicated to the 'stats' files
	readSamplesHelp  string = "Total number of reads that have been recorded."
	readMaximumHelp  string = "The maximum read size in bytes."
	readMinimumHelp  string = "The minimum read size in bytes."
	readTotalHelp    string = "The total number of bytes that have been read."
	writeSamplesHelp string = "Total number of writes that have been recorded."
	writeMaximumHelp string = "The maximum write size in bytes."
	writeMinimumHelp string = "The minimum write size in bytes."
	writeTotalHelp   string = "The total number of bytes that have been written."
	jobStatsHelp     string = "Number of operations the filesystem has performed."
	statsHelp        string = "Number of operations the filesystem has performed."

	// Help text dedicated to the 'brw_stats' file
	pagesPerBlockRWHelp    string = "Total number of pages per block RPC."
	discontiguousPagesHelp string = "Total number of logical discontinuities per RPC."
	ioTimeHelp             string = "Total time in milliseconds the filesystem has spent processing various object sizes."
	diskIOSizeHelp         string = "Total number of operations the filesystem has performed for the given size."
	diskIOsInFlightHelp    string = "Current number of I/O operations that are processing during the snapshot."

	// Help text dedicated to the 'rpc_stats' file
	pagesPerRPCHelp  string = "Total number of pages per RPC."
	rpcsInFlightHelp string = "Current number of RPCs that are processing during the snapshot."
	offsetHelp       string = "Current RPC offset by size."

	// Help text dedicated to the 'encrypt_page_pools' file
	physicalPagesHelp     string = "Capacity of physical memory."
	pagesPerPoolHelp      string = "Number of pages per pool."
	maxPagesHelp          string = "Maximum number of pages that can be held."
	maxPoolsHelp          string = "Number of pools."
	totalPagesHelp        string = "Number of pages in all pools."
	totalFreeHelp         string = "Current number of pages available."
	maxPagesReachedHelp   string = "Total number of pages reached."
	growsHelp             string = "Total number of grows."
	growsFailureHelp      string = "Total number of failures while attempting to add pages."
	shrinksHelp           string = "Total number of shrinks."
	cacheAccessHelp       string = "Total number of times cache has been accessed."
	cacheMissingHelp      string = "Total number of cache misses."
	lowFreeMarkHelp       string = "Lowest number of free pages reached."
	maxWaitQueueDepthHelp string = "Maximum waitqueue length."
	outOfMemHelp          string = "Total number of out of memory requests."

	// string mappings for 'health_check' values
	healthCheckHealthy   string = "1"
	healthCheckUnhealthy string = "0"

	//repeated strings replaced by constants
	mdStats          string = "md_stats"
	encryptPagePools string = "encrypt_page_pools"
)

var (
	// OstEnabled specifies whether to collect OST metrics
	OstEnabled bool
	// MdtEnabled specifies whether to collect MDT metrics
	MdtEnabled bool
	// MgsEnabled specifies whether to collect MGS metrics
	MgsEnabled bool
	// MdsEnabled specifies whether to collect MDS metrics
	MdsEnabled bool
	// ClientEnabled specifies whether to collect Client metrics
	ClientEnabled bool
	// GenericEnabled specifies whether to collect Generic metrics
	GenericEnabled bool
)

type lustreJobsMetric struct {
	jobID string
	lustreStatsMetric
}

type lustreBRWMetric struct {
	size      string
	operation string
	value     string
}

type multistatParsingStruct struct {
	index   int
	pattern string
}

func init() {
	Factories["procfs"] = newLustreSource
}

type lustreProcfsSource struct {
	lustreProcMetrics []lustreProcMetric
	basePath          string
}

func (s *lustreProcfsSource) generateOSTMetricTemplates() {
	metricMap := map[string][]lustreHelpStruct{
		"obdfilter/*": {
			{"blocksize", "blocksize_bytes", "Filesystem block size in bytes", s.gaugeMetric, false},
			{"brw_size", "brw_size_megabytes", "Block read/write size in megabytes", s.gaugeMetric, false},
			{"brw_stats", "pages_per_bulk_rw_total", pagesPerBlockRWHelp, s.counterMetric, false},
			{"brw_stats", "discontiguous_pages_total", discontiguousPagesHelp, s.counterMetric, false},
			{"brw_stats", "disk_io", diskIOsInFlightHelp, s.gaugeMetric, false},
			{"brw_stats", "io_time_milliseconds_total", ioTimeHelp, s.counterMetric, false},
			{"brw_stats", "disk_io_total", diskIOSizeHelp, s.counterMetric, false},
			{"degraded", "degraded", "Binary indicator as to whether or not the pool is degraded - 0 for not degraded, 1 for degraded", s.gaugeMetric, false},
			{"filesfree", "inodes_free", "The number of inodes (objects) available", s.gaugeMetric, false},
			{"filestotal", "inodes_maximum", "The maximum number of inodes (objects) the filesystem can hold", s.gaugeMetric, false},
			{"grant_compat_disable", "grant_compat_disabled", "Binary indicator as to whether clients with OBD_CONNECT_GRANT_PARAM setting will be granted space", s.gaugeMetric, false},
			{"grant_precreate", "grant_precreate_capacity_bytes", "Maximum space in bytes that clients can preallocate for objects", s.gaugeMetric, false},
			{"job_cleanup_interval", "job_cleanup_interval_seconds", "Interval in seconds between cleanup of tuning statistics", s.gaugeMetric, false},
			{"job_stats", "job_read_samples_total", readSamplesHelp, s.counterMetric, false},
			{"job_stats", "job_read_minimum_size_bytes", readMinimumHelp, s.gaugeMetric, false},
			{"job_stats", "job_read_maximum_size_bytes", readMaximumHelp, s.gaugeMetric, false},
			{"job_stats", "job_read_bytes_total", readTotalHelp, s.counterMetric, false},
			{"job_stats", "job_write_samples_total", writeSamplesHelp, s.counterMetric, false},
			{"job_stats", "job_write_minimum_size_bytes", writeMinimumHelp, s.gaugeMetric, false},
			{"job_stats", "job_write_maximum_size_bytes", writeMaximumHelp, s.gaugeMetric, false},
			{"job_stats", "job_write_bytes_total", writeTotalHelp, s.counterMetric, false},
			{"job_stats", "job_stats_total", jobStatsHelp, s.counterMetric, true},
			{"kbytesavail", "available_kilobytes", "Number of kilobytes readily available in the pool", s.gaugeMetric, false},
			{"kbytesfree", "free_kilobytes", "Number of kilobytes allocated to the pool", s.gaugeMetric, false},
			{"kbytestotal", "capacity_kilobytes", "Capacity of the pool in kilobytes", s.gaugeMetric, false},
			{"lfsck_speed_limit", "lfsck_speed_limit", "Maximum operations per second LFSCK (Lustre filesystem verification) can run", s.gaugeMetric, false},
			{"num_exports", "exports_total", "Total number of times the pool has been exported", s.counterMetric, false},
			{"precreate_batch", "precreate_batch", "Maximum number of objects that can be included in a single transaction", s.gaugeMetric, false},
			{"recovery_time_hard", "recovery_time_hard_seconds", "Maximum timeout 'recover_time_soft' can increment to for a single server", s.gaugeMetric, false},
			{"recovery_time_soft", "recovery_time_soft_seconds", "Duration in seconds for a client to attempt to reconnect after a crash (automatically incremented if servers are still in an error state)", s.gaugeMetric, false},
			{"soft_sync_limit", "soft_sync_limit", "Number of RPCs necessary before triggering a sync", s.gaugeMetric, false},
			{"stats", "read_samples_total", readSamplesHelp, s.counterMetric, false},
			{"stats", "read_minimum_size_bytes", readMinimumHelp, s.gaugeMetric, false},
			{"stats", "read_maximum_size_bytes", readMaximumHelp, s.gaugeMetric, false},
			{"stats", "read_bytes_total", readTotalHelp, s.counterMetric, false},
			{"stats", "write_samples_total", writeSamplesHelp, s.counterMetric, false},
			{"stats", "write_minimum_size_bytes", writeMinimumHelp, s.gaugeMetric, false},
			{"stats", "write_maximum_size_bytes", writeMaximumHelp, s.gaugeMetric, false},
			{"stats", "write_bytes_total", writeTotalHelp, s.counterMetric, false},
			{"stats", "stats_total", statsHelp, s.counterMetric, true},
			{"sync_journal", "sync_journal_enabled", "Binary indicator as to whether or not the journal is set for asynchronous commits", s.gaugeMetric, false},
			{"tot_dirty", "exports_dirty_total", "Total number of exports that have been marked dirty", s.counterMetric, false},
			{"tot_granted", "exports_granted_total", "Total number of exports that have been marked granted", s.counterMetric, false},
			{"tot_pending", "exports_pending_total", "Total number of exports that have been marked pending", s.counterMetric, false},
		},
		"ldlm/namespaces/filter-*": {
			{"lock_count", "lock_count_total", "Number of locks", s.counterMetric, false},
			{"lock_timeouts", "lock_timeout_total", "Number of lock timeouts", s.counterMetric, false},
			{"contended_locks", "lock_contended_total", "Number of contended locks", s.counterMetric, false},
			{"contention_seconds", "lock_contention_seconds_total", "Time in seconds during which locks were contended", s.counterMetric, false},
			{"pool/cancel", "lock_cancel_total", "Total number of cancelled locks", s.counterMetric, false},
			{"pool/cancel_rate", "lock_cancel_rate", "Lock cancel rate", s.gaugeMetric, false},
			{"pool/grant", "locks_grant_total", "Total number of granted locks", s.counterMetric, false},
			{"pool/granted", "locks_granted", "Number of granted less cancelled locks", s.untypedMetric, false},
			{"pool/grant_plan", "lock_grant_plan", "Number of planned lock grants per second", s.gaugeMetric, false},
			{"pool/grant_rate", "lock_grant_rate", "Lock grant rate", s.gaugeMetric, false},
			{"pool/recalc_freed", "recalc_freed_total", "Number of locks that have been freed", s.counterMetric, false},
			{"pool/recalc_timing", "recalc_timing_seconds_total", "Number of seconds spent locked", s.counterMetric, false},
			{"pool/shrink_freed", "shrink_freed_total", "Number of shrinks that have been freed", s.counterMetric, false},
			{"pool/shrink_request", "shrink_requests_total", "Number of shrinks that have been requested", s.counterMetric, false},
			{"pool/slv", "server_lock_volume", "Current value for server lock volume (SLV)", s.gaugeMetric, false},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			newMetric := newLustreProcMetric(item.filename, item.promName, "ost", path, item.helpText, item.hasMultipleVals, item.metricFunc)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
}

func (s *lustreProcfsSource) generateMDTMetricTemplates() {
	metricMap := map[string][]lustreHelpStruct{
		"mdt/*": {
			{mdStats, "stats_total", statsHelp, s.counterMetric, true},
			{"num_exports", "exports_total", "Total number of times the pool has been exported", s.counterMetric, false},
			{"job_stats", "job_stats_total", jobStatsHelp, s.counterMetric, true},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			newMetric := newLustreProcMetric(item.filename, item.promName, "mdt", path, item.helpText, item.hasMultipleVals, item.metricFunc)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
}

func (s *lustreProcfsSource) generateMGSMetricTemplates() {
	metricMap := map[string][]lustreHelpStruct{
		"mgs/MGS/osd/": {
			{"blocksize", "blocksize_bytes", "Filesystem block size in bytes", s.gaugeMetric, false},
			{"filesfree", "inodes_free", "The number of inodes (objects) available", s.gaugeMetric, false},
			{"filestotal", "inodes_maximum", "The maximum number of inodes (objects) the filesystem can hold", s.gaugeMetric, false},
			{"kbytesavail", "available_kilobytes", "Number of kilobytes readily available in the pool", s.gaugeMetric, false},
			{"kbytesfree", "free_kilobytes", "Number of kilobytes allocated to the pool", s.gaugeMetric, false},
			{"kbytestotal", "capacity_kilobytes", "Capacity of the pool in kilobytes", s.gaugeMetric, false},
			{"quota_iused_estimate", "quota_iused_estimate", "Returns '1' if a valid address is returned within the pool, referencing whether free space can be allocated", s.gaugeMetric, false},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			newMetric := newLustreProcMetric(item.filename, item.promName, "mgs", path, item.helpText, item.hasMultipleVals, item.metricFunc)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
}

func (s *lustreProcfsSource) generateMDSMetricTemplates() {
	metricMap := map[string][]lustreHelpStruct{
		"mds/MDS/osd": {
			{"blocksize", "blocksize_bytes", "Filesystem block size in bytes", s.gaugeMetric, false},
			{"filesfree", "inodes_free", "The number of inodes (objects) available", s.gaugeMetric, false},
			{"filestotal", "inodes_maximum", "The maximum number of inodes (objects) the filesystem can hold", s.gaugeMetric, false},
			{"kbytesavail", "available_kilobytes", "Number of kilobytes readily available in the pool", s.gaugeMetric, false},
			{"kbytesfree", "free_kilobytes", "Number of kilobytes allocated to the pool", s.gaugeMetric, false},
			{"kbytestotal", "capacity_kilobytes", "Capacity of the pool in kilobytes", s.gaugeMetric, false},
			{"quota_iused_estimate", "quota_iused_estimate", "Returns '1' if a valid address is returned within the pool, referencing whether free space can be allocated", s.gaugeMetric, false},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			newMetric := newLustreProcMetric(item.filename, item.promName, "mds", path, item.helpText, item.hasMultipleVals, item.metricFunc)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
}

func (s *lustreProcfsSource) generateClientMetricTemplates() {
	metricMap := map[string][]lustreHelpStruct{
		"llite/*": {
			{"blocksize", "blocksize_bytes", "Filesystem block size in bytes", s.gaugeMetric, false},
			{"checksum_pages", "checksum_pages_enabled", "Returns '1' if data checksumming is enabled for the client", s.gaugeMetric, false},
			{"default_easize", "default_ea_size_bytes", "Default Extended Attribute (EA) size in bytes", s.gaugeMetric, false},
			{"filesfree", "inodes_free", "The number of inodes (objects) available", s.gaugeMetric, false},
			{"filestotal", "inodes_maximum", "The maximum number of inodes (objects) the filesystem can hold", s.gaugeMetric, false},
			{"kbytesavail", "available_kilobytes", "Number of kilobytes readily available in the pool", s.gaugeMetric, false},
			{"kbytesfree", "free_kilobytes", "Number of kilobytes allocated to the pool", s.gaugeMetric, false},
			{"kbytestotal", "capacity_kilobytes", "Capacity of the pool in kilobytes", s.gaugeMetric, false},
			{"lazystatfs", "lazystatfs_enabled", "Returns '1' if lazystatfs (a non-blocking alternative to statfs) is enabled for the client", s.gaugeMetric, false},
			{"max_easize", "maximum_ea_size_bytes", "Maximum Extended Attribute (EA) size in bytes", s.gaugeMetric, false},
			{"max_read_ahead_mb", "maximum_read_ahead_megabytes", "Maximum number of megabytes to read ahead", s.gaugeMetric, false},
			{"max_read_ahead_per_file_mb", "maximum_read_ahead_per_file_megabytes", "Maximum number of megabytes per file to read ahead", s.gaugeMetric, false},
			{"max_read_ahead_whole_mb", "maximum_read_ahead_whole_megabytes", "Maximum file size in megabytes for a file to be read in its entirety", s.gaugeMetric, false},
			{"statahead_agl", "statahead_agl_enabled", "Returns '1' if the Asynchronous Glimpse Lock (AGL) for statahead is enabled", s.gaugeMetric, false},
			{"statahead_max", "statahead_maximum", "Maximum window size for statahead", s.gaugeMetric, false},
			{"stats", "read_samples_total", readSamplesHelp, s.counterMetric, false},
			{"stats", "read_minimum_size_bytes", readMinimumHelp, s.gaugeMetric, false},
			{"stats", "read_maximum_size_bytes", readMaximumHelp, s.gaugeMetric, false},
			{"stats", "read_bytes_total", readTotalHelp, s.counterMetric, false},
			{"stats", "write_samples_total", writeSamplesHelp, s.counterMetric, false},
			{"stats", "write_minimum_size_bytes", writeMinimumHelp, s.gaugeMetric, false},
			{"stats", "write_maximum_size_bytes", writeMaximumHelp, s.gaugeMetric, false},
			{"stats", "write_bytes_total", writeTotalHelp, s.counterMetric, false},
			{"stats", "stats_total", statsHelp, s.counterMetric, true},
			{"xattr_cache", "xattr_cache_enabled", "Returns '1' if extended attribute cache is enabled", s.gaugeMetric, false},
		},
		"mdc/*": {
			{"rpc_stats", "rpcs_in_flight", rpcsInFlightHelp, s.gaugeMetric, true},
		},
		"osc/*": {
			{"rpc_stats", "pages_per_rpc_total", pagesPerRPCHelp, s.counterMetric, false},
			{"rpc_stats", "rpcs_in_flight", rpcsInFlightHelp, s.gaugeMetric, true},
			{"rpc_stats", "rpcs_offset", offsetHelp, s.gaugeMetric, false},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			newMetric := newLustreProcMetric(item.filename, item.promName, "client", path, item.helpText, item.hasMultipleVals, item.metricFunc)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
}

func (s *lustreProcfsSource) generateGenericMetricTemplates() {
	metricMap := map[string][]lustreHelpStruct{
		"": {
			{"health_check", "health_check", "Current health status for the indicated instance: " + healthCheckHealthy + " refers to 'healthy', " + healthCheckUnhealthy + " refers to 'unhealthy'", s.gaugeMetric, false},
		},
		"sptlrpc": {
			{"encrypt_page_pools", "physical_pages", physicalPagesHelp, s.gaugeMetric, false},
			{"encrypt_page_pools", "pages_per_pool", pagesPerPoolHelp, s.gaugeMetric, false},
			{"encrypt_page_pools", "maximum_pages", maxPagesHelp, s.gaugeMetric, false},
			{"encrypt_page_pools", "maximum_pools", maxPoolsHelp, s.gaugeMetric, false},
			{"encrypt_page_pools", "pages_in_pools", totalPagesHelp, s.gaugeMetric, false},
			{"encrypt_page_pools", "free_pages", totalFreeHelp, s.gaugeMetric, false},
			{"encrypt_page_pools", "maximum_pages_reached_total", maxPagesReachedHelp, s.counterMetric, false},
			{"encrypt_page_pools", "grows_total", growsHelp, s.counterMetric, false},
			{"encrypt_page_pools", "grows_failure_total", growsFailureHelp, s.counterMetric, false},
			{"encrypt_page_pools", "shrinks_total", shrinksHelp, s.counterMetric, false},
			{"encrypt_page_pools", "cache_access_total", cacheAccessHelp, s.counterMetric, false},
			{"encrypt_page_pools", "cache_miss_total", cacheMissingHelp, s.counterMetric, false},
			{"encrypt_page_pools", "free_page_low", lowFreeMarkHelp, s.gaugeMetric, false},
			{"encrypt_page_pools", "maximum_waitqueue_depth", maxWaitQueueDepthHelp, s.gaugeMetric, false},
			{"encrypt_page_pools", "out_of_memory_request_total", outOfMemHelp, s.counterMetric, false},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			newMetric := newLustreProcMetric(item.filename, item.promName, "Generic", path, item.helpText, item.hasMultipleVals, item.metricFunc)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
}

func newLustreSource() LustreSource {
	var l lustreProcfsSource
	l.basePath = filepath.Join(ProcLocation, "fs/lustre")
	//control which node metrics you pull via flags
	if OstEnabled {
		l.generateOSTMetricTemplates()
	}
	if MdtEnabled {
		l.generateMDTMetricTemplates()
	}
	if MgsEnabled {
		l.generateMGSMetricTemplates()
	}
	if MdsEnabled {
		l.generateMDSMetricTemplates()
	}
	if ClientEnabled {
		l.generateClientMetricTemplates()
	}
	if GenericEnabled {
		l.generateGenericMetricTemplates()
	}
	return &l
}

func (s *lustreProcfsSource) Update(ch chan<- prometheus.Metric) (err error) {
	var directoryDepth int

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
			switch metric.filename {
			case "health_check":
				err = s.parseTextFile(metric.source, path, directoryDepth, metric.helpText, metric.promName, func(nodeType string, nodeName string, name string, helpText string, value float64) {
					ch <- metric.metricFunc([]string{nodeType}, []string{nodeName}, name, helpText, value)
				})
				if err != nil {
					return err
				}
			case "brw_stats", "rpc_stats":
				err = s.parseBRWStats(metric.source, path, directoryDepth, metric.helpText, metric.promName, metric.hasMultipleVals, func(nodeType string, brwOperation string, brwSize string, nodeName string, name string, helpText string, value float64, extraLabel string, extraLabelValue string) {
					if extraLabelValue == "" {
						ch <- metric.metricFunc([]string{nodeType, "operation", "size"}, []string{nodeName, brwOperation, brwSize}, name, helpText, value)
					} else {
						ch <- metric.metricFunc([]string{nodeType, "operation", "size", extraLabel}, []string{nodeName, brwOperation, brwSize, extraLabelValue}, name, helpText, value)
					}
				})
				if err != nil {
					return err
				}
			case "job_stats":
				err = s.parseJobStats(metric.source, path, directoryDepth, metric.helpText, metric.promName, metric.hasMultipleVals, func(nodeType string, jobid string, nodeName string, name string, helpText string, value float64, extraLabel string, extraLabelValue string) {
					if extraLabelValue == "" {
						ch <- metric.metricFunc([]string{nodeType, "jobid"}, []string{nodeName, jobid}, name, helpText, value)
					} else {
						ch <- metric.metricFunc([]string{nodeType, "jobid", extraLabel}, []string{nodeName, jobid, extraLabelValue}, name, helpText, value)
					}
				})
				if err != nil {
					return err
				}
			default:
				err = s.parseFile(metric.source, metric.filename, path, directoryDepth, metric.helpText, metric.promName, metric.hasMultipleVals, func(nodeType string, nodeName string, name string, helpText string, value float64, extraLabel string, extraLabelValue string) {
					if extraLabelValue == "" {
						ch <- metric.metricFunc([]string{nodeType}, []string{nodeName}, name, helpText, value)
					} else {
						ch <- metric.metricFunc([]string{nodeType, extraLabel}, []string{nodeName, extraLabelValue}, name, helpText, value)
					}
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func getStatsOperationMetrics(statsFile string, promName string, helpText string) (metricList []lustreStatsMetric, err error) {
	operationSlice := []multistatParsingStruct{
		{pattern: "open", index: 1},
		{pattern: "close", index: 1},
		{pattern: "getattr", index: 1},
		{pattern: "setattr", index: 1},
		{pattern: "getxattr", index: 1},
		{pattern: "setxattr", index: 1},
		{pattern: "statfs", index: 1},
		{pattern: "seek", index: 1},
		{pattern: "readdir", index: 1},
		{pattern: "truncate", index: 1},
		{pattern: "alloc_inode", index: 1},
		{pattern: "removexattr", index: 1},
		{pattern: "unlink", index: 1},
		{pattern: "inode_permission", index: 1},
		{pattern: "create", index: 1},
		{pattern: "get_info", index: 1},
		{pattern: "set_info_async", index: 1},
		{pattern: "connect", index: 1},
		{pattern: "ping", index: 1},
	}
	for _, operation := range operationSlice {
		opStat := regexCaptureString(operation.pattern+" .*", statsFile)
		if len(opStat) < 1 {
			continue
		}
		r, err := regexp.Compile(" +")
		if err != nil {
			continue
		}
		bytesSplit := r.Split(opStat, -1)
		result, err := strconv.ParseFloat(bytesSplit[operation.index], 64)
		if err != nil {
			return nil, err
		}
		l := lustreStatsMetric{
			title:           promName,
			help:            helpText,
			value:           result,
			extraLabel:      "operation",
			extraLabelValue: operation.pattern,
		}
		metricList = append(metricList, l)
	}
	return metricList, nil
}

func getStatsIOMetrics(statsFile string, promName string, helpText string) (metricList []lustreStatsMetric, err error) {
	// bytesSplit is in the following format:
	// bytesString: {name} {number of samples} 'samples' [{units}] {minimum} {maximum} {sum}
	// bytesSplit:   [0]    [1]                 [2]       [3]       [4]       [5]       [6]
	bytesMap := map[string]multistatParsingStruct{
		readSamplesHelp:       {pattern: "read_bytes .*", index: 1},
		readMinimumHelp:       {pattern: "read_bytes .*", index: 4},
		readMaximumHelp:       {pattern: "read_bytes .*", index: 5},
		readTotalHelp:         {pattern: "read_bytes .*", index: 6},
		writeSamplesHelp:      {pattern: "write_bytes .*", index: 1},
		writeMinimumHelp:      {pattern: "write_bytes .*", index: 4},
		writeMaximumHelp:      {pattern: "write_bytes .*", index: 5},
		writeTotalHelp:        {pattern: "write_bytes .*", index: 6},
		physicalPagesHelp:     {pattern: "physical pages: .*", index: 2},
		pagesPerPoolHelp:      {pattern: "pages per pool: .*", index: 3},
		maxPagesHelp:          {pattern: "max pages: .*", index: 2},
		maxPoolsHelp:          {pattern: "max pools: .*", index: 2},
		totalPagesHelp:        {pattern: "total pages: .*", index: 2},
		totalFreeHelp:         {pattern: "total free: .*", index: 2},
		maxPagesReachedHelp:   {pattern: "max pages reached: .*", index: 3},
		growsHelp:             {pattern: "grows: .*", index: 1},
		growsFailureHelp:      {pattern: "grows failure: .*", index: 2},
		shrinksHelp:           {pattern: "shrinks: .*", index: 1},
		cacheAccessHelp:       {pattern: "cache access: .*", index: 2},
		cacheMissingHelp:      {pattern: "cache missing: .*", index: 2},
		lowFreeMarkHelp:       {pattern: "low free mark: .*", index: 3},
		maxWaitQueueDepthHelp: {pattern: "max waitqueue depth: .*", index: 3},
		outOfMemHelp:          {pattern: "out of mem: .*", index: 3},
	}
	pattern := bytesMap[helpText].pattern
	bytesString := regexCaptureString(pattern, statsFile)
	if len(bytesString) < 1 {
		return nil, nil
	}
	r, err := regexp.Compile(" +")
	if err != nil {
		return nil, err
	}
	bytesSplit := r.Split(bytesString, -1)
	result, err := strconv.ParseFloat(bytesSplit[bytesMap[helpText].index], 64)
	if err != nil {
		return nil, err
	}
	l := lustreStatsMetric{
		title:           promName,
		help:            helpText,
		value:           result,
		extraLabel:      "",
		extraLabelValue: "",
	}
	metricList = append(metricList, l)

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
			if len(fields) >= 6 {
				size, readRPCs, writeRPCs := fields[0], fields[1], fields[5]
				size = strings.Replace(size, ":", "", -1)
				metricList = append(metricList, lustreBRWMetric{size: size, operation: "read", value: readRPCs})
				metricList = append(metricList, lustreBRWMetric{size: size, operation: "write", value: writeRPCs})
			} else if len(fields) >= 1 {
				size, rpcs := fields[0], fields[1]
				size = strings.Replace(size, ":", "", -1)
				metricList = append(metricList, lustreBRWMetric{size: size, operation: "read", value: rpcs})
			} else {
				continue
			}
		}
	}
	return metricList, nil
}

func parseStatsFile(helpText string, promName string, path string, hasMultipleVals bool) (metricList []lustreStatsMetric, err error) {
	statsFileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	statsFile := string(statsFileBytes[:])
	var statsList []lustreStatsMetric
	if hasMultipleVals {
		statsList, err = getStatsOperationMetrics(statsFile, promName, helpText)
	} else {
		statsList, err = getStatsIOMetrics(statsFile, promName, helpText)
	}
	if err != nil {
		return nil, err
	}
	if statsList != nil {
		metricList = append(metricList, statsList...)
	}

	return metricList, nil
}

func getJobStatsIOMetrics(jobBlock string, jobID string, promName string, helpText string) (metricList []lustreJobsMetric, err error) {
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
	// If the metric isn't located in the map, don't try to parse a value for it.
	if _, exists := opMap[helpText]; !exists {
		return nil, nil
	}
	pattern := opMap[helpText].pattern
	opStat := regexCaptureString(pattern+": .*", jobBlock)
	opNumbers := regexCaptureNumbers(opStat)
	if len(opNumbers) < 1 {
		return nil, nil
	}
	result, err := strconv.ParseFloat(strings.TrimSpace(opNumbers[opMap[helpText].index]), 64)
	if err != nil {
		return nil, err
	}
	l := lustreStatsMetric{
		title:           promName,
		help:            helpText,
		value:           result,
		extraLabel:      "",
		extraLabelValue: "",
	}
	metricList = append(metricList, lustreJobsMetric{jobID, l})

	return metricList, err
}

func getJobNum(jobBlock string) (jobID string, err error) {
	jobID = regexCaptureString("job_id: .*", jobBlock)
	jobIDlist := regexCaptureNumbers(jobID)
	if len(jobIDlist) < 1 {
		return "", nil
	}
	return strings.Trim(jobIDlist[0], " "), nil
}

func getJobStatsOperationMetrics(jobBlock string, jobID string, promName string, helpText string) (metricList []lustreJobsMetric, err error) {
	operationSlice := []multistatParsingStruct{
		{index: 0, pattern: "open"},
		{index: 0, pattern: "close"},
		{index: 0, pattern: "mknod"},
		{index: 0, pattern: "link"},
		{index: 0, pattern: "unlink"},
		{index: 0, pattern: "mkdir"},
		{index: 0, pattern: "rmdir"},
		{index: 0, pattern: "rename"},
		{index: 0, pattern: "getattr"},
		{index: 0, pattern: "setattr"},
		{index: 0, pattern: "getxattr"},
		{index: 0, pattern: "setxattr"},
		{index: 0, pattern: "statfs"},
		{index: 0, pattern: "sync"},
		{index: 0, pattern: "samedir_rename"},
		{index: 0, pattern: "crossdir_rename"},
		{index: 0, pattern: "punch"},
		{index: 0, pattern: "destroy"},
		{index: 0, pattern: "create"},
		{index: 0, pattern: "get_info"},
		{index: 0, pattern: "set_info"},
		{index: 0, pattern: "quotactl"},
	}
	for _, operation := range operationSlice {
		opStat := regexCaptureString(operation.pattern+": .*", jobBlock)
		opNumbers := regexCaptureStrings("[0-9]*\\.[0-9]+|[0-9]+", opStat)
		if len(opNumbers) < 1 {
			continue
		}
		var result float64
		result, err = strconv.ParseFloat(strings.TrimSpace(opNumbers[operation.index]), 64)
		if err != nil {
			return nil, err
		}
		l := lustreStatsMetric{
			title:           promName,
			help:            helpText,
			value:           result,
			extraLabel:      "operation",
			extraLabelValue: operation.pattern,
		}
		metricList = append(metricList, lustreJobsMetric{jobID, l})
	}
	return metricList, err
}

func parseJobStatsText(jobStats string, promName string, helpText string, hasMultipleVals bool) (metricList []lustreJobsMetric, err error) {
	jobs := regexCaptureStrings("(?ms:job_id:.*?(-|\\z))", jobStats)
	if len(jobs) < 1 {
		return nil, nil
	}
	var jobList []lustreJobsMetric
	for _, job := range jobs {
		jobID, err := getJobNum(job)
		if err != nil {
			return nil, err
		}
		if hasMultipleVals {
			jobList, err = getJobStatsOperationMetrics(job, jobID, promName, helpText)
		} else {
			jobList, err = getJobStatsIOMetrics(job, jobID, promName, helpText)
		}
		if err != nil {
			return nil, err
		}
		if jobList != nil {
			metricList = append(metricList, jobList...)
		}
	}
	return metricList, nil
}

func (s *lustreProcfsSource) parseJobStats(nodeType string, path string, directoryDepth int, helpText string, promName string, hasMultipleVals bool, handler func(string, string, string, string, string, float64, string, string)) (err error) {
	_, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}
	jobStatsBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	jobStatsFile := string(jobStatsBytes[:])

	metricList, err := parseJobStatsText(jobStatsFile, promName, helpText, hasMultipleVals)
	if err != nil {
		return err
	}

	for _, item := range metricList {
		handler(nodeType, item.jobID, nodeName, item.lustreStatsMetric.title, item.lustreStatsMetric.help, item.lustreStatsMetric.value, item.lustreStatsMetric.extraLabel, item.lustreStatsMetric.extraLabelValue)
	}
	return nil
}

func (s *lustreProcfsSource) parseBRWStats(nodeType string, path string, directoryDepth int, helpText string, promName string, hasMultipleVals bool, handler func(string, string, string, string, string, string, float64, string, string)) (err error) {
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
		pagesPerRPCHelp:        "pages per rpc",
		rpcsInFlightHelp:       "rpcs in flight",
		offsetHelp:             "offset",
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
	extraLabel := ""
	extraLabelValue := ""
	if hasMultipleVals {
		extraLabel = "type"
		pathElements := strings.Split(path, "/")
		extraLabelValue = pathElements[len(pathElements)-3]
	}
	for _, item := range metricList {
		value, err := strconv.ParseFloat(item.value, 64)
		if err != nil {
			return err
		}
		handler(nodeType, item.operation, convertToBytes(item.size), nodeName, promName, helpText, value, extraLabel, extraLabelValue)
	}
	return nil
}

func (s *lustreProcfsSource) parseTextFile(nodeType string, path string, directoryDepth int, helpText string, promName string, handler func(string, string, string, string, float64)) (err error) {
	filename, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	fileString := string(fileBytes[:])
	switch filename {
	case "health_check":
		if strings.TrimSpace(fileString) == "healthy" {
			value, err := strconv.ParseFloat(strings.TrimSpace(healthCheckHealthy), 64)
			if err != nil {
				return err
			}
			handler(nodeType, nodeName, promName, helpText, value)
		} else {
			value, err := strconv.ParseFloat(strings.TrimSpace(healthCheckUnhealthy), 64)
			if err != nil {
				return err
			}
			handler(nodeType, nodeName, promName, helpText, value)
		}
	}
	return nil
}

func (s *lustreProcfsSource) parseFile(nodeType string, metricType string, path string, directoryDepth int, helpText string, promName string, hasMultipleVals bool, handler func(string, string, string, string, float64, string, string)) (err error) {
	_, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}
	switch metricType {
	case stats, mdStats, encryptPagePools:
		metricList, err := parseStatsFile(helpText, promName, path, hasMultipleVals)
		if err != nil {
			return err
		}

		for _, metric := range metricList {
			handler(nodeType, nodeName, metric.title, metric.help, metric.value, metric.extraLabel, metric.extraLabelValue)
		}
	default:
		value, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		convertedValue, err := strconv.ParseFloat(strings.TrimSpace(string(value)), 64)
		if err != nil {
			return err
		}
		handler(nodeType, nodeName, promName, helpText, convertedValue, "", "")
	}
	return nil
}

func (s *lustreProcfsSource) counterMetric(labels []string, labelValues []string, name string, helpText string, value float64) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", name),
			helpText,
			labels,
			nil,
		),
		prometheus.CounterValue,
		value,
		labelValues...,
	)
}

func (s *lustreProcfsSource) gaugeMetric(labels []string, labelValues []string, name string, helpText string, value float64) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", name),
			helpText,
			labels,
			nil,
		),
		prometheus.GaugeValue,
		value,
		labelValues...,
	)
}

func (s *lustreProcfsSource) untypedMetric(labels []string, labelValues []string, name string, helpText string, value float64) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", name),
			helpText,
			labels,
			nil,
		),
		prometheus.UntypedValue,
		value,
		labelValues...,
	)
}
