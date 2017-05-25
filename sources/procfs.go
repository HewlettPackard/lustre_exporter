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
	// Help text dedicated to the 'stats' files
	readSamplesHelp     string = "Total number of reads that have been recorded."
	readMaximumHelp     string = "The maximum read size in bytes."
	readMinimumHelp     string = "The minimum read size in bytes."
	readTotalHelp       string = "The total number of bytes that have been read."
	writeSamplesHelp    string = "Total number of writes that have been recorded."
	writeMaximumHelp    string = "The maximum write size in bytes."
	writeMinimumHelp    string = "The minimum write size in bytes."
	writeTotalHelp      string = "The total number of bytes that have been written."
	openHelp            string = "Number of open operations the filesystem has performed."
	closeHelp           string = "Number of close operations the filesystem has performed."
	mknodHelp           string = "Number of mknod operations the filesystem has performed."
	linkHelp            string = "Number of link operations the filesystem has performed."
	unlinkHelp          string = "Number of unlink operations the filesystem has performed."
	mkdirHelp           string = "Number of mkdir operations the filesystem has performed."
	rmdirHelp           string = "Number of rmdir operations the filesystem has performed."
	renameHelp          string = "Number of rename operations the filesystem has performed."
	getattrHelp         string = "Number of getattr operations the filesystem has performed."
	setattrHelp         string = "Number of setattr operations the filesystem has performed."
	getxattrHelp        string = "Number of getxattr operations the filesystem has performed."
	setxattrHelp        string = "Number of setxattr operations the filesystem has performed."
	statfsHelp          string = "Number of statfs operations the filesystem has performed."
	syncHelp            string = "Number of sync operations the filesystem has performed."
	samedirRenameHelp   string = "Number of samedir_rename operations the filesystem has performed."
	crossdirRenameHelp  string = "Number of crossdir_rename operations the filesystem has performed."
	punchHelp           string = "Number of punch operations the filesystem has performed."
	destroyHelp         string = "Number of destroy operations the filesystem has performed."
	createHelp          string = "Number of create operations the filesystem has performed."
	getInfoHelp         string = "Number of get_info operations the filesystem has performed."
	setInfoHelp         string = "Number of set_info operations the filesystem has performed."
	quotactlHelp        string = "Number of quotactl operations the filesystem has performed."
	seekHelp            string = "Number of seek operations the filesystem has performed."
	readdirHelp         string = "Number of readdir operations the filesystem has performed."
	truncateHelp        string = "Number of truncate operations the filesystem has performed."
	allocInodeHelp      string = "Number of alloc operations the filesystem has performed."
	removexattrHelp     string = "Number of removexattr operations the filesystem has performed."
	inodePermissionHelp string = "Number of inode_permission operations the filesystem has performed."
	setInfoAsyncHelp    string = "Number of set_info_async operations the filesystem has performed."
	connectHelp         string = "Number of connect operations the filesystem has performed."
	pingHelp            string = "Number of ping operations the filesystem has performed."

	// Help text dedicated to the 'brw_stats' file
	pagesPerBlockRWHelp    string = "Total number of pages per RPC."
	discontiguousPagesHelp string = "Total number of logical discontinuities per RPC."
	ioTimeHelp             string = "Total time in milliseconds the filesystem has spent processing various object sizes."
	diskIOSizeHelp         string = "Total number of operations the filesystem has performed for the given size."
	diskIOsInFlightHelp    string = "Current number of I/O operations that are processing during the snapshot."

	// string mappings for 'health_check' values
	healthCheckHealthy   string = "1"
	healthCheckUnhealthy string = "0"
)

type prometheusType func([]string, []string, string, string, uint64) prometheus.Metric

