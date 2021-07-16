/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/types"

	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/pravega/zookeeper-operator/pkg/controller/zookeepercluster"
	"github.com/pravega/zookeeper-operator/pkg/version"
	"github.com/pravega/zookeeper-operator/pkg/yamlexporter"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	log         = logf.Log.WithName("cmd")
	versionFlag bool
)

func init() {
	flag.BoolVar(&versionFlag, "version", false, "Show version and quit")
}

func printVersion() {
	log.Info(fmt.Sprintf("zookeeper-operator Version: %v", version.Version))
	log.Info(fmt.Sprintf("Git SHA: %s", version.GitSHA))
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("operator-sdk Version: %v", sdkVersion.Version))
}

func main() {
	flags := flag.NewFlagSet("myFlagSet", flag.ExitOnError)
	ifilePtr := flags.String("i", "./ZookeeperCluster.yaml", "Input YAML file")
	odirPtr := flags.String("o", ".", "YAML output directory")

	_ = flags.Parse(os.Args[1:])

	log.Info(fmt.Sprintf("Input YAML file -i:%s", *ifilePtr))
	log.Info(fmt.Sprintf("Output YAML Directory -o:%s", *odirPtr))

	// Read input YAML file -- This is the ZookeeperCluster Resource YAML file
	log.Info(fmt.Sprintf("Reading YAML file from the file:%s", *ifilePtr))
	z, err := yamlexporter.ReadInputClusterYAMLFile(*ifilePtr)
	if err != nil {
		log.Error(err, "read input YAML file failed")
		os.Exit(1)
	}

	// create base output directory and sub-directories named based on the deployment phase
	log.Info(fmt.Sprintf("create base output dir:%s and phase based subdirs", *odirPtr))
	err = yamlexporter.CreateYAMLOutputDir(*odirPtr)
	if err != nil {
		log.Error(err, "create output dir failed")
		os.Exit(1)
	}

	// we need to provide our own UID for ECSCluster Resource, since the rest of the resources will reference UID of ECSCluster and it must be there.
	rand.Seed(time.Now().UnixNano())
	uid := strconv.FormatUint(rand.Uint64(), 10)
	log.Info(fmt.Sprintf("UID of the ECSCluster Resource:%s\n", uid))
	z.UID = types.UID(uid)

	yamlexporter.YAMLOutputDir = *odirPtr

	reconcilerZookeeper := zookeepercluster.YAMLExporterReconciler(z)

	// Generate YAML files
	err = reconcilerZookeeper.GenerateYAML(z)
	if err != nil {
		log.Error(err, "YAML file generation failed")
		os.Exit(1)
	}
}
