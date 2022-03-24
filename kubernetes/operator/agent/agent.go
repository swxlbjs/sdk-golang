package agent

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
	"time"
)

const (
	exporterAddr    = "localhost:9000"
	exporterMetrics = "/metrics"
	exporterHealthy = "/healthy"
)

type Auth struct {
	Get   func(i types.NamespacedName, l logr.Logger, c *client.Client) (interface{}, error)
	Check func(i types.NamespacedName, l logr.Logger, c *client.Client) (interface{}, error)
	Sync  func(i types.NamespacedName, l logr.Logger, c *client.Client, o interface{}, n interface{}) error
}

type Monitor struct {
	Metrics func(i types.NamespacedName, l logr.Logger, c *client.Client) http.Handler
	Healthy func(i types.NamespacedName, l logr.Logger, c *client.Client) http.Handler
}

type Status struct {
	Details func(i types.NamespacedName, l logr.Logger, c *client.Client) error
	Phase   func(i types.NamespacedName, l logr.Logger, c *client.Client) error
}

type Agent struct {
	ID        string
	Ident     types.NamespacedName
	Logs      logr.Logger
	Client    *client.Client
	ClientSet *kubernetes.Clientset
	Interval  int
	Server    http.Server
	Auth      Auth
	Monitor   Monitor
	Status    Status

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// general
// 监控和健康检查
func (a *Agent) general(l logr.Logger, ctx context.Context) {
	a.wg.Add(1)
	defer a.wg.Done()
	mux := http.NewServeMux()
	mux.Handle(exporterMetrics, a.Monitor.Metrics(a.Ident, l, a.Client))
	mux.Handle(exporterHealthy, a.Monitor.Healthy(a.Ident, l, a.Client))
	// TODO
	// Server 可修改
	a.Server = http.Server{Addr: exporterAddr}
	a.Server.Handler = mux
	go func() {
		for {
			select {
			case <-ctx.Done():
				if err := a.Server.Shutdown(context.TODO()); err != nil {
					l.Error(err, "")
				}
				return
			}
		}

	}()
	l.Error(a.Server.ListenAndServe(), "")
}

func (a *Agent) election(l logr.Logger, ctx context.Context) {
	a.wg.Add(1)
	defer a.wg.Done()
	leaderelection.RunOrDie(ctx,
		leaderelection.LeaderElectionConfig{
			Lock: &resourcelock.ConfigMapLock{
				ConfigMapMeta: metav1.ObjectMeta{
					Name:      a.Ident.Name,
					Namespace: a.Ident.Namespace,
				},
				Client: a.ClientSet.CoreV1(),
				LockConfig: resourcelock.ResourceLockConfig{
					Identity: a.ID,
				},
			},
			ReleaseOnCancel: true,
			LeaseDuration:   60 * time.Second, //租约时间
			RenewDeadline:   15 * time.Second, //更新租约的
			RetryPeriod:     5 * time.Second,  //非leader节点重试时间
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(ctx context.Context) {
					a.wg.Add(1)
					defer a.wg.Done()
					ticker := time.NewTicker(time.Duration(a.Interval) * time.Second)
					defer ticker.Stop()
					var err error
					var authGet, authCheck interface{}
					for {
						select {
						case <-ticker.C:
							// auth部分
							if authGet, err = a.Auth.Get(a.Ident, l, a.Client); err != nil {
								l.Error(err, "authGet")
							}
							if authCheck, err = a.Auth.Check(a.Ident, l, a.Client); err != nil {
								l.Error(err, "authCheck")
							}
							if err = a.Auth.Sync(a.Ident, l, a.Client, authGet, authCheck); err != nil {
								l.Error(err, "authSyncErr")
							}
							// status 部分
							if err = a.Status.Details(a.Ident, l, a.Client); err != nil {
								l.Error(err, "statusDetails")
							}
							if err = a.Status.Phase(a.Ident, l, a.Client); err != nil {
								l.Error(err, "statusPhase")
							}
						case <-ctx.Done():
							l.Info("leader stopped")
							return
						}
					}
				},
				OnStoppedLeading: func() {
					l.Info("old leader is lost")
				},
				OnNewLeader: func(identity string) {
					if identity == a.ID {
						l.Info("new leader is myself")
					} else {
						l.Info("new leader elected: %s", identity)
					}
				},
			},
		},
	)
}

func (a *Agent) Start() {
	var ctx context.Context
	ctx, a.cancel = context.WithCancel(context.Background())
	defer a.cancel()
	// 定义基础
	defer a.wg.Wait()
	// exporter协程后台运行
	go a.general(a.Logs.WithName("general"), ctx)
	// 主从选举，主程序运行鉴权和状态更新
	a.election(a.Logs.WithName("election"), ctx)
}
func (a *Agent) Background() {
	go a.Start()
}

func (a *Agent) Stop() {
	//defer a.wg.Wait()
	a.cancel()
}

func NewAgent() *Agent {
	return &Agent{
		ID:   uuid.New().String(),
		Logs: ctrl.Log.WithName("agent"),
		wg:   sync.WaitGroup{},
	}
}
