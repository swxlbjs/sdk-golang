package scheduler

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
	"time"
)

func TestScheduler(*testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	s := getExampleScheduler()
	j1 := getExampleJob1()
	j2 := getExampleJob2()
	j3 := getExampleJob3()
	_ = s.AddJob(&j1)
	_ = s.AddJob(&j2)
	go func() {
		time.Sleep(time.Second * 10)
		s.Stop()
	}()
	go func() {
		time.Sleep(time.Second * 2)
		_ = s.AddJob(&j3)
	}()
	s.Start()
}
