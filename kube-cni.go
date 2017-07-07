package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	cniDir  = "/etc/cni/net.d"
	cniConf = `
{
        "cniVersion": "0.3.1",
        "name": "mynet",
        "plugins": [
                {
                        "type": "ptp",
                        "ipMasq": true,
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
                        "snat": false
                }
        ]
}
`
)

func main() {
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
	glog.Infof("Adding config %q to %q", cniConf, cniDir)
	if err := ioutil.WriteFile(cniDir, []byte(cniConf), 0644); err != nil {
		glog.Errorf("failed to write cni configuration to %q", cniDir)
		os.Exit(1)
	}
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
