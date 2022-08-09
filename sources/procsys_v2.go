package sources

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type procsysV2 struct {
}

type procsysV2Ctx struct {
	s                  *lustreProcsysSource
	fr                 *fileReader
	lastover           time.Time
	metrics            []prometheus.Metric
}

var insProcsysV2 = &procsysV2{}

func (v2 *procsysV2)newCtx(s *lustreProcsysSource) *procsysV2Ctx{
	return &procsysV2Ctx{
	  s            : s,
		fr           : newFileReader(),
	}
}

func (ctx *procsysV2Ctx)release(){
	ctx.fr.release()
}

func (ctx *procsysV2Ctx)update(ch chan<- prometheus.Metric) {
	for _, m := range ctx.metrics {
		ch <- m
	}
}

func (ctx *procsysV2Ctx)prepareFiles() (err error) {
	for _, metric := range ctx.s.lustreProcMetrics {
		_, err := ctx.fr.glob(filepath.Join(ctx.s.basePath, metric.path, metric.filename), true)
		if err != nil {
			return err
		}
	}

	ctx.fr.wait(true)

	return nil
}

func (ctx *procsysV2Ctx)collect() (err error){
	var metricType string

	s := ctx.s
	ctx.prepareFiles()

	for _, metric := range s.lustreProcMetrics {
		paths, err := ctx.fr.glob(filepath.Join(s.basePath, metric.path, metric.filename))
		if err != nil {
			return err
		}
		if paths == nil {
			continue
		}
		for _, path := range paths {
			metricType = single
			if metric.filename == stats {
				metricType = stats
			}
			err = ctx.parseFile(metric.source, metricType, path, metric.helpText, metric.promName, func(nodeType string, nodeName string, name string, helpText string, value float64) {
				ctx.metrics = append(ctx.metrics, metric.metricFunc([]string{"component", "target"}, []string{nodeType, nodeName}, name, helpText, value))
			})
			if err != nil {
				return err
			}
		}
	}

	ctx.lastover = time.Now()

	return nil
}

func (ctx *procsysV2Ctx)Update(ch chan<- prometheus.Metric) (err error) {
	for _, metric := range ctx.metrics {
		ch <- metric
	}
	return nil
}

func (ctx *procsysV2Ctx) parseFile(nodeType string, metricType string, path string, helpText string, promName string, handler func(string, string, string, string, float64)) (err error) {
	_, nodeName, err := parseFileElements(path, 0)
	if err != nil {
		return err
	}
	switch metricType {
	case single:
		value, err := ctx.fr.readFile(path)
		if err != nil {
			return err
		}
		convertedValue, err := strconv.ParseFloat(strings.TrimSpace(string(value)), 64)
		if err != nil {
			return err
		}
		handler(nodeType, nodeName, promName, helpText, convertedValue)
	case stats:
		statsFileBytes, err := ctx.fr.readFile(path)
		if err != nil {
			return err
		}
		statsFile := string(statsFileBytes[:])
		metric, err := parseSysStatsFile(helpText, promName, statsFile)
		if err != nil {
			return err
		}
		handler(nodeType, nodeName, metric.title, helpText, metric.value)
	}
	return nil
}