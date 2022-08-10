package sources

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"lustre_exporter/log"

	"github.com/prometheus/client_golang/prometheus"
)

type procfsV2 struct {
	reNum *regexp.Regexp

}

type procfsV2Ctx struct {
  s                  *lustreProcfsSource
	fr                 *fileReader
	filesJobStats      map[string]map[string]map[string][]int64
	lastover           time.Time
	metrics_           []prometheus.Metric
}

var insProcfsV2 = &procfsV2{}

func (v2 *procfsV2)newCtx(s  *lustreProcfsSource) *procfsV2Ctx{
	return &procfsV2Ctx{
	  s            : s,
		fr           : newFileReader(),
		filesJobStats: map[string]map[string]map[string][]int64{},
	}
}


func (ctx *procfsV2Ctx)update(ch chan<- prometheus.Metric) {
	for _, m := range ctx.metrics_ {
		ch <- m
	}
}

func (ctx *procfsV2Ctx)release()  {
	ctx.filesJobStats = nil
	ctx.fr.release()
}

func (ctx *procfsV2Ctx)prepareFiles() (err error) {
	for _, metric := range ctx.s.lustreProcMetrics {
		_, err := ctx.fr.glob(filepath.Join(ctx.s.basePath, metric.path, metric.filename), true)
		if err != nil {
			return err
		}
	}
	ctx.fr.wait(true)

	return nil
}

func (ctx *procfsV2Ctx)collect() error {
	var metricType string
	var directoryDepth int

	s := ctx.s

	ctx.prepareFiles()

	var metrics []prometheus.Metric

	for _, metric := range s.lustreProcMetrics {
		directoryDepth = strings.Count(metric.filename, "/")
		paths, err := ctx.fr.glob(filepath.Join(s.basePath, metric.path, metric.filename))
		if err != nil {
			return err
		}
		if paths == nil {
			continue
		}
		for _, path := range paths {
			metricType = single
			switch metric.filename {
			case "brw_stats", "rpc_stats":
				err = ctx.parseBRWStats(metric.source, "stats", path, directoryDepth, metric.helpText, metric.promName, metric.hasMultipleVals, func(nodeType string, brwOperation string, brwSize string, nodeName string, name string, helpText string, value float64, extraLabel string, extraLabelValue string) {
					if extraLabelValue == "" {
						metrics = append(metrics, metric.metricFunc([]string{"component", "target", "operation", "size"}, []string{nodeType, nodeName, brwOperation, brwSize}, name, helpText, value))
					} else {
						metrics = append(metrics, metric.metricFunc([]string{"component", "target", "operation", "size", extraLabel}, []string{nodeType, nodeName, brwOperation, brwSize, extraLabelValue}, name, helpText, value))
					}
				})
				if err != nil {
					return err
				}
			case "job_stats":
				err = ctx.parseJobStats(metric.source, "job_stats", path, directoryDepth, metric.helpText, metric.promName, metric.hasMultipleVals, func(nodeType string, jobid string, nodeName string, name string, helpText string, value float64, extraLabel string, extraLabelValue string) {
					if extraLabelValue == "" {
						metrics = append(metrics, metric.metricFunc([]string{"component", "target", "jobid"}, []string{nodeType, nodeName, jobid}, name, helpText, value))
					} else {
						metrics = append(metrics, metric.metricFunc([]string{"component", "target", "jobid", extraLabel}, []string{nodeType, nodeName, jobid, extraLabelValue}, name, helpText, value))
					}
				})
				if err != nil {
					return err
				}
			default:
				if metric.filename == stats {
					metricType = stats
				} else if metric.filename == mdStats {
					metricType = mdStats
				} else if metric.filename == encryptPagePools {
					metricType = encryptPagePools
				}
				err = ctx.parseFile(metric.source, metricType, path, directoryDepth, metric.helpText, metric.promName, metric.hasMultipleVals, func(nodeType string, nodeName string, name string, helpText string, value float64, extraLabel string, extraLabelValue string) {
					if extraLabelValue == "" {
						metrics = append(metrics, metric.metricFunc([]string{"component", "target"}, []string{nodeType, nodeName}, name, helpText, value))
					} else {
						metrics = append(metrics, metric.metricFunc([]string{"component", "target", extraLabel}, []string{nodeType, nodeName, extraLabelValue}, name, helpText, value))
					}
				})
				if err != nil {
					return err
				}
			}
		}
	}

	ctx.metrics_ = metrics

	return nil
}

