package scheduler

import (
	"context"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sort"
	"sync"
	"time"
)

type Scheduler struct {
	// 运行的任务列表
	jobs []*Job
	// 待添加到运行的任务列表
	jobAdd chan *Job
	// 待停止运行的任务列表
	jobDel chan string
	// 操作锁，读写
	jobMu sync.RWMutex
	// 运行锁
	runMu sync.Mutex
	// 运行状态
	runStatus bool
	// 运行任务协程计数器
	jobWg sync.WaitGroup
	// 立即停止
	isStop context.CancelFunc
	// 日志
	Logs logr.Logger
}

func (s *Scheduler) now() time.Time {
	return time.Now().In(time.Local)
}

func (s *Scheduler) AddJob(j *Job) error {
	s.jobMu.Lock()
	defer s.jobMu.Unlock()
	j.Next = s.now()
	if !s.runStatus {
		s.jobs = append(s.jobs, j)
	} else {
		s.jobAdd <- j
	}
	s.Logs.Info("add", "job", j.Name)
	return nil
}
func (s *Scheduler) ListJob() []Job {
	var jobs []Job
	for _, j := range s.jobs {
		jobs = append(jobs, *j)
	}
	return jobs
}

func (s *Scheduler) startJob(j Job) {
	s.jobWg.Add(1)
	go func() {
		defer s.jobWg.Done()
		j.CMD()
	}()
}

func (s *Scheduler) run(ctx context.Context, l logr.Logger) {
	// 定义启动时运行时间
	now := s.now()
	for {
		// job 按next时间排序
		sort.Sort(sortJobs(s.jobs))
		var timer *time.Timer
		if len(s.jobs) == 0 || s.jobs[0].Next.IsZero() {
			time.Sleep(time.Second * 5)
			continue
		} else {
			timer = time.NewTimer(s.jobs[0].Next.Sub(now))
		}
		for {
			select {
			case <-ctx.Done():
				timer.Stop()
				l.Info("stop")
				return
			case now = <-timer.C:
				now = s.now()
				for _, j := range s.jobs {
					if j.Next.After(now) || j.Next.IsZero() {
						break
					}
					s.startJob(*j)
					j.Prev = now
					j.Next = j.Prev.Add(j.Interval)
					l.Info("run", "Now runtime", now, "job", j.Name, "Next runtime", j.Next)
				}
			case jobNew := <-s.jobAdd:
				timer.Stop()
				now = s.now()
				jobNew.Next = now.Add(jobNew.Interval)
				s.jobs = append(s.jobs, jobNew)
				l.Info("add", "Now runtime", now, "job", jobNew.Name, "Next runtime", jobNew.Next)
			case jobName := <-s.jobDel:
				timer.Stop()
				now = s.now()
				var jobs []*Job
				for _, j := range s.jobs {
					if j.Name != jobName {
						jobs = append(jobs, j)
					}
				}
				l.Info("del", "job", jobName)
			}
			break
		}
	}
}

func (s *Scheduler) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.runMu.Lock()
	s.isStop = cancel
	if s.runStatus {
		s.runMu.Unlock()
		return
	}
	s.runStatus = true
	s.runMu.Unlock()
	s.run(ctx, s.Logs.WithName("run"))
}

func (s *Scheduler) Stop() {
	s.runMu.Lock()
	defer s.runMu.Unlock()
	if s.runStatus {
		s.runStatus = false
	}
	go func() {
		s.jobWg.Wait()
		s.isStop()
	}()
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		jobs:      nil,
		jobAdd:    make(chan *Job),
		jobDel:    make(chan string),
		jobMu:     sync.RWMutex{},
		runMu:     sync.Mutex{},
		runStatus: false,
		jobWg:     sync.WaitGroup{},
		isStop:    nil,
		Logs:      ctrl.Log.WithName("scheduler"),
	}
}
