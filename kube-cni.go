package main

import (
	goflag "flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/util/logs"
)

var (
	cniPath         = "/host/etc/cni/net.d/kube-cni.conflist"
	cniConfTemplate = `
{
  "name": "gce-pod-network",
  "cniVersion": "0.3.0",: [
    {
      "type": "ptp",
      "ipam": {
        "type": "host-local",
        "subnet": "%s",
	"routes": [
	  {
	    "dst": "0.0.0.0/0"
	  }
	]
      }
    },
    {
      "type": "portmap",
      "capabilities": {"portMappings": true},
      "noSnat": true
    }
  ]
}
`
)

func main() {
	goflag.CommandLine.Parse([]string{})
	logs.InitLogs()
	defer logs.FlushLogs()
	node, _ := os.Hostname()
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Errorf("failed to get cluster config: %v", err)
		os.Exit(1)
	}
	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Errorf("failed to create new k8s client: %v", err)
		os.Exit(1)
	}
	cidr, err := getPodCidr(c, node)
	if err != nil {
		glog.Errorf("failed to get pod cidr: %v", err)
		os.Exit(1)
	}
	glog.Infof("Install CNI on %q", node)
	cniConf := fmt.Sprintf(cniConfTemplate, cidr)
	if err := ioutil.WriteFile(cniPath, []byte(cniConf), 0644); err != nil {
		glog.Errorf("failed to write cni configuration to %q: %v", cniPath, err)
		os.Exit(1)
	}
	select {}
}

func getPodCidr(client *kubernetes.Clientset, node string) (string, error) {
	n, err := client.Nodes().Get(node, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	if n.Spec.PodCIDR == "" {
		err = fmt.Errorf("podCidr for node %q not found", node)
		return "", err
	}

	return n.Spec.PodCIDR, nil
}
