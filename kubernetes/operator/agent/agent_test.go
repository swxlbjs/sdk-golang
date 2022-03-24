package agent

import (
	"github.com/swxlbjs/sdk-golang/kubernetes/operator/common"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
	"time"
)

var schema = runtime.NewScheme()

const (
	name      = "test"
	namespace = "kube-system"
)

func init() {
	testing.Init()
	os.Setenv("Name", "test")
	os.Setenv("Namespace", "default")
	os.Setenv("Ident", "doe")
	_ = clientgoscheme.AddToScheme(schema)

}

func TestAgentRunner(*testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	cfg := ctrl.GetConfigOrDie()
	auth := Auth{
		Get:   exampleAuthGet,
		Check: exampleAuthCheck,
		Sync:  exampleAuthSync,
	}

	monitor := Monitor{
		Metrics: exampleMonitorExporterGet,
		Healthy: exampleMonitorHealthyGet,
	}
	status := Status{
		Details: exampleStatusDetails,
		Phase:   exampleStatusPhase,
	}
	var cl client.Client
	var cs *kubernetes.Clientset
	var err error
	if cl, err = common.GetClient(cfg, schema); err != nil {
		return
	}
	if cs, err = common.GetClientSet(cfg); err != nil {
		return
	}
	a := NewAgent()
	//a.ID = "tt"
	a.Ident = types.NamespacedName{Name: name, Namespace: namespace}
	a.Interval = 5
	a.Client = &cl
	a.ClientSet = cs
	a.Auth = auth
	a.Monitor = monitor
	a.Status = status
	//a.Server = http.Server{}
	go func() {
		time.Sleep(7 * time.Second)
		a.Stop()
	}()
	a.Start()
	time.Sleep(10 * time.Second)

	a.Background()
	time.Sleep(30 * time.Second)
	a.Stop()

}