type lustreProcMetric struct {
	filename   string
	promName   string
	source     string //The parent data source (OST, MDS, MGS, etc)
	path       string //Path to retrieve metric from
	helpText   string
	metricFunc prometheusType
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
	filename   string
	promName   string // Name to be used in Prometheus
	helpText   string
	metricFunc prometheusType
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

func newLustreProcMetric(filename string, promName string, source string, path string, helpText string, metricFunc prometheusType) lustreProcMetric {
	var m lustreProcMetric
	m.filename = filename
	m.promName = promName
	m.source = source
	m.path = path
	m.helpText = helpText
	m.metricFunc = metricFunc

	return m
}

func (s *lustreSource) generateOSTMetricTemplates() error {
	metricMap := map[string][]lustreHelpStruct{
		"obdfilter/*": {
			{"blocksize", "blocksize_bytes", "Filesystem block size in bytes", s.gaugeMetric},
			{"brw_size", "brw_size_megabytes", "Block read/write size in megabytes", s.gaugeMetric},
			{"brw_stats", "pages_per_bulk_rw_total", pagesPerBlockRWHelp, s.counterMetric},
			{"brw_stats", "discontiguous_pages_total", discontiguousPagesHelp, s.counterMetric},
			{"brw_stats", "disk_io", diskIOsInFlightHelp, s.gaugeMetric},
			{"brw_stats", "io_time_milliseconds_total", ioTimeHelp, s.counterMetric},
			{"brw_stats", "disk_io_total", diskIOSizeHelp, s.counterMetric},
			{"degraded", "degraded", "Binary indicator as to whether or not the pool is degraded - 0 for not degraded, 1 for degraded", s.gaugeMetric},
			{"filesfree", "inodes_free", "The number of inodes (objects) available", s.gaugeMetric},
			{"filestotal", "inodes_maximum", "The maximum number of inodes (objects) the filesystem can hold", s.gaugeMetric},
			{"grant_compat_disable", "grant_compat_disabled", "Binary indicator as to whether clients with OBD_CONNECT_GRANT_PARAM setting will be granted space", s.gaugeMetric},
			{"grant_precreate", "grant_precreate_capacity_bytes", "Maximum space in bytes that clients can preallocate for objects", s.gaugeMetric},
			{"job_cleanup_interval", "job_cleanup_interval_seconds", "Interval in seconds between cleanup of tuning statistics", s.gaugeMetric},
			{"job_stats", "job_read_samples_total", readSamplesHelp, s.counterMetric},
			{"job_stats", "job_read_minimum_size_bytes", readMinimumHelp, s.gaugeMetric},
			{"job_stats", "job_read_maximum_size_bytes", readMaximumHelp, s.gaugeMetric},
			{"job_stats", "job_read_bytes_total", readTotalHelp, s.counterMetric},
			{"job_stats", "job_write_samples_total", writeSamplesHelp, s.counterMetric},
			{"job_stats", "job_write_minimum_size_bytes", writeMinimumHelp, s.gaugeMetric},
			{"job_stats", "job_write_maximum_size_bytes", writeMaximumHelp, s.gaugeMetric},
			{"job_stats", "job_write_bytes_total", writeTotalHelp, s.counterMetric},
			{"job_stats", "job_getattr_total", getattrHelp, s.counterMetric},
			{"job_stats", "job_setattr_total", setattrHelp, s.counterMetric},
			{"job_stats", "job_punch_total", punchHelp, s.counterMetric},
			{"job_stats", "job_sync_total", syncHelp, s.counterMetric},
			{"job_stats", "job_destroy_total", destroyHelp, s.counterMetric},
			{"job_stats", "job_create_total", createHelp, s.counterMetric},
			{"job_stats", "job_statfs_total", statfsHelp, s.counterMetric},
			{"job_stats", "job_get_info_total", getInfoHelp, s.counterMetric},
			{"job_stats", "job_set_info_total", setInfoHelp, s.counterMetric},
			{"job_stats", "job_quotactl_total", quotactlHelp, s.counterMetric},
			{"kbytesavail", "available_kilobytes", "Number of kilobytes readily available in the pool", s.gaugeMetric},
			{"kbytesfree", "free_kilobytes", "Number of kilobytes allocated to the pool", s.gaugeMetric},
			{"kbytestotal", "capacity_kilobytes", "Capacity of the pool in kilobytes", s.gaugeMetric},
			{"lfsck_speed_limit", "lfsck_speed_limit", "Maximum operations per second LFSCK (Lustre filesystem verification) can run", s.gaugeMetric},
			{"num_exports", "exports_total", "Total number of times the pool has been exported", s.counterMetric},
			{"precreate_batch", "precreate_batch", "Maximum number of objects that can be included in a single transaction", s.gaugeMetric},
			{"recovery_time_hard", "recovery_time_hard_seconds", "Maximum timeout 'recover_time_soft' can increment to for a single server", s.gaugeMetric},
			{"recovery_time_soft", "recovery_time_soft_seconds", "Duration in seconds for a client to attempt to reconnect after a crash (automatically incremented if servers are still in an error state)", s.gaugeMetric},
			{"soft_sync_limit", "soft_sync_limit", "Number of RPCs necessary before triggering a sync", s.gaugeMetric},
			{"stats", "read_samples_total", readSamplesHelp, s.counterMetric},
			{"stats", "read_minimum_size_bytes", readMinimumHelp, s.gaugeMetric},
			{"stats", "read_maximum_size_bytes", readMaximumHelp, s.gaugeMetric},
			{"stats", "read_bytes_total", readTotalHelp, s.counterMetric},
			{"stats", "write_samples_total", writeSamplesHelp, s.counterMetric},
			{"stats", "write_minimum_size_bytes", writeMinimumHelp, s.gaugeMetric},
			{"stats", "write_maximum_size_bytes", writeMaximumHelp, s.gaugeMetric},
			{"stats", "write_bytes_total", writeTotalHelp, s.counterMetric},
			{"stats", "create_total", createHelp, s.counterMetric},
			{"stats", "statfs_total", statfsHelp, s.counterMetric},
			{"stats", "get_info_total", getInfoHelp, s.counterMetric},
			{"stats", "set_info_async_total", setInfoAsyncHelp, s.counterMetric},
			{"stats", "connect_total", connectHelp, s.counterMetric},
			{"stats", "ping_total", pingHelp, s.counterMetric},
			{"sync_journal", "sync_journal_enabled", "Binary indicator as to whether or not the journal is set for asynchronous commits", s.gaugeMetric},
			{"tot_dirty", "exports_dirty_total", "Total number of exports that have been marked dirty", s.counterMetric},
			{"tot_granted", "exports_granted_total", "Total number of exports that have been marked granted", s.counterMetric},
			{"tot_pending", "exports_pending_total", "Total number of exports that have been marked pending", s.counterMetric},
		},
		"ldlm/namespaces/filter-*": {
			{"lock_count", "lock_count_total", "Number of locks", s.counterMetric},
			{"lock_timeouts", "lock_timeout_total", "Number of lock timeouts", s.counterMetric},
			{"contended_locks", "lock_contended_total", "Number of contended locks", s.counterMetric},
			{"contention_seconds", "lock_contention_seconds_total", "Time in seconds during which locks were contended", s.counterMetric},
			{"pool/cancel", "lock_cancel_total", "Total number of cancelled locks", s.counterMetric},
			{"pool/cancel_rate", "lock_cancel_rate", "Lock cancel rate", s.gaugeMetric},
			{"pool/grant", "locks_grant_total", "Total number of granted locks", s.counterMetric},
			{"pool/granted", "locks_granted", "Number of granted less cancelled locks", s.untypedMetric},
			{"pool/grant_plan", "lock_grant_plan", "Number of planned lock grants per second", s.gaugeMetric},
			{"pool/grant_rate", "lock_grant_rate", "Lock grant rate", s.gaugeMetric},
			{"pool/recalc_freed", "recalc_freed_total", "Number of locks that have been freed", s.counterMetric},
			{"pool/recalc_timing", "recalc_timing_seconds_total", "Number of seconds spent locked", s.counterMetric},
			{"pool/shrink_freed", "shrink_freed_total", "Number of shrinks that have been freed", s.counterMetric},
			{"pool/shrink_request", "shrink_requests_total", "Number of shrinks that have been requested", s.counterMetric},
			{"pool/slv", "server_lock_volume", "Current value for server lock volume (SLV)", s.gaugeMetric},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			newMetric := newLustreProcMetric(item.filename, item.promName, "ost", path, item.helpText, item.metricFunc)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
	return nil
}

func (s *lustreSource) generateMDTMetricTemplates() error {
	metricMap := map[string][]lustreHelpStruct{
		"mdt/*": {
			{"md_stats", "open_total", openHelp, s.counterMetric},
			{"md_stats", "close_total", closeHelp, s.counterMetric},
			{"md_stats", "getattr_total", getattrHelp, s.counterMetric},
			{"md_stats", "setattr_total", setattrHelp, s.counterMetric},
			{"md_stats", "getxattr_total", getxattrHelp, s.counterMetric},
			{"md_stats", "setxattr_total", setxattrHelp, s.counterMetric},
			{"md_stats", "statfs_total", statfsHelp, s.counterMetric},
			{"md_stats", "mknod_total", mknodHelp, s.counterMetric},
			{"md_stats", "link_total", linkHelp, s.counterMetric},
			{"md_stats", "unlink_total", unlinkHelp, s.counterMetric},
			{"md_stats", "mkdir_total", mkdirHelp, s.counterMetric},
			{"md_stats", "rmdir_total", rmdirHelp, s.counterMetric},
			{"md_stats", "rename_total", renameHelp, s.counterMetric},
			{"md_stats", "sync_total", syncHelp, s.counterMetric},
			{"md_stats", "samedir_rename_total", samedirRenameHelp, s.counterMetric},
			{"md_stats", "crossdir_rename_total", crossdirRenameHelp, s.counterMetric},
			{"num_exports", "exports_total", "Total number of times the pool has been exported", s.counterMetric},
			{"job_stats", "job_open_total", openHelp, s.counterMetric},
			{"job_stats", "job_close_total", closeHelp, s.counterMetric},
			{"job_stats", "job_mknod_total", mknodHelp, s.counterMetric},
			{"job_stats", "job_link_total", linkHelp, s.counterMetric},
			{"job_stats", "job_unlink_total", unlinkHelp, s.counterMetric},
			{"job_stats", "job_mkdir_total", mkdirHelp, s.counterMetric},
			{"job_stats", "job_rmdir_total", rmdirHelp, s.counterMetric},
			{"job_stats", "job_rename_total", renameHelp, s.counterMetric},
			{"job_stats", "job_getattr_total", getattrHelp, s.counterMetric},
			{"job_stats", "job_setattr_total", setattrHelp, s.counterMetric},
			{"job_stats", "job_getxattr_total", getxattrHelp, s.counterMetric},
			{"job_stats", "job_setxattr_total", setxattrHelp, s.counterMetric},
			{"job_stats", "job_statfs_total", statfsHelp, s.counterMetric},
			{"job_stats", "job_sync_total", syncHelp, s.counterMetric},
			{"job_stats", "job_samedir_rename_total", samedirRenameHelp, s.counterMetric},
			{"job_stats", "job_crossdir_rename_total", crossdirRenameHelp, s.counterMetric},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			newMetric := newLustreProcMetric(item.filename, item.promName, "mdt", path, item.helpText, item.metricFunc)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
	return nil
}

func (s *lustreSource) generateMGSMetricTemplates() error {
	metricMap := map[string][]lustreHelpStruct{
		"mgs/MGS/osd/": {
			{"blocksize", "blocksize_bytes", "Filesystem block size in bytes", s.gaugeMetric},
			{"filesfree", "inodes_free", "The number of inodes (objects) available", s.gaugeMetric},
			{"filestotal", "inodes_maximum", "The maximum number of inodes (objects) the filesystem can hold", s.gaugeMetric},
			{"kbytesavail", "available_kilobytes", "Number of kilobytes readily available in the pool", s.gaugeMetric},
			{"kbytesfree", "free_kilobytes", "Number of kilobytes allocated to the pool", s.gaugeMetric},
			{"kbytestotal", "capacity_kilobytes", "Capacity of the pool in kilobytes", s.gaugeMetric},
			{"quota_iused_estimate", "quota_iused_estimate", "Returns '1' if a valid address is returned within the pool, referencing whether free space can be allocated", s.gaugeMetric},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			newMetric := newLustreProcMetric(item.filename, item.promName, "mgs", path, item.helpText, item.metricFunc)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
	return nil
}

func (s *lustreSource) generateMDSMetricTemplates() error {
	metricMap := map[string][]lustreHelpStruct{
		"mds/MDS/osd": {
			{"blocksize", "blocksize_bytes", "Filesystem block size in bytes", s.gaugeMetric},
			{"filesfree", "inodes_free", "The number of inodes (objects) available", s.gaugeMetric},
			{"filestotal", "inodes_maximum", "The maximum number of inodes (objects) the filesystem can hold", s.gaugeMetric},
			{"kbytesavail", "available_kilobytes", "Number of kilobytes readily available in the pool", s.gaugeMetric},
			{"kbytesfree", "free_kilobytes", "Number of kilobytes allocated to the pool", s.gaugeMetric},
			{"kbytestotal", "capacity_kilobytes", "Capacity of the pool in kilobytes", s.gaugeMetric},
			{"quota_iused_estimate", "quota_iused_estimate", "Returns '1' if a valid address is returned within the pool, referencing whether free space can be allocated", s.gaugeMetric},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			newMetric := newLustreProcMetric(item.filename, item.promName, "mds", path, item.helpText, item.metricFunc)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
	return nil
}

func (s *lustreSource) generateClientMetricTemplates() error {
	metricMap := map[string][]lustreHelpStruct{
		"llite": {
			{"blocksize", "blocksize_bytes", "Filesystem block size in bytes", s.gaugeMetric},
			{"checksum_pages", "checksum_pages_enabled", "Returns '1' if data checksumming is enabled for the client", s.gaugeMetric},
			{"default_easize", "default_ea_size_bytes", "Default Extended Attribute (EA) size in bytes", s.gaugeMetric},
			{"filesfree", "inodes_free", "The number of inodes (objects) available", s.gaugeMetric},
			{"filestotal", "inodes_maximum", "The maximum number of inodes (objects) the filesystem can hold", s.gaugeMetric},
			{"kbytesavail", "available_kilobytes", "Number of kilobytes readily available in the pool", s.gaugeMetric},
			{"kbytesfree", "free_kilobytes", "Number of kilobytes allocated to the pool", s.gaugeMetric},
			{"kbytestotal", "capacity_kilobytes", "Capacity of the pool in kilobytes", s.gaugeMetric},
			{"lazystatfs", "lazystatfs_enabled", "Returns '1' if lazystatfs (a non-blocking alternative to statfs) is enabled for the client", s.gaugeMetric},
			{"max_easize", "maximum_ea_size_bytes", "Maximum Extended Attribute (EA) size in bytes", s.gaugeMetric},
			{"max_read_ahead_mb", "maximum_read_ahead_megabytes", "Maximum number of megabytes to read ahead", s.gaugeMetric},
			{"max_read_ahead_per_file_mb", "maximum_read_ahead_per_file_megabytes", "Maximum number of megabytes per file to read ahead", s.gaugeMetric},
			{"max_read_ahead_whole_mb", "maximum_read_ahead_whole_megabytes", "Maximum file size in megabytes for a file to be read in its entirety", s.gaugeMetric},
			{"statahead_agl", "statahead_agl_enabled", "Returns '1' if the Asynchronous Glimpse Lock (AGL) for statahead is enabled", s.gaugeMetric},
			{"statahead_max", "statahead_maximum", "Maximum window size for statahead", s.gaugeMetric},
			{"stats", "read_samples_total", readSamplesHelp, s.counterMetric},
			{"stats", "read_minimum_size_bytes", readMinimumHelp, s.gaugeMetric},
			{"stats", "read_maximum_size_bytes", readMaximumHelp, s.gaugeMetric},
			{"stats", "read_bytes_total", readTotalHelp, s.counterMetric},
			{"stats", "write_samples_total", writeSamplesHelp, s.counterMetric},
			{"stats", "write_minimum_size_bytes", writeMaximumHelp, s.gaugeMetric},
			{"stats", "write_maximum_size_bytes", writeMaximumHelp, s.gaugeMetric},
			{"stats", "write_bytes_total", writeTotalHelp, s.counterMetric},
			{"stats", "open_total", openHelp, s.counterMetric},
			{"stats", "close_total", closeHelp, s.counterMetric},
			{"stats", "seek_total", seekHelp, s.counterMetric},
			{"stats", "readdir_total", readdirHelp, s.counterMetric},
			{"stats", "truncate_total", truncateHelp, s.counterMetric},
			{"stats", "getattr_total", getattrHelp, s.counterMetric},
			{"stats", "unlink_total", unlinkHelp, s.counterMetric},
			{"stats", "alloc_inode_total", allocInodeHelp, s.counterMetric},
			{"stats", "setxattr_total", setxattrHelp, s.counterMetric},
			{"stats", "getxattr_total", getxattrHelp, s.counterMetric},
			{"stats", "removexattr_total", removexattrHelp, s.counterMetric},
			{"stats", "inode_permission_total", inodePermissionHelp, s.counterMetric},
			{"xattr_cache", "xattr_cache_enabled", "Returns '1' if extended attribute cache is enabled", s.gaugeMetric},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			newMetric := newLustreProcMetric(item.filename, item.promName, "client", path, item.helpText, item.metricFunc)
			s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
		}
	}
	return nil
}

func (s *lustreSource) generateGenericMetricTemplates() error {
	metricList := []lustreHelpStruct{
		{"health_check", "health_check", "Current health status for the indicated instance: " + healthCheckHealthy + " refers to 'healthy', " + healthCheckUnhealthy + " refers to 'unhealthy'", s.gaugeMetric},
	}
	for _, item := range metricList {
		newMetric := newLustreProcMetric(item.filename, item.promName, "Generic", "", item.helpText, item.metricFunc)
		s.lustreProcMetrics = append(s.lustreProcMetrics, newMetric)
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
	l.generateClientMetricTemplates()
	l.generateGenericMetricTemplates()
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
			case "health_check":
				err = s.parseTextFile(metric.source, "health_check", path, directoryDepth, metric.helpText, metric.promName, func(nodeType string, nodeName string, name string, helpText string, value uint64) {
					ch <- metric.metricFunc([]string{nodeType}, []string{nodeName}, name, helpText, value)
				})
				if err != nil {
					return err
				}
			case "brw_stats":
				err = s.parseBRWStats(metric.source, "brw_stats", path, directoryDepth, metric.helpText, metric.promName, func(nodeType string, brwOperation string, brwSize string, nodeName string, name string, helpText string, value uint64) {
					ch <- metric.metricFunc([]string{nodeType, "operation", "size"}, []string{nodeName, brwOperation, brwSize}, name, helpText, value)
				})
				if err != nil {
					return err
				}
			case "job_stats":
				err = s.parseJobStats(metric.source, "job_stats", path, directoryDepth, metric.helpText, metric.promName, func(nodeType string, jobid string, nodeName string, name string, helpText string, value uint64) {
					ch <- metric.metricFunc([]string{nodeType, "jobid"}, []string{nodeName, jobid}, name, helpText, value)
				})
				if err != nil {
					return err
				}
			default:
				if metric.filename == "stats" {
					metricType = "stats"
				} else if metric.filename == "md_stats" {
					metricType = "md_stats"
				}
				err = s.parseFile(metric.source, metricType, path, directoryDepth, metric.helpText, metric.promName, func(nodeType string, nodeName string, name string, helpText string, value uint64) {
					ch <- metric.metricFunc([]string{nodeType}, []string{nodeName}, name, helpText, value)
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func parseReadWriteBytes(statsFile string, helpText string, promName string) (metricList []lustreStatsMetric, err error) {
	// bytesSplit is in the following format:
	// bytesString: {name} {number of samples} 'samples' [{units}] {minimum} {maximum} {sum}
	// bytesSplit:   [0]    [1]                 [2]       [3]       [4]       [5]       [6]
	bytesMap := map[string]multistatParsingStruct{
		readSamplesHelp:     {pattern: "read_bytes .*", index: 1},
		readMinimumHelp:     {pattern: "read_bytes .*", index: 4},
		readMaximumHelp:     {pattern: "read_bytes .*", index: 5},
		readTotalHelp:       {pattern: "read_bytes .*", index: 6},
		writeSamplesHelp:    {pattern: "write_bytes .*", index: 1},
		writeMinimumHelp:    {pattern: "write_bytes .*", index: 4},
		writeMaximumHelp:    {pattern: "write_bytes .*", index: 5},
		writeTotalHelp:      {pattern: "write_bytes .*", index: 6},
		openHelp:            {pattern: "open .*", index: 1},
		closeHelp:           {pattern: "close .*", index: 1},
		getattrHelp:         {pattern: "getattr .*", index: 1},
		setattrHelp:         {pattern: "setattr .*", index: 1},
		getxattrHelp:        {pattern: "getxattr .*", index: 1},
		setxattrHelp:        {pattern: "setxattr .*", index: 1},
		statfsHelp:          {pattern: "statfs .*", index: 1},
		seekHelp:            {pattern: "seek .*", index: 1},
		readdirHelp:         {pattern: "readdir .*", index: 1},
		truncateHelp:        {pattern: "truncate .*", index: 1},
		allocInodeHelp:      {pattern: "alloc_inode .*", index: 1},
		removexattrHelp:     {pattern: "removexattr .*", index: 1},
		unlinkHelp:          {pattern: "unlink .*", index: 1},
		inodePermissionHelp: {pattern: "inode_permission .*", index: 1},
		createHelp:          {pattern: "create .*", index: 1},
		getInfoHelp:         {pattern: "get_info .*", index: 1},
		setInfoAsyncHelp:    {pattern: "set_info_async .*", index: 1},
		connectHelp:         {pattern: "connect .*", index: 1},
		pingHelp:            {pattern: "ping .*", index: 1},
	}
	bytesString := regexCaptureString(bytesMap[helpText].pattern, statsFile)
	if len(bytesString) < 1 {
		return nil, nil
	}
	r, err := regexp.Compile(" +")
	if err != nil {
		return nil, err
	}
	bytesSplit := r.Split(bytesString, -1)
	result, err := strconv.ParseUint(bytesSplit[bytesMap[helpText].index], 10, 64)
	if err != nil {
		return nil, err
	}
	metricList = append(metricList, lustreStatsMetric{title: promName, help: helpText, value: result})

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

func parseStatsFile(helpText string, promName string, path string) (metricList []lustreStatsMetric, err error) {
	statsFileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	statsFile := string(statsFileBytes[:])

	statsList, err := parseReadWriteBytes(statsFile, helpText, promName)
	if err != nil {
		return nil, err
	}
	if statsList != nil {
		metricList = append(metricList, statsList...)
	}

	return metricList, nil
}

func getJobStatsByOperation(jobBlock string, jobID string, promName string, helpText string) (metricList []lustreJobsMetric, err error) {
	// opMap matches the given helpText value with the placement of the numeric fields within each metric line.
	// For example, the number of samples is the first number in the line and has a helpText of readSamplesHelp,
	// hence the 'index' value of 0. 'pattern' is the regex capture pattern for the desired line.
	opMap := map[string]multistatParsingStruct{
		readSamplesHelp:    {index: 0, pattern: "read_bytes"},
		readMinimumHelp:    {index: 1, pattern: "read_bytes"},
		readMaximumHelp:    {index: 2, pattern: "read_bytes"},
		readTotalHelp:      {index: 3, pattern: "read_bytes"},
		writeSamplesHelp:   {index: 0, pattern: "write_bytes"},
		writeMinimumHelp:   {index: 1, pattern: "write_bytes"},
		writeMaximumHelp:   {index: 2, pattern: "write_bytes"},
		writeTotalHelp:     {index: 3, pattern: "write_bytes"},
		openHelp:           {index: 0, pattern: "open"},
		closeHelp:          {index: 0, pattern: "close"},
		mknodHelp:          {index: 0, pattern: "mknod"},
		linkHelp:           {index: 0, pattern: "link"},
		unlinkHelp:         {index: 0, pattern: "unlink"},
		mkdirHelp:          {index: 0, pattern: "mkdir"},
		rmdirHelp:          {index: 0, pattern: "rmdir"},
		renameHelp:         {index: 0, pattern: "rename"},
		getattrHelp:        {index: 0, pattern: "getattr"},
		setattrHelp:        {index: 0, pattern: "setattr"},
		getxattrHelp:       {index: 0, pattern: "getxattr"},
		setxattrHelp:       {index: 0, pattern: "setxattr"},
		statfsHelp:         {index: 0, pattern: "statfs"},
		syncHelp:           {index: 0, pattern: "sync"},
		samedirRenameHelp:  {index: 0, pattern: "samedir_rename"},
		crossdirRenameHelp: {index: 0, pattern: "crossdir_rename"},
		punchHelp:          {index: 0, pattern: "punch"},
		destroyHelp:        {index: 0, pattern: "destroy"},
		createHelp:         {index: 0, pattern: "create"},
		getInfoHelp:        {index: 0, pattern: "get_info"},
		setInfoHelp:        {index: 0, pattern: "set_info"},
		quotactlHelp:       {index: 0, pattern: "quotactl"},
	}
	// If the metric isn't located in the map, don't try to parse a value for it.
	if _, exists := opMap[helpText]; !exists {
		return nil, nil
	}
	pattern := opMap[helpText].pattern
	opStat := regexCaptureString(pattern+": .*", jobBlock)
	opNumbers := regexCaptureStrings("[0-9]*\\.[0-9]+|[0-9]+", opStat)
	if len(opNumbers) < 1 {
		return nil, nil
	}
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

func getJobNum(jobBlock string) (jobID string, err error) {
	jobID = regexCaptureString("job_id: .*", jobBlock)
	jobID = regexCaptureString("[0-9]*\\.[0-9]+|[0-9]+", jobID)
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
		jobList, err := getJobStatsByOperation(job, jobID, promName, helpText)
		if err != nil {
			return nil, err
		}
        if jobList != nil {
    		metricList = append(metricList, jobList...)
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

func (s *lustreSource) parseTextFile(nodeType string, metricType string, path string, directoryDepth int, helpText string, promName string, handler func(string, string, string, string, uint64)) (err error) {
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
			value, err := strconv.ParseUint(strings.TrimSpace(string(healthCheckHealthy)), 10, 64)
			if err != nil {
				return err
			}
			handler(nodeType, nodeName, promName, helpText, value)
		} else {
			value, err := strconv.ParseUint(strings.TrimSpace(string(healthCheckUnhealthy)), 10, 64)
			if err != nil {
				return err
			}
			handler(nodeType, nodeName, promName, helpText, value)
		}
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
		metricList, err := parseStatsFile(helpText, promName, path)
		if err != nil {
			return err
		}

		for _, metric := range metricList {
			handler(nodeType, nodeName, metric.title, helpText, metric.value)
		}
	case "md_stats":
		metricList, err := parseStatsFile(helpText, promName, path)
		if err != nil {
			return err
		}

		for _, metric := range metricList {
			handler(nodeType, nodeName, metric.title, helpText, metric.value)
		}
	}
	return nil
}

func (s *lustreSource) counterMetric(labels []string, labelValues []string, name string, helpText string, value uint64) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", name),
			helpText,
			labels,
			nil,
		),
		prometheus.CounterValue,
		float64(value),
		labelValues...,
	)
}

func (s *lustreSource) gaugeMetric(labels []string, labelValues []string, name string, helpText string, value uint64) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", name),
			helpText,
			labels,
			nil,
		),
		prometheus.GaugeValue,
		float64(value),
		labelValues...,
	)
}

func (s *lustreSource) untypedMetric(labels []string, labelValues []string, name string, helpText string, value uint64) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", name),
			helpText,
			labels,
			nil,
		),
		prometheus.UntypedValue,
		float64(value),
		labelValues...,
	)
}
