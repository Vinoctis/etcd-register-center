package sdk

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"context"
	"fmt"
	"log"
	"sync"
)

type ServiceInfo struct {
	leaseID clientv3.LeaseID
	stopChn chan struct{}
}

type Registry struct {
	client *clientv3.Client
	services map[string]*ServiceInfo
	mu sync.RWMutex
}

func NewRegistry(client *clientv3.Client) *Registry {
	return &Registry{
		client: client,
		services: make(map[string]*ServiceInfo),
	}
}

func (r *Registry) Register(serviceName, addr string, ttl int64) error {

	key := fmt.Sprintf("/services/%s/%s", serviceName, addr)

	// 检查是否已经注册
	if info, exists := r.services[key];exists {
		close(info.stopChn)
		delete(r.services, key)
	}
	// 申请租约
	leaseResp, err := r.client.Grant(context.Background(), ttl)
	if err != nil {
		return nil
	}
	info := &ServiceInfo{
		leaseID: leaseResp.ID,
		stopChn: make(chan struct{}),
	}
	r.services[key] = info
	_, err = r.client.Put(context.Background(), key, addr, clientv3.WithLease(info.leaseID))
	if err != nil {
		return err
	}

	//保持心跳
	go r.keepAlive(key, info.leaseID, info.stopChn)

	return nil
}

func (r *Registry) keepAlive(key string, leaseID clientv3.LeaseID, stopChn chan struct{}) {
	ch , err := r.client.KeepAlive(context.Background(), leaseID)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case <- stopChn:
			return
		case _, ok := <-ch:
			if !ok {
				log.Println("lease expired, re-registering")
				return 
			}
		}
	}
}

//取消注册，服务下线
func (r *Registry) Deregister(serviceName , address string ) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := fmt.Sprintf("/services/%s/%s", serviceName, address)
	if info, exists := r.services[key];exists {
		r.client.Delete(context.Background(), key)
		close(info.stopChn)
		delete(r.services, key)
	}
}