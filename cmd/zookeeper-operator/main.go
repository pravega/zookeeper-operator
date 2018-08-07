package main

import (
	"context"
	"runtime"

	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	k8sutil "github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	stub "github.com/pravega/zookeeper-operator/pkg/stub"

	"fmt"
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
	namespace, err := getWatchNamespaceAllowBlank()
	if err != nil {
		logrus.Fatalf("Failed to get watch namespace: %v", err)
	}
	resyncPeriod := 5
	logrus.Infof("Watching %s, %s, %s, %d", resource, kind, namespace, resyncPeriod)
	sdk.Watch(resource, kind, namespace, resyncPeriod)
	sdk.Handle(stub.NewHandler())
	sdk.Run(context.TODO())
}

// GetWatchNamespaceAllowBlank returns the namespace the operator should be watching for changes
func getWatchNamespaceAllowBlank() (string, error) {
	ns, found := os.LookupEnv(k8sutil.WatchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", k8sutil.WatchNamespaceEnvVar)
	}
	return ns, nil
}
