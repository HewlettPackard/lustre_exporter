package sources

import (
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/gammazero/workerpool"
)

type fileReader struct {
	files              map[string][]byte
	pathGlobs          map[string][]string
	pool               *workerpool.WorkerPool
	mu                 sync.Locker
	wg                 *sync.WaitGroup
}

func newFileReader() *fileReader{
	fr := &fileReader{
		files        : map[string][]byte{},
		pathGlobs    : map[string][]string{},
		mu           : &sync.Mutex{},
		wg           : &sync.WaitGroup{},
	}

	fr.pool = workerpool.New(8)

	return fr
}

func (fr *fileReader)readFileAsync(path string) {
	if fr.pool == nil {
		return 
	}

	fr.wg.Add(1)
	fn := func() {
		data, err := ioutil.ReadFile(filepath.Clean(path))
		if err == nil {
			fr.mu.Lock()
			defer fr.mu.Unlock()
			fr.files[path] = data
		}

		fr.wg.Done()
	}
	fr.pool.Submit(fn)
}

func (fr *fileReader)wait(release ...bool) {
	fr.wg.Wait()

	if len(release) > 0 && release[0] {
		fr.pool.StopWait()
		fr.pool = nil
	}
}

func (fr *fileReader)readFile(path string) ([]byte, error) {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	data, ok := fr.files[path]
	if ok && data != nil{
	  //log.Infof("cache meet: %s\n", path)
		return data, nil
	}

	//log.Infof("cache miss: %s\n", path)
	data, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return data, err
	}

	fr.files[path] = data
	return data, nil
}

func (fr *fileReader)glob(path string, read...bool) (paths []string, err error) {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	paths, ok := fr.pathGlobs[path]
	if !ok {
		paths, err = filepath.Glob(path)
		if err != nil {
			return nil, err
		}
		fr.pathGlobs[path] = paths
	}

	if len(read) > 0 && read[0] {
		for _, path := range paths {
			if _, ok := fr.files[path]; !ok {
				fr.files[path] = nil
				fr.readFileAsync(path)
			}
		}
	}

	return paths, nil
}

func (fr *fileReader)release(){
	fr.files = nil
	fr.pathGlobs = nil
}