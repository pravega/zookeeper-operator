package zk

import (
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/q8s-io/zookeeper-operator-pravega/pkg/apis/zookeeper/v1beta1"
)

// MakeConfigMap returns a zookeeper config map
func MakeConfigMap(z *v1beta1.ZookeeperCluster) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.ConfigMapName(),
			Namespace: z.Namespace,
			Labels:    z.Spec.Labels,
		},
		Data: map[string]string{
			"zoo.cfg":                makeZkConfigString(z.Spec),
			"log4j.properties":       makeZkLog4JConfigString(),
			"log4j-quiet.properties": makeZkLog4JQuietConfigString(),
			"env.sh":                 makeZkEnvConfigString(z),
		},
	}
}

func makeZkConfigString(s v1beta1.ZookeeperClusterSpec) string {
	return "4lw.commands.whitelist=cons, envi, conf, crst, srvr, stat, mntr, ruok\n" +
		"dataDir=/data\n" +
		"standaloneEnabled=false\n" +
		"reconfigEnabled=true\n" +
		"skipACL=yes\n" +
		"metricsProvider.className=org.apache.zookeeper.metrics.prometheus.PrometheusMetricsProvider\n" +
		"metricsProvider.httpPort=7000\n" +
		"metricsProvider.exportJvmInfo=true\n" +
		"initLimit=" + strconv.Itoa(s.Conf.InitLimit) + "\n" +
		"syncLimit=" + strconv.Itoa(s.Conf.SyncLimit) + "\n" +
		"tickTime=" + strconv.Itoa(s.Conf.TickTime) + "\n" +
		"globalOutstandingLimit=" + strconv.Itoa(s.Conf.GlobalOutstandingLimit) + "\n" +
		"preAllocSize=" + strconv.Itoa(s.Conf.PreAllocSize) + "\n" +
		"snapCount=" + strconv.Itoa(s.Conf.SnapCount) + "\n" +
		"commitLogCount=" + strconv.Itoa(s.Conf.CommitLogCount) + "\n" +
		"snapSizeLimitInKb=" + strconv.Itoa(s.Conf.SnapSizeLimitInKb) + "\n" +
		"maxCnxns=" + strconv.Itoa(s.Conf.MaxCnxns) + "\n" +
		"maxClientCnxns=" + strconv.Itoa(s.Conf.MaxClientCnxns) + "\n" +
		"minSessionTimeout=" + strconv.Itoa(s.Conf.MinSessionTimeout) + "\n" +
		"maxSessionTimeout=" + strconv.Itoa(s.Conf.MaxSessionTimeout) + "\n" +
		"autopurge.snapRetainCount=" + strconv.Itoa(s.Conf.AutoPurgeSnapRetainCount) + "\n" +
		"autopurge.purgeInterval=" + strconv.Itoa(s.Conf.AutoPurgePurgeInterval) + "\n" +
		"quorumListenOnAllIPs=" + strconv.FormatBool(s.Conf.QuorumListenOnAllIPs) + "\n" +
		"dynamicConfigFile=/data/zoo.cfg.dynamic\n" +
		"clientPort=" + strconv.Itoa(int(s.Ports[0].ContainerPort))
}

func makeZkLog4JQuietConfigString() string {
	return "log4j.rootLogger=ERROR, CONSOLE\n" +
		"log4j.appender.CONSOLE=org.apache.log4j.ConsoleAppender\n" +
		"log4j.appender.CONSOLE.Threshold=ERROR\n" +
		"log4j.appender.CONSOLE.layout=org.apache.log4j.PatternLayout\n" +
		"log4j.appender.CONSOLE.layout.ConversionPattern=%d{ISO8601} [myid:%X{myid}] - %-5p [%t:%C{1}@%L] - %m%n\n"
}

func makeZkLog4JConfigString() string {
	return "zookeeper.root.logger=CONSOLE\n" +
		"zookeeper.console.threshold=INFO\n" +
		"log4j.rootLogger=${zookeeper.root.logger}\n" +
		"log4j.appender.CONSOLE=org.apache.log4j.ConsoleAppender\n" +
		"log4j.appender.CONSOLE.Threshold=${zookeeper.console.threshold}\n" +
		"log4j.appender.CONSOLE.layout=org.apache.log4j.PatternLayout\n" +
		"log4j.appender.CONSOLE.layout.ConversionPattern=%d{ISO8601} [myid:%X{myid}] - %-5p [%t:%C{1}@%L] - %m%n\n"
}

func makeZkEnvConfigString(z *v1beta1.ZookeeperCluster) string {
	ports := z.ZookeeperPorts()
	return "#!/usr/bin/env bash\n\n" +
		// "DOMAIN=" + headlessDomain(z) + "\n" +
		"DOMAIN=" + z.GetName() + "-service" + "\n" +
		"QUORUM_PORT=" + strconv.Itoa(int(ports.Quorum)) + "\n" +
		"LEADER_PORT=" + strconv.Itoa(int(ports.Leader)) + "\n" +
		"CLIENT_HOST=" + z.GetClientServiceName() + "\n" +
		"CLIENT_PORT=" + strconv.Itoa(int(ports.Client)) + "\n" +
		"CLUSTER_NAME=" + z.GetName() + "\n" +
		"CLUSTER_SIZE=" + fmt.Sprint(z.Spec.Replicas) + "\n"
}
