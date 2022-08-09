package sources

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type sysfsV2 struct {
}

type sysfsV2Ctx struct {
	s                  *lustreSysSource
	fr                 *fileReader
	lastover           time.Time
	metrics            []prometheus.Metric
}

var insSysfsV2 = &sysfsV2{}

func (v2 *sysfsV2)newCtx(s  *lustreSysSource) *sysfsV2Ctx{
	return &sysfsV2Ctx{
	  s            : s,
		fr           : newFileReader(),
	}
}

func (ctx *sysfsV2Ctx)release(){
	ctx.fr.release()
}

func (ctx *sysfsV2Ctx)update(ch chan<- prometheus.Metric) {
	for _, m := range ctx.metrics {
		ch <- m
	}
}

func (ctx *sysfsV2Ctx)collect() (err error){
	var directoryDepth int
	
	s := ctx.s
	var metrics []prometheus.Metric

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
				err = s.parseTextFile(metric.source, "health_check", path, directoryDepth, metric.helpText, metric.promName, func(nodeType string, nodeName string, name string, helpText string, value float64) {
					metrics = append(metrics, metric.metricFunc([]string{"component", "target"}, []string{nodeType, nodeName}, name, helpText, value))
				})
				if err != nil {
					return err
				}
			}
		}
	}

	ctx.metrics = metrics

	return nil
}

func (c *sysfsV2Ctx) parseTextFile(nodeType string, metricType string, path string, directoryDepth int, helpText string, promName string, handler func(string, string, string, string, float64)) (err error) {
	filename, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}
	fileBytes, err := c.fr.readFile(path)
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