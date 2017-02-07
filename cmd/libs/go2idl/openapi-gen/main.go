/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This package generates openAPI definition file to be used in open API spec generation on API servers. To generate
// definition for a specific type or package add "+k8s:openapi-gen=true" tag to the type/package comment lines. To
// exclude a type from a tagged package, add "+k8s:openapi-gen=false" tag to the type comment lines.
package main

import (
	"path/filepath"
	"sort"
	"strings"

	"k8s.io/gengo/args"
	"k8s.io/kubernetes/cmd/libs/go2idl/client-gen/types"
	"k8s.io/kubernetes/cmd/libs/go2idl/openapi-gen/generators"

	"github.com/golang/glog"
)

var (
	inputVersions = []string{
		"api/",
		"authentication/",
		"authorization/",
		"autoscaling/",
		"batch/",
		"certificates/",
		"extensions/",
		"rbac/",
		"storage/",
		"apps/",
		"policy/",
	}
	basePath = "k8s.io/kubernetes/pkg/apis"
)


func main() {
	arguments := args.Default()

	// Override defaults.
	arguments.OutputFileBaseName = "openapi_generated"
	arguments.GoHeaderFilePath = filepath.Join(args.DefaultSourceTree(), "k8s.io/kubernetes/hack/boilerplate/boilerplate.go.txt")
	arguments.OutputPackagePath = "k8s.io/kubernetes/pkg/client/listers/apps/v1beta1"

	inputPath, _, _, err := parseInputVersions()
	if err != nil {
		glog.Fatalf("Error: %v", err)
	}

	//includedTypesOverrides, err := parseIncludedTypesOverrides()
	if err != nil {
		glog.Fatalf("Unexpected error: %v", err)
	}
	arguments.InputDirs = inputPath

	// Run it.
	if err := arguments.Execute(
		generators.NameSystems(),
		generators.DefaultNameSystem(),
		generators.Packages,
	); err != nil {
		glog.Fatalf("Error: %v", err)
	}
	glog.V(2).Info("Completed successfully.")
}

func parseInputVersions() (paths []string, groups []types.GroupVersions, gvToPath map[types.GroupVersion]string, err error) {
	var seenGroups = make(map[types.Group]*types.GroupVersions)
	gvToPath = make(map[types.GroupVersion]string)
	for _, input := range inputVersions {
		gvPath, gvString := parsePathGroupVersion(input)
		gv, err := types.ToGroupVersion(gvString)
		if err != nil {
			return nil, nil, nil, err
		}
		if group, ok := seenGroups[gv.Group]; ok {
			(*seenGroups[gv.Group]).Versions = append(group.Versions, gv.Version)
		} else {
			seenGroups[gv.Group] = &types.GroupVersions{
				Group:    gv.Group,
				Versions: []types.Version{gv.Version},
			}
		}

		path := versionToPath(gvPath, gv.Group.String(), gv.Version.String())
		paths = append(paths, path)
		gvToPath[gv] = path
	}
	var groupNames []string
	for groupName := range seenGroups {
		groupNames = append(groupNames, groupName.String())
	}
	sort.Strings(groupNames)
	for _, groupName := range groupNames {
		groups = append(groups, *seenGroups[types.Group(groupName)])
	}

	return paths, groups, gvToPath, nil
}

func versionToPath(gvPath string, group string, version string) (path string) {
	// special case for the core group
	if group == "api" {
		path = filepath.Join(basePath, "../api", version)
	} else {
		path = filepath.Join(basePath, gvPath, group, version)
	}
	return
}

func parsePathGroupVersion(pgvString string) (gvPath string, gvString string) {
	subs := strings.Split(pgvString, "/")
	length := len(subs)
	switch length {
	case 0, 1, 2:
		return "", pgvString
	default:
		return strings.Join(subs[:length-2], "/"), strings.Join(subs[length-2:], "/")
	}
}