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

package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strconv"
	"time"

	"kube-ipam/backend/allocator"

	"github.com/coreos/etcd/clientv3"
	"kube-ipam/backend"
)

const ETCDPrefix string = "/etcd-cni/networks/"

// Store is a simple disk-backed store that creates one file per IP
// address in a given directory. The contents of the file are the container ID.
type Store struct {
	EtcdClient    *clientv3.Client
	EtcdKeyPrefix string
}

// Store implements the Store interface
var _ backend.Store = &Store{}

func New(name string, ipamConfig *allocator.IPAMConfig) (*Store, error) {
	etcdClient, err := connectStore(ipamConfig.EtcdConfig)
	if err != nil {
		panic(err)
	}
	netConfig, err := netConfigJson(ipamConfig)
	etcdKeyPrefix, err := initStore(name, netConfig, etcdClient)
	// write values in Store object
	store := &Store{
		EtcdClient:    etcdClient,
		EtcdKeyPrefix: etcdKeyPrefix,
	}
	return store, nil
}

func initStore(name string, netConfig string, etcdClient *clientv3.Client) (string, error) {
	key := ETCDPrefix + name

	_, err := etcdClient.Put(context.TODO(), key, netConfig)
	if err != nil {
		panic(err)
	}
	return key, nil
}

func netConfigJson(ipamConfig *allocator.IPAMConfig) (string, error) {
	conf, err := json.Marshal(ipamConfig.Ranges)
	return string(conf), err
}

func (s *Store) Reserve(id string, ip net.IP, rangeID string) (bool, error) {
	usedIPPrefix := s.EtcdKeyPrefix + "/used/"

	key := usedIPPrefix + ip.String()
	resp, err := s.EtcdClient.Get(context.TODO(), key)
	if err != nil {
		return false, err
	}
	if resp.Count > 0 {
		return false, nil
	}
	value := id
	_, err = s.EtcdClient.Put(context.TODO(), key, value)
	if err != nil {
		return false, err
	}

	key = s.EtcdKeyPrefix + "/lastReserved/" + rangeID
	_, err = s.EtcdClient.Put(context.TODO(), key, ip.String())
	if err != nil {
		return false, err
	}
	return true, nil
}

// LastReservedIP returns the last reserved IP if exists
func (s *Store) LastReservedIP(rangeID string) (net.IP, error) {
	key := s.EtcdKeyPrefix + "/lastReserved/" + rangeID
	resp, err := s.EtcdClient.Get(context.TODO(), key)
	if err != nil {
		return nil, err
	}
	if resp.Count == 0 {
		return nil, errors.New("Can not find last reserved ip")
	}
	data := string(resp.Kvs[0].Value)
	return net.ParseIP(string(data)), nil
}

func (s *Store) Release(ip net.IP) error {
	key := s.EtcdKeyPrefix + "/used/" + ip.String()
	_, err := s.EtcdClient.Delete(context.TODO(), key)
	if err != nil {
		return err
	} else {
		return nil
	}
}

// N.B. This function eats errors to be tolerant and
// release as much as possible
func (s *Store) ReleaseByID(id string) error {
	key := s.EtcdKeyPrefix + "/used/"
	resp, err := s.EtcdClient.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	if resp.Count > 0 {
		for _, kv := range resp.Kvs {
			if string(kv.Value) == id {
				_, err = s.EtcdClient.Delete(context.TODO(), string(kv.Key))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *Store) Close() error {
	// stub we don't need close anything
	return nil
}

func (s *Store) Lock() error {
	key := s.EtcdKeyPrefix + "/lock"

	kv := clientv3.NewKV(s.EtcdClient)

	retryTimes := 20
	leaseTTL := 10

	getLock := false

	for i := 0; i < retryTimes; i++ {
		// define a lease
		lease := clientv3.NewLease(s.EtcdClient)
		var leaseResp *clientv3.LeaseGrantResponse
		var err error
		if leaseResp, err = lease.Grant(context.TODO(), int64(leaseTTL)); err != nil {
			return err
		}

		// get leaseId
		leaseId := leaseResp.ID

		// define txn
		txn := kv.Txn(context.TODO())
		txn.If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
			Then(clientv3.OpPut(key, strconv.FormatInt(int64(leaseId), 10), clientv3.WithLease(leaseId))).
			Else(clientv3.OpGet(key))

		// commit txn
		var txnResp *clientv3.TxnResponse
		if txnResp, err = txn.Commit(); err != nil {
			return err
		}

		// return if successed
		if txnResp.Succeeded {
			getLock = true
			break
			// try again
		} else {
			time.Sleep(time.Second * 2)
			continue
		}
	}

	if getLock {
		return nil
	} else {
		return errors.New("Can not get lock")
	}
}

func (s *Store) Unlock() error {
	key := s.EtcdKeyPrefix + "/lock"
	resp, err := s.EtcdClient.Get(context.TODO(), key)
	if err != nil {
		return err
	}
	if resp.Count > 0 {
		value := string(resp.Kvs[0].Value)
		leaseId, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		lease := clientv3.NewLease(s.EtcdClient)
		lease.Revoke(context.TODO(), clientv3.LeaseID(leaseId))
	}
	return nil
}
