package agent

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Auth
func exampleAuthGet(i types.NamespacedName, l logr.Logger, c *client.Client) (interface{}, error) {
	l = l.WithName("Auth").WithName("Get")
	l.Info("test")
	return "", nil
}
func exampleAuthCheck(i types.NamespacedName, l logr.Logger, c *client.Client) (interface{}, error) {
	l = l.WithName("Auth").WithName("Check")
	l.Info("test")
	return "", nil
}
func exampleAuthSync(i types.NamespacedName, l logr.Logger, c *client.Client, i1 interface{}, i2 interface{}) error {
	l = l.WithName("Auth").WithName("Sync")
	l.Info("test")
	return nil
}

// Monitor

func exampleMonitorExporterGet(i types.NamespacedName, l logr.Logger, c *client.Client) http.Handler {
	return exampleMonitorExporter{}
}

type exampleMonitorExporter struct {
}

func (i exampleMonitorExporter) ServeHTTP(http.ResponseWriter, *http.Request) {
}

func exampleMonitorHealthyGet(i types.NamespacedName, l logr.Logger, c *client.Client) http.Handler {
	return exampleMonitorHealthy{}
}

type exampleMonitorHealthy struct {
}

func (i exampleMonitorHealthy) ServeHTTP(http.ResponseWriter, *http.Request) {
}

// Status

func exampleStatusDetails(i types.NamespacedName, l logr.Logger, c *client.Client) error {
	l = l.WithName("exampleMonitor").WithName("Details")
	l.Info("test")
	return nil
}
func exampleStatusPhase(i types.NamespacedName, l logr.Logger, c *client.Client) error {
	l = l.WithName("exampleMonitor").WithName("Phase")
	l.Info("test")
	return nil
}
