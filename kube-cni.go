package main

import (
	"fmt"
	"io/ioutil"
	"os"

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
		fmt.Println(err)
		os.Exit(1)
	}
	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		os.Exit(1)
	}
	cidr, err := getPodCidr(c, node)
	if err != nil {
		os.Exit(1)
	}
	fmt.Sprintf(cniConf, cidr)
	ioutil.WriteFile(cniDir, []byte(cniConf), 0644)
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
