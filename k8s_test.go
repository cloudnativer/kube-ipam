package main

import (
	"fmt"
	"testing"
)

func TestNewClient(t *testing.T) {
	_, err := NewClient()
	fmt.Println(err)

	//kubeconfig, err := ioutil.ReadFile("/etc/kubernetes/ssl/kubectl.kubeconfig")
	//fmt.Println(err)
	//fmt.Println(string(kubeconfig))

}

func TestGetPodInfo(t *testing.T) {
	client, err := NewClient()
	fmt.Println(err)
	annotations, err:= GetPodInfo(client, "housj-test-6d9b74fd4d-dbbsd", "default")
fmt.Println(err)
	fmt.Printf("pod info %+v \n", annotations)
	fmt.Println(annotations[static_ipv4_annotation])
}
