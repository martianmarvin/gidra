package sleep

import (
	"errors"
	"time"

	"github.com/martianmarvin/gidra/task"
)

// Sleeps for specified number of seconds
// Required Parameters:
// - seconds number of seconds to sleep

func init() {
	task.Register("sleep", NewSleepTask)
}

type SleepTask struct {
}

func NewSleepTask() task.Task {
	return &SleepTask{}
}

func parseSeconds(v interface{}) time.Duration {
	switch n := v.(type) {
	case time.Duration:
		return n * time.Second
	case int:
		return time.Duration(n) * time.Second
	case int32:
		return time.Duration(n) * time.Second
	case int64:
		return time.Duration(n) * time.Second
	case string:
		d, err := time.ParseDuration(n)
		if err != nil {
			return time.Duration(0)
		}
		if d > time.Second {
			return d
		} else {
			return d * time.Second
		}
	default:
		return time.Duration(0)

	}
}

func (t *SleepTask) Execute(vars map[string]interface{}) (err error) {
	if v, ok := vars["seconds"]; !ok {
		err = errors.New("Missing required parameter: seconds")
		return
	} else {
		time.Sleep(parseSeconds(v))
	}

	return err
}
