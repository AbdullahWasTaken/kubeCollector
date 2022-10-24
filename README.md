## kubeCollector

A kubernetes cluster state collector

### Synopsis

KubeCollector is a general purpose state collector for kubernetes clusters 
	without any restrictions on the third-party resources installed on the server.

	Examples:
		# Collect the state into default location
		kubeCollector /.kube/config
	
		# Collect the state into specified location
		kubeCollector /.kube/config -out /outputDir

```
kubeCollector <kubeconfig> [flags]
```

### Options

```
  -h, --help         help for kubeCollector
      --out string   output directory (default "out")
```