package collector

import (
	"context"

	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

type ClusterState = map[string]unstructured.UnstructuredList

// GetState returns the state `ClusterState` of all the resources instantiated
// in the kubernetes cluster accessible using `disClient` and `dynClient`.
func GetState(disClient *discovery.DiscoveryClient, dynClient *dynamic.Interface) ClusterState {
	groupList, err := disClient.ServerGroups()
	if err != nil {
		log.Fatal(err)
	}
	cs := ClusterState{}
	// Iterating over groups supported by the server.
	for _, group := range groupList.Groups {
		// Iterating over all versions (may not be preferred) supported by the server.
		for _, gvs := range group.Versions {
			apiResourceList, err := disClient.ServerResourcesForGroupVersion(gvs.GroupVersion)
			if err != nil {
				log.Error(err)
			}
			// Iterating over the resources present under a particular group and version.
			for _, apiRes := range apiResourceList.APIResources {
				// schmea.GroupVersionResource is uniquely defined by Group, Version, and ResourceName.
				gvr := schema.GroupVersionResource{Group: group.Name, Version: gvs.Version, Resource: apiRes.Name}
				// pruning out resources that do not support LIST action.
				if slices.Contains(apiRes.Verbs, "list") {
					res := GetServerResources(*dynClient, gvr)
					if len(res.Items) > 0 {
						cs[gvr.Resource] = res
					}
				}
			}
		}
	}
	return cs
}

// GetNamespacedState returns the state `ClusterState` of all the resources present in
// the namespace `ns` of the kubernetes cluster accessible using `disClient` and `dynClient`.
func GetNamespacedState(disClient *discovery.DiscoveryClient, dynClient *dynamic.Interface, ns string) ClusterState {
	groupList, err := disClient.ServerGroups()
	if err != nil {
		log.Fatal(err)
	}
	cs := ClusterState{}
	// Iterating over groups supported by the server.
	for _, group := range groupList.Groups {
		// Iterating over all versions (may not be preferred) supported by the server.
		for _, gvs := range group.Versions {
			apiResourceList, err := disClient.ServerResourcesForGroupVersion(gvs.GroupVersion)
			if err != nil {
				log.Error(err)
			}
			// Iterating over the resources present under a particular group and version.
			for _, apiRes := range apiResourceList.APIResources {
				// schmea.GroupVersionResource is uniquely defined by Group, Version, and ResourceName.
				gvr := schema.GroupVersionResource{Group: group.Name, Version: gvs.Version, Resource: apiRes.Name}
				// pruning out resources that do not support LIST action and are not namespaced.
				if slices.Contains(apiRes.Verbs, "list") && apiRes.Namespaced {
					res := GetNamespacedResources(*dynClient, gvr, ns)
					if len(res.Items) > 0 {
						cs[gvr.Resource] = res
					}
				}
			}
		}
	}
	return cs
}

// GetServerResources performs the LIST action over a resource type
// and returns all instances of it as a `Unstructured.UnstructuredList` object.
func GetServerResources(cli dynamic.Interface, gvr schema.GroupVersionResource) unstructured.UnstructuredList {
	resourcelist, err := cli.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Error(err)
		return unstructured.UnstructuredList{}
	}
	return *resourcelist
}

// GetNamespacedResources performs the LIST action over a namespaced resource type
// and returns all instances of it as a `Unstructured.UnstructuredList` object.
func GetNamespacedResources(cli dynamic.Interface, gvr schema.GroupVersionResource, ns string) unstructured.UnstructuredList {
	resourcelist, err := cli.Resource(gvr).Namespace(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Error(err)
		return unstructured.UnstructuredList{}
	}
	return *resourcelist
}
