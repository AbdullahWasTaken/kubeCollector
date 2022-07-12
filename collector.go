package collector

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func Workflow() {
	///////////////////////////////////////////////////////////
	//					CONFIGURATION						 //
	///////////////////////////////////////////////////////////
	home := homedir.HomeDir()
	var kubeconfig string = filepath.Join(home, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	///////////////////////////////////////////////////////////
	//						QUERYING						 //
	///////////////////////////////////////////////////////////
	const resources string = "pods"    // check if valid and supported
	const namespace string = "default" // None->all namespaces, ""-> default

	// client := clientset.CoreV1().RESTClient()
	client := getClient(clientset, resources)

	// var reply rest.Result = client.Get().
	// 	Resource(resources).
	// 	Namespace(namespace).
	// 	VersionedParams(&metav1.ListOptions{}, scheme.ParameterCodec).
	// 	Do(context.Background())

	var reply rest.Result = query(client, resources, namespace)

	///////////////////////////////////////////////////////////
	// 						DECODING						 //
	///////////////////////////////////////////////////////////
	runObj, err := reply.Get() // decode the runtime.Object
	if err != nil {
		panic(err.Error())
	}

	// Output Format: JSON
	b, err := json.MarshalIndent(runObj, "", "  ")
	if err != nil {
		panic(err.Error())
	}
	// Output Mode: stdout, file, dgraph
	_ = ioutil.WriteFile("result.json", b, os.ModePerm)
}

func getClient(clientset *kubernetes.Clientset, resource string) rest.Interface {
	var v reflect.Value = reflect.ValueOf(clientset)
	const api = "CoreV1"
	const restClient = "RESTClient"
	var result rest.Interface = v.MethodByName(api).Call([]reflect.Value{})[0].MethodByName(restClient).Call([]reflect.Value{})[0].Interface().(rest.Interface)
	return result
}

func query(client rest.Interface, resource, namespace string) rest.Result {
	var req *rest.Request = client.Get()
	req = req.Resource(resource)
	req = req.Namespace(namespace)
	res := req.VersionedParams(&metav1.ListOptions{}, scheme.ParameterCodec).Do(context.Background())
	return res
}
