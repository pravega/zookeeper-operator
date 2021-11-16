package controller

import (
	"github.com/pravega/zookeeper-operator/pkg/controller/zookeeperbackup"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, zookeeperbackup.Add)
}