var	brwStatsMetricBlocks = map[string]string{
		pagesPerBlockRWHelp:    "pages per bulk r/w",
		discontiguousPagesHelp: "discontiguous pages",
		diskIOsInFlightHelp:    "disk I/Os in flight",
		ioTimeHelp:             "I/O time",
		diskIOSizeHelp:         "disk I/O size",
		pagesPerRPCHelp:        "pages per rpc",
		rpcsInFlightHelp:       "rpcs in flight",
		offsetHelp:             "offset",
	}

func (ctx *procfsV2Ctx) parseBRWStats(nodeType string, metricType string, path string, directoryDepth int, helpText string, promName string, hasMultipleVals bool, handler func(string, string, string, string, string, string, float64, string, string)) (err error) {
	_, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}

	statsFileBytes, err := ctx.fr.readFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	statsFile := string(statsFileBytes[:])
	block := regexCaptureString("(?ms:^"+brwStatsMetricBlocks[helpText]+".*?(\n\n|\\z))", statsFile)
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

func (ctx *procfsV2Ctx) parseFile(nodeType string, metricType string, path string, directoryDepth int, helpText string, promName string, hasMultipleVals bool, handler func(string, string, string, string, float64, string, string)) (err error) {
	_, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}
	switch metricType {
	case single:
		value, err := ctx.fr.readFile(filepath.Clean(path))
		if err != nil {
			return err
		}
		convertedValue, err := strconv.ParseFloat(strings.TrimSpace(string(value)), 64)
		if err != nil {
			return err
		}
		handler(nodeType, nodeName, promName, helpText, convertedValue, "", "")
	case stats, mdStats, encryptPagePools:
		metricList, err := ctx.parseStatsFile(helpText, promName, path, hasMultipleVals)
		if err != nil {
			return err
		}

		for _, metric := range metricList {
			handler(nodeType, nodeName, metric.title, metric.help, metric.value, metric.extraLabel, metric.extraLabelValue)
		}
	}
	return nil
}

func (ctx *procfsV2Ctx)parseStatsFile(helpText string, promName string, path string, hasMultipleVals bool) (metricList []lustreStatsMetric, err error) {
	statsFileBytes, err := ctx.fr.readFile(filepath.Clean(path))
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

func (ctx *procfsV2Ctx) parseJobStats(nodeType string, metricType string, path string, directoryDepth int, helpText string, promName string, hasMultipleVals bool, handler func(string, string, string, string, string, float64, string, string)) (err error) {
	_, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}

	metricList, err := ctx.parseJobStatsText(path, promName, helpText, hasMultipleVals)
	if err != nil {
		return err
	}

	for _, item := range metricList {
		handler(nodeType, item.jobID, nodeName, item.lustreStatsMetric.title, item.lustreStatsMetric.help, item.lustreStatsMetric.value, item.lustreStatsMetric.extraLabel, item.lustreStatsMetric.extraLabelValue)
	}
	return nil
}

func (ctx *procfsV2Ctx)parseJobStatsText(path string, promName string, helpText string, hasMultipleVals bool) (metricList []lustreJobsMetric, err error){

	var jobList []lustreJobsMetric

	jobsStats, ok := ctx.filesJobStats[path]
	if !ok {
		jobStatsBytes, err := ctx.fr.readFile(path)
		if err != nil {
			return nil, err
		}
		jobStatsContent := string(jobStatsBytes[:])
		splits := strings.Split(jobStatsContent, "- ")
		if len(splits) <= 1{
			return nil, nil
		}
		jobs := splits[1:]

		jobsStats = map[string]map[string][]int64{}
		for _, job := range jobs {
			jobid, data, err := ctx.parsingJobStats(job)
			if err != nil {
				log.Warnf("parsing jobstat failed: %s", err)
				continue
			}
			jobsStats[jobid] = data
		}
		ctx.filesJobStats[path] = jobsStats
	}

	if hasMultipleVals {
		jobList, err = ctx.getJobStatsOperationMetrics(jobsStats, promName, helpText)
	} else {
		jobList, err = ctx.getJobStatsIOMetrics(jobsStats, promName, helpText)
	}
	if jobList != nil {
		metricList = append(metricList, jobList...)
	}
	return metricList, nil
}

