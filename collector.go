package collector

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Collector represents the state of a kubernetes system accessed through config
type Collector struct {
	state  map[schema.GroupVersionResource]unstructured.UnstructuredList
	config *rest.Config
}

// NewCollector creates a new Collector instance
// and build the config object from kubeconfig path.
func NewCollector(kubeconfig string) *Collector {
	st := map[schema.GroupVersionResource]unstructured.UnstructuredList{}
	conf, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	return &Collector{state: st, config: conf}
}

// Collect is responsible for querying the Kubernetes API server
// for the states of all available resources.
func (c *Collector) Collect() (cs map[schema.GroupVersionResource]unstructured.UnstructuredList) {
	cs = map[schema.GroupVersionResource]unstructured.UnstructuredList{}
	dc, err := dynamic.NewForConfig(c.config)
	if err != nil {
		log.Error(err)
	}
	tc, err := kubernetes.NewForConfig(c.config)
	if err != nil {
		log.Error(err)
	}
	res, _ := getResources(tc.DiscoveryClient)
	for _, gvr := range res {
		st, err := getResourceState(dc, gvr)
		if err != nil {
			log.Error(err)
		} else {
			if len(st.Items) > 0 {
				cs[gvr] = st
			}
		}
	}
	return cs
}

// SaveState converts the state of the Collector into JSON format and
// stores it individually in a directory under `outDir` called JSON.
func (c *Collector) SaveState(outDir string) {
	jsonPath := filepath.Join(outDir, "JSON")
	err := os.MkdirAll(jsonPath, os.ModePerm)
	if err != nil {
		log.Panic(err)
	}
	for gvr, st := range c.state {
		// convert Unstructured.Undtructred to JSON
		jsonStr, err := json.MarshalIndent(st, "", "    ")
		if err != nil {
			log.Error(err)
		} else {
			// write JSON to file
			err = os.WriteFile(filepath.Join(jsonPath, gvr.Resource+".json"), jsonStr, os.ModePerm)
			if err != nil {
				log.Error(err)
			}
		}
	}
}

// getResources returns a list of `schema.GroupVersionResource` which are
// responsible for uniquely identifying every resource installed on the server.
func getResources(client *discovery.DiscoveryClient) ([]schema.GroupVersionResource, error) {
	res := []schema.GroupVersionResource{}
	grps, err := client.ServerGroups()
	if err != nil {
		return nil, err
	}
	// Iterating over groups supported by the server.
	for _, group := range grps.Groups {
		// Iterating over all versions (may not be preferred) supported by the server.
		for _, gv := range group.Versions {
			apiResources, err := client.ServerResourcesForGroupVersion(gv.GroupVersion)
			if err != nil {
				return nil, err
			}
			// Iterating over the resources present under a particular group and version.
			for _, apiRes := range apiResources.APIResources {
				// schmea.GroupVersionResource is uniquely defined by Group, Version, and ResourceName.
				gvr := schema.GroupVersionResource{Group: group.Name, Version: gv.Version, Resource: apiRes.Name}
				// pruning out resources that do not support LIST action.
				if slices.Contains(apiRes.Verbs, "list") {
					res = append(res, gvr)
				}
			}
		}
	}
	return res, nil
}

// getResourceState returns a list of state of resources
// unqiuely identified by `schema.GroupVersionResource`.
func getResourceState(client dynamic.Interface, gvr schema.GroupVersionResource) (unstructured.UnstructuredList, error) {
	resourcelist, err := client.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return unstructured.UnstructuredList{}, err
	}
	return *resourcelist, nil
}
