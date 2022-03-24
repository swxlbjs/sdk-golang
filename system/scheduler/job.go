package scheduler

import "time"

type sortJobs []*Job

func (s sortJobs) Len() int      { return len(s) }
func (s sortJobs) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortJobs) Less(i, j int) bool {
	if s[i].Next.IsZero() {
		return false
	}
	if s[j].Next.IsZero() {
		return true
	}
	return s[i].Next.Before(s[j].Next)
}

type Job struct {
	// 任务名称
	Name string
	// 执行间隔
	Interval time.Duration
	// 下次执行时间
	Next time.Time
	// 上次执行时间
	Prev time.Time
	// 执行函数
	CMD func()
}
