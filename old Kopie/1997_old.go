/*
Copyright 2019 Google LLC

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

package utils

import (
	"fmt"
	"os"
	"strings"

	metadataClient "github.com/google/knative-gcp/pkg/gclient/metadata"
)

const (
	clusterNameAttr = "cluster-name"
	// ProjectIDEnvKey is the name of environmental variable for project ID
	ProjectIDEnvKey = "PROJECT_ID"
)

// defaultMetadataClientCreator is a create function to get a default metadata client. This can be
// swapped during testing.
var defaultMetadataClientCreator func() metadataClient.Client = metadataClient.NewDefaultMetadataClient

// projectIDFromEnv loads project ID from env once when startup.
var projectIDFromEnv string

func init() {
	projectIDFromEnv = os.Getenv(ProjectIDEnvKey)
}

// ProjectIDEnvConfig is a struct to parse project ID from env var
type ProjectIDEnvConfig struct {
	ProjectID string `envconfig:"PROJECT_ID"`
}

// ProjectIDOrDefault returns the project ID by performing the following order:
// 1) if the input project ID is valid, simply use it.
// 2) if there is a PROJECT_ID environmental variable, use it.
// 3) use metadataClient to resolve project id.
func ProjectIDOrDefault(projectID string) (string, error) {
	if projectID != "" {
		return projectID, nil
	}
	if projectIDFromEnv != "" {
		return projectIDFromEnv, nil
	}
	// Otherwise, ask GKE metadata server.
	projectGKE, err := defaultMetadataClientCreator().ProjectID()
	if err != nil {
		return "", err
	}
	return projectGKE, nil
}

// ClusterName returns the cluster name for a particular resource.
func ClusterName(clusterName string, client metadataClient.Client) (string, error) {
	// If clusterName is set, then return that one.
	if clusterName != "" {
		return clusterName, nil
	}
	clusterName, err := client.InstanceAttributeValue(clusterNameAttr)
	if err != nil {
		return "", err
	}
	return clusterName, nil
// ZoneToRegion converts a GKE zone to its region
func ZoneToRegion(zone string) (string, error) {
	fields := strings.Split(zone, "-")
	if len(fields) == 3 {
		return fmt.Sprintf("%s-%s", fields[0], fields[1]), nil
	}
	// We can as well treat xx-xx as region and simply return,
	// but let's be strict here
	return "", fmt.Errorf("zone %s is not valid", zone)
}

// ClusterRegion returns the region of the cluster
func ClusterRegion(clusterRegion string, clientCreator func() metadataClient.Client) (string, error) {
	if clusterRegion != "" {
		return clusterRegion, nil
	}
	zone, err := clientCreator().Zone()
	if err != nil {
		return "", err
	}
	return ZoneToRegion(zone)
}

// ClusterRegionGetter is a handler that gets cluster region
type ClusterRegionGetter func() (string, error)

// NewClusterRegionGetter returns CluterRegionGetter in production code
func NewClusterRegionGetter() ClusterRegionGetter {
	return func() (string, error) {
		return ClusterRegion("", defaultMetadataClientCreator)
	}
}
