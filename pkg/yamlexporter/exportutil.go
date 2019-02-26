/**
 * Copyright (c) 2019 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package yamlexporter

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
)

// YAMLOutputDir where the zookeeper YAML resources will get generated
var YAMLOutputDir string

// CreateOutputSubDir creates a subdirectories where we want create the YAML file
func CreateOutputSubDir(clusterName string, compName string) (string, error) {
	fpath := filepath.Join(clusterName, compName)
	return fpath, createDirIfNotExist(fpath)
}

// GenerateOutputYAMLFile writes YAML output for a resource
func GenerateOutputYAMLFile(subdir string, depType string, data interface{}) error {
	filename := filepath.Join(subdir, depType+"."+"yaml")
	fileFd, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		_ = fileFd.Close()
	}()
	yamlWriter := bufio.NewWriter(fileFd)
	defer yamlWriter.Flush()
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	n, err := yamlWriter.Write(yamlData)
	if err != nil {
		return errors.Wrapf(err, "write failed total bytes written:%d", n)
	}
	return nil
}

func createDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateYAMLOutputDir create output directory for YAML output
func CreateYAMLOutputDir(maindir string) error {
	err := createDirIfNotExist(maindir)
	if err != nil {
		return err
	}
	return nil
}

// ReadInputClusterYAMLFile will read input YAML file and returns Go struct for ZookeeperCluster
func ReadInputClusterYAMLFile(inyamlfile string) (*v1beta1.ZookeeperCluster, error) {
	if _, err := os.Stat(inyamlfile); os.IsNotExist(err) {
		return nil, err
	}
	var z v1beta1.ZookeeperCluster
	source, err := ioutil.ReadFile(inyamlfile)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(source, &z)
	if err != nil {
		return nil, err
	}
	return &z, err
}
