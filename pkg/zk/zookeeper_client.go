package zk

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	"github.com/samuel/go-zookeeper/zk"
)

type ZookeeperClient interface {
	Connect(string) error
	CreateNode(*v1beta1.ZookeeperCluster, string) error
	NodeExists(string) (*zk.Stat, error)
	UpdateNode(string, string, int32) error
	Close()
}

type DefaultZookeeperClient struct {
	conn *zk.Conn
}

func (client *DefaultZookeeperClient) Connect(zkUri string) (err error) {
	host := []string{zkUri}
	conn, _, err := zk.Connect(host, time.Second*5)
	if err != nil {
		return fmt.Errorf("Failed to connect to zookeeper: %s, Reason: %v", zkUri, err)
	}
	client.conn = conn
	return nil
}

func (client *DefaultZookeeperClient) CreateNode(zoo *v1beta1.ZookeeperCluster, zNodePath string) (err error) {
	paths := strings.Split(zNodePath, "/")
	pathLength := len(paths)
	var parentPath string
	for i := 1; i < pathLength-1; i++ {
		parentPath += "/" + paths[i]
		if _, err := client.conn.Create(parentPath, nil, 0, zk.WorldACL(zk.PermAll)); err != nil {
			return fmt.Errorf("Error creating parent zkNode: %s: %v", parentPath, err)
		}
	}
	data := "CLUSTER_SIZE=" + strconv.Itoa(int(zoo.Spec.Replicas))
	childNode := parentPath + "/" + paths[pathLength-1]
	if _, err := client.conn.Create(childNode, []byte(data), 0, zk.WorldACL(zk.PermAll)); err != nil {
		return fmt.Errorf("Error creating sub zkNode: %s: %v", childNode, err)
	}
	return nil
}

func (client *DefaultZookeeperClient) UpdateNode(path string, data string, version int32) (err error) {
	if _, err := client.conn.Set(path, []byte(data), version); err != nil {
		return fmt.Errorf("Error updating zkNode: %v", err)
	}
	return nil
}

func (client *DefaultZookeeperClient) NodeExists(zNodePath string) (nodeStat *zk.Stat, err error) {
	exists, zNodeStat, err := client.conn.Exists(zNodePath)
	if err != nil || !exists {
		return nil, fmt.Errorf("Znode exists check failed for path %s: %v", zNodePath, err)
	}
	return zNodeStat, err
}

func (client *DefaultZookeeperClient) Close() {
	client.conn.Close()
}
