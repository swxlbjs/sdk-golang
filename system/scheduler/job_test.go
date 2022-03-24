package scheduler

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sort"
	"testing"
	"time"
)

func TestJob(*testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	var l = ctrl.Log.WithName("TestJob")
	j1 := getExampleJob1()
	j2 := getExampleJob2()
	j3 := getExampleJob3()
	j1.Next = time.Now().Add(time.Second * 10)
	j2.Next = time.Now().Add(time.Second * 7)
	j3.Next = time.Now().Add(time.Second * 4)
	jobs := []*Job{&j1, &j2, &j3}
	sort.Sort(sortJobs(jobs))
	for _, j := range jobs {
		l.Info("", "job", j.Name)
	}
}
