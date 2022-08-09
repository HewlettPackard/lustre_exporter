package sources

import (
	"fmt"
	"testing"
)

func TestParsingJobStats(t *testing.T){

  jobstr := `job_id:          kworker/14:1.0
	snapshot_time:   1652255649
	read_bytes:      { samples:           0, unit: bytes, min:       0, max:       0, sum:               0 }
	write_bytes:     { samples:           0, unit: bytes, min:       0, max:       0, sum:               0 }
	getattr:         { samples:           0, unit:  reqs }
	setattr:         { samples:           0, unit:  reqs }
	punch:           { samples:           0, unit:  reqs }
	sync:            { samples:           0, unit:  reqs }
	destroy:         { samples:           0, unit:  reqs }
	create:          { samples:           0, unit:  reqs }
	statfs:          { samples:           0, unit:  reqs }
	get_info:        { samples:           0, unit:  reqs }
	set_info:        { samples:         286, unit:  reqs }
	quotactl:        { samples:           0, unit:  reqs }`

	jobid, data, _ := insProcfsV2.newCtx(nil).parsingJobStats(jobstr)
	fmt.Printf("%s: %v\n", jobid, data)
}