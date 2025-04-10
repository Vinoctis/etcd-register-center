package sdk

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"context"
	"fmt"
)

type Discover struct {
	client *clientv3.Client
}

func NewDiscover(client *clientv3.Client) *Discover {
	return &Discover{client: client}
}

func (d *Discover) DiscoverServices(serviceName string) ([]string, error) {
	keyPrefix := fmt.Sprintf("/services/%s", serviceName)
	resp, err := d.client.Get(context.Background(),keyPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	var services []string
	for  _, kv := range resp.Kvs {
		services = append(services, string(kv.Value))
	}
	return services, nil 
}
//监听服务变化，上线/下线 ，修改等， 确保调用最新的可用服务。
func (d *Discover) Watch(serviceName string, callback func([]string)) {
	keyPrefix := fmt.Sprintf("/services/%s", serviceName)
	watchChan := d.client.Watch(context.Background(), keyPrefix, clientv3.WithPrefix())

	go func() {
		for wresp := range watchChan {
			var services []string 
			for _,ev := range wresp.Events {
				services = append(services, string(ev.Kv.Value))
			}
			callback(services)
		}
	}()
}