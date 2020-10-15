package main

import (
	"fmt"
	"testing"
)

func TestNewClient(t *testing.T) {
	_, err := NewClient()
	fmt.Println(err)

	//kubeconfig, err := ioutil.ReadFile("./kube-config")
	//fmt.Println(err)
	//fmt.Println(string(kubeconfig))

}

func TestGetPodInfo(t *testing.T) {
	client, err := NewClient()
	fmt.Println(err)
	annotations, err:= GetPodInfo(client, "fm-barge-backend-stable-dd67c77b5-c7nxc", "default")
fmt.Println(err)
	fmt.Printf("pod info %+v \n", annotations)
	fmt.Println(annotations[static_ipv4_annotation])
}
