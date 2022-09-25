package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AbdullahWasTaken/kube-miner/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// GetClusterState is a naive implementation of operational state extraction from a kubernetes cluster.
func GetClusterState(cfg *config.Config) {
	groupList, err := cfg.DisClient.ServerGroups()
	if err != nil {
		log.Fatal(err)
	}
	// Iterating over groups supported by the server.
	for _, group := range groupList.Groups {
		// Iterating over all versions (may not be preferred) supported by the server.
		for _, gvs := range group.Versions {
			apiResourceList, err := cfg.DisClient.ServerResourcesForGroupVersion(gvs.GroupVersion)
			if err != nil {
				log.Error(err)
			}
			// Iterating over the resources present under a particular group and version.
			for _, apiRes := range apiResourceList.APIResources {
				// schmea.GroupVersionResource is uniquely defined by Group, Version, and ResourceName.
				gvr := schema.GroupVersionResource{Group: group.Name, Version: gvs.Version, Resource: apiRes.Name}
				// pruning out resources that do not support LIST action.
				if slices.Contains(apiRes.Verbs, "list") {
					GetResources(*cfg.DynClient, gvr, cfg.Out)
				}
			}
		}
	}
}

// GetResources performs the LIST action over a valid scheme.GroupVersionResource
// iterating over each resource present and encoding the Unstructured.Unstructured object into JSON
func GetResources(client dynamic.Interface, gvr schema.GroupVersionResource, path string) {
	resourcelist, err := client.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Error(err)
		return
	}

	// creating the output path if it doesn't exists.
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	if numRes := len(resourcelist.Items); numRes > 0 {
		log.Infof("Found %v instances of resource type %v\n", numRes, gvr.Resource)
		b, _ := json.MarshalIndent(resourcelist.Items, "", "    ")
		os.WriteFile(filepath.Join(path, fmt.Sprintf("%v.json", gvr.Resource)), b, os.ModePerm)
	}
}
