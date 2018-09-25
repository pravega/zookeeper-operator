package main

import (
	"context"
	"flag"
	"fmt"
	"runtime"
	"time"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/pravega/zookeeper-operator/pkg/stub"
	"github.com/pravega/zookeeper-operator/pkg/version"

	"os"

	"github.com/sirupsen/logrus"
)

var (
	printVersion bool
)

func init() {
	flag.BoolVar(&printVersion, "version", false, "Show version and quit")
	flag.Parse()
}

func main() {
	if printVersion {
		fmt.Println("zookeeper-operator Version:", version.Version)
		fmt.Println("Go Version:", runtime.Version())
		fmt.Printf("Go OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("operator-sdk Version: %v", sdkVersion.Version)
		os.Exit(0)
	}

	logrus.Infof("zookeeper-operator Version: %v", version.Version)
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)

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
