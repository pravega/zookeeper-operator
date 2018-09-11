package main

import (
	"context"
	"runtime"
	"time"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/pravega/zookeeper-operator/pkg/stub"

	"github.com/sirupsen/logrus"
	"os"
)

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
}

func main() {
	printVersion()

	resource := "zookeeper.pravega.io/v1beta1"
	kind := "ZookeeperCluster"
	namespace := getWatchNamespaceAllowBlank()
	resyncPeriod := 5
	logrus.Infof("Watching %s, %s, %s, %d", resource, kind, namespace, resyncPeriod)
	sdk.Watch(resource, kind, namespace, time.Duration(resyncPeriod)*time.Second)
	sdk.Handle(stub.NewHandler())
	sdk.Run(context.TODO())
}

// GetWatchNamespaceAllowBlank returns the namespace the operator should be watching for changes
func getWatchNamespaceAllowBlank() string {
	ns, found := os.LookupEnv(k8sutil.WatchNamespaceEnvVar)
	if !found {
		logrus.Infof("%s is not set, watching all namespaces", k8sutil.WatchNamespaceEnvVar)
		ns = ""
	}
	return ns
}