func (ctx *procfsV2Ctx)getJobID(line string) (string, error) {

	idx := strings.Index(line, "job_id:")
	if idx < 0 {
		return "", fmt.Errorf("not found")
	}

	idx2 := strings.Index(line[idx:], "\n")
	if idx2 < 0 {
		return getValidUtf8String(strings.TrimSpace(line[idx+len("job_id:"):])), nil
	}
	return getValidUtf8String(strings.TrimSpace(line[idx+len("job_id:"):idx2])), nil
}

func (ctx *procfsV2Ctx)parsingJobStats(job string) (string, map[string][]int64, error) {

	lines := strings.Split(job, "\n")

	if len(lines) < 1 {
		return "", nil, fmt.Errorf("invalid content for parsing jobStat: %s", job)
	}

	jobid, err := ctx.getJobID(lines[0])
	if err != nil {
		return "", nil, fmt.Errorf("can not found jobid in content: %s", job)
	}

	j := map[string][]int64{}
	for _, line := range lines[1:] {

		idx := strings.Index(line, ":")
		if idx < 1{
			continue
		}

		key := strings.TrimSpace(line[:idx])
		numStrs := insProcfsV2.reNum.FindAllString(line[idx:], -1)
		nums := make([]int64, 0, 1)
		for _, numStr := range numStrs {
			num, err := strconv.ParseInt(strings.TrimSpace(numStr), 10, 64)
			if err != nil {
				continue
			}
			nums = append(nums, num)
		}

		j[key] = nums
	} 

	return jobid, j, nil
}

var jobStatsOperationSlice []multistatParsingStruct = []multistatParsingStruct{
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

func (ctx *procfsV2Ctx)getJobStatsOperationMetrics(jobsStats map[string]map[string][]int64, promName string, helpText string) (metricList []lustreJobsMetric, err error) {
	for jobid, stats := range jobsStats {
		for _, operation := range jobStatsOperationSlice {
			vals, ok := stats[operation.pattern]
			if !ok || len(vals) <= operation.index {
				continue
			}
			l := lustreStatsMetric{
				title:           promName,
				help:            helpText,
				value:           float64(vals[operation.index]),
				extraLabel:      "operation",
				extraLabelValue: operation.pattern,
			}
			metricList = append(metricList, lustreJobsMetric{jobid, l})
		}
	}

	return metricList, err
}

var jobStatsMultistatParsingStruct map[string]multistatParsingStruct = map[string]multistatParsingStruct{
		readSamplesHelp:  {index: 0, pattern: "read_bytes"},
		readMinimumHelp:  {index: 1, pattern: "read_bytes"},
		readMaximumHelp:  {index: 2, pattern: "read_bytes"},
		readTotalHelp:    {index: 3, pattern: "read_bytes"},
		writeSamplesHelp: {index: 0, pattern: "write_bytes"},
		writeMinimumHelp: {index: 1, pattern: "write_bytes"},
		writeMaximumHelp: {index: 2, pattern: "write_bytes"},
		writeTotalHelp:   {index: 3, pattern: "write_bytes"},
	}

func (ctx *procfsV2Ctx)getJobStatsIOMetrics(jobsStats map[string]map[string][]int64, promName string, helpText string) (metricList []lustreJobsMetric, err error){
	// opMap matches the given helpText value with the placement of the numeric fields within each metric line.
	// For example, the number of samples is the first number in the line and has a helpText of readSamplesHelp,
	// hence the 'index' value of 0. 'pattern' is the regex capture pattern for the desired line.
	opMap := jobStatsMultistatParsingStruct
	// If the metric isn't located in the map, don't try to parse a value for it.
	if _, exists := opMap[helpText]; !exists {
		return nil, nil
	}

	operation := opMap[helpText]
	for jobid, stats := range jobsStats {
		vals, ok := stats[operation.pattern]
		if !ok || len(vals) <= operation.index {
			continue
		}

		l := lustreStatsMetric{
			title:           promName,
			help:            helpText,
			value:           float64(vals[operation.index]),
			extraLabel:      "",
			extraLabelValue: "",
		}
		metricList = append(metricList, lustreJobsMetric{jobid, l})
	}

	return 
}

func init(){
	insProcfsV2.reNum = regexp.MustCompile("[0-9]*\\.[0-9]+|[0-9]+")
}