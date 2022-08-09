package sources

import (
	"lustre_exporter/log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var MAX_WORKER = 4

type  runner struct {
	mu          sync.Mutex
	workers     map[*worker]*worker
	lastSuccess *worker
}

type worker struct {
  r       *runner
  wg      sync.WaitGroup
  mu      sync.Mutex
  list    map[string]LustreSource
	ctxs    []collectorCtx
	creat   time.Time
	start   time.Time
	end     time.Time
	results []*result
}

type result struct {
	name   string
	result string
	cost   time.Duration
}

func newWorker(list map[string]LustreSource, r *runner) *worker{
	return &worker{
	  r   : r,
		list: list,
		creat: time.Now(),
	}
}

func (w *worker)run() {
  w.start = time.Now()
	w.wg.Add(len(w.list))
  go func() {
		for name, c := range w.list {
			ctx := c.newCtx()
			w.ctxs = append(w.ctxs, ctx)
			go func(name string, c LustreSource) {
				err := ctx.collect()
				duration := time.Since(w.start)
				if err != nil {
					log.Errorf("ERROR: %q source failed after %f seconds: %s", name, duration.Seconds(), err)
					w.results = append(w.results, &result{name: name, result: "error", cost: duration })
				} else {
					w.results = append(w.results, &result{name: name, result: "success", cost: duration })
				}
				w.wg.Done()
			}(name, c)
		}
		w.wg.Wait()
		w.end = time.Now()
		w.r.doneForWorker(w)
	}()
}

func (w *worker)wait() {
	w.wg.Wait()
}

func (w *worker)release() {
	for _, ctx := range w.ctxs {
		ctx.release()
	}
}

func (w *worker)update(sv *prometheus.SummaryVec, ch chan<- prometheus.Metric) {
	for _, ctx := range w.ctxs {
		ctx.update(ch)
	}

	for _, result := range w.results {
		sv.WithLabelValues(result.name, result.result).Observe(result.cost.Seconds())
	}
	sv.Collect(ch)
}

var insRunner = &runner{
	workers: map[*worker]*worker{},
}

func Runner() *runner{
	return insRunner
}

func (r *runner)getRunnerWorker(list map[string]LustreSource) (ret *worker) {

	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.workers) >= MAX_WORKER{
		for _, worker := range r.workers {
			if ret == nil {
				ret = worker
			} else if ret.creat.Before(worker.creat){
				ret = worker
			}
		}
		return
	}

	ret = newWorker(list, r)
	r.workers[ret] = ret

	ret.run()

	return
}

func (r *runner)doneForWorker(w *worker){
	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastSuccess = w
	delete(r.workers, w)

	go w.release()
}

func (r *runner)Update(list map[string]LustreSource, sv *prometheus.SummaryVec, ch chan<- prometheus.Metric){

	if CollectVersion == "v2" {
		r.updateV2(list, sv, ch)
		return
	}

	r.updateV1(list, sv, ch)
}

func (r *runner)updateV2(list map[string]LustreSource, sv *prometheus.SummaryVec, ch chan<- prometheus.Metric){

	now := time.Now()
	lastSuccess := r.lastSuccess
	if lastSuccess != nil {
	  // return directly if collect over in 1 second
		if now.Sub(lastSuccess.end) <= time.Second {
			lastSuccess.update(sv, ch)
			return
		}
	}

	w := r.getRunnerWorker(list)
	w.wait()
	w.update(sv, ch)
}

func (r *runner)updateV1(list map[string]LustreSource, sv *prometheus.SummaryVec, ch chan<- prometheus.Metric){
	wg := sync.WaitGroup{}
	wg.Add(len(list))
	for name, c := range list {
		go func(name string, c LustreSource) {

			result := "success"
			begin := time.Now()
			err := c.Update(ch)
			duration := time.Since(begin)
			if err != nil {
				log.Errorf("ERROR: %q source failed after %f seconds: %s", name, duration.Seconds(), err)
				result = "error"
			} else {
				log.Debugf("OK: %q source succeeded after %f seconds: %s", name, duration.Seconds(), err)
			}
			sv.WithLabelValues(name, result).Observe(duration.Seconds())


			wg.Done()
		}(name, c)
	}
	wg.Wait()
	sv.Collect(ch)
}
