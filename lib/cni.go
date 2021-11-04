// Copyright 2015 CNI authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kipam

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	"kube-ipam/backend/allocator"
	"kube-ipam/backend/etcd"

	log "github.com/sirupsen/logrus"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
)

// Set the version number and release date of Kube-ipam.
const (
	static_ipv4_annotation    string = "kube-ipam.ip"
	static_netmask_annotation string = "kube-ipam.netmask"
	static_gateway_annotation string = "kube-ipam.gateway"
)

func loadNetConf(bytes []byte) (*types.NetConf, string, error) {
	n := &types.NetConf{}
	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, "", fmt.Errorf("failed to load netconf: %v", err)
	}
	return n, n.CNIVersion, nil
}

func CmdCheck(args *skel.CmdArgs) error {
	_, _, err := allocator.LoadIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}
	return nil
}

func CmdAdd(args *skel.CmdArgs) error {
	ipamConf, confVersion, err := allocator.LoadIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}

	var kubeCfg string
	if ipamConf.KubeConfig != "" {
		kubeCfg = ipamConf.KubeConfig
	} else {
		kubeCfg = "/etc/kubernetes/pki/kubectl.kubeconfig"
	}

	podIp, podNetmask, podGateway, err := LoadIPFromPodAnnotation(args.Args, kubeCfg)
	if err != nil {
		log.Errorf("load ip error %v", err)
	}

	log.Infof("load ip from k8s pod annotation %s %s %s", podIp, podNetmask, podGateway)

	result := &current.Result{}
	if ipamConf.ResolvConf != "" {
		dns, err := parseResolvConf(ipamConf.ResolvConf)
		if err != nil {
			return err
		}
		log.Infof("load ResolvConf")
		result.DNS = *dns
	}

	log.Infof("Connect to etcd datastore")
	store, err := etcd.New(ipamConf.Name, ipamConf)
	if err != nil {
		return err
	}
	defer store.Close()

	// Keep the allocators we used, so we can release all IPs if an error
	// occurs after we start allocating
	allocs := []*allocator.IPAllocator{}

	// Store all requested IPs in a map, so we can easily remove ones we use
	// and error if some remain
	requestedIPs := map[string]net.IP{} //net.IP cannot be a key

	for _, ip := range ipamConf.IPArgs {
		requestedIPs[ip.String()] = ip
	}

	//It is applicable to the case that the podIp is of multiple IP addresses, and the podIp is split and converted into a podIpArray array.
	//podIpArray := strings.Split(podIp, `,`)

	if podIp != "" {

		log.Infof("Get the configuration of fixed IP")
		// Check to see if there are any custom IPs requested.
		var ipConf *current.IPConfig
		podNetmaskArray := strings.Split(podNetmask, `.`)
		podNetmask0, _ := strconv.Atoi(podNetmaskArray[0])
		podNetmask1, _ := strconv.Atoi(podNetmaskArray[1])
		podNetmask2, _ := strconv.Atoi(podNetmaskArray[2])
		podNetmask3, _ := strconv.Atoi(podNetmaskArray[3])

		ipConf = &current.IPConfig{
			Version: "4",
			Address: net.IPNet{
				IP:   net.ParseIP(podIp),
				Mask: net.IPv4Mask(byte(podNetmask0), byte(podNetmask1), byte(podNetmask2), byte(podNetmask3)),
			},
			Gateway: net.ParseIP(podGateway)}

		result.IPs = append(result.IPs, ipConf)

	} else {

		log.Infof("Get the configuration of random IP")
		for idx, rangeset := range ipamConf.Ranges {
			allocator := allocator.NewIPAllocator(&rangeset, store, idx)

			// Check to see if there are any custom IPs requested in this range.
			var requestedIP net.IP
			var ipConf *current.IPConfig

			for k, ip := range requestedIPs {
				if rangeset.Contains(ip) {
					requestedIP = ip
					delete(requestedIPs, k)
					break
				}
			}
			ipConf, err = allocator.Get(args.ContainerID, requestedIP)

			if err != nil {
				// Deallocate all already allocated IPs
				for _, alloc := range allocs {
					_ = alloc.Release(args.ContainerID)
				}
				return fmt.Errorf("failed to allocate for range %d: %v", idx, err)
			}

			allocs = append(allocs, allocator)

			result.IPs = append(result.IPs, ipConf)
		}

	}

	// If an IP was requested that wasn't fulfilled, fail
	if len(requestedIPs) != 0 {
		for _, alloc := range allocs {
			_ = alloc.Release(args.ContainerID)
		}
		errstr := "failed to allocate all requested IPs:"
		for _, ip := range requestedIPs {
			errstr = errstr + " " + ip.String()
		}
		return fmt.Errorf(errstr)
	}

	result.Routes = ipamConf.Routes

	return types.PrintResult(result, confVersion)

}

func LoadIPFromPodAnnotation(args string, kubeCfg string) (string, string, string, error) {
	log.Debugf("read args.args ==> %s", args)
	k8sArgs := K8sArgs{}
	if err := types.LoadArgs(args, &k8sArgs); err != nil {
		log.Errorf("read k8s args error %v", err)
		return "", "", "", err
	}
	client, err := NewClient(kubeCfg)
	if err != nil {
		log.Errorf("create k8s client error %v", err)
		return "", "", "", err
	}
	annotations, err := GetPodInfo(client, string(k8sArgs.K8S_POD_NAME), string(k8sArgs.K8S_POD_NAMESPACE))
	log.Infof("pod %s annotations %+v", string(k8sArgs.K8S_POD_NAME), annotations)
	return annotations[static_ipv4_annotation], annotations[static_netmask_annotation], annotations[static_gateway_annotation], nil
}

func CmdDel(args *skel.CmdArgs) error {
	ipamConf, _, err := allocator.LoadIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}

	store, err := etcd.New(ipamConf.Name, ipamConf)
	if err != nil {
		return err
	}
	defer store.Close()

	// Loop through all ranges, releasing all IPs, even if an error occurs
	var errors []string
	for idx, rangeset := range ipamConf.Ranges {
		ipAllocator := allocator.NewIPAllocator(&rangeset, store, idx)

		err := ipAllocator.Release(args.ContainerID)
		if err != nil {
			errors = append(errors, err.Error())
		}
	}

	if errors != nil {
		return fmt.Errorf(strings.Join(errors, ";"))
	}
	return nil
}
