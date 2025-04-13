package main 

import (
	etcd "github.com/Vinoctis/etcd-register-center/sdk"
	"log"
	"fmt"
	"time"
)

func main() {
	endpoints := []string{"localhost:2379"}
	client, err := etcd.NewEtcdClient(endpoints)
	if err != nil {
		log.Fatal("Failed to create etcd client: %v", err)
	}
	fmt.Println("Etcd client created successfully", client)

	serviceName := "payment-grpc"
	address := "localhost:9090"
	ttl := int64(10)

	//注册服务
	registry := etcd.NewRegistry(client)
	if err := registry.Register(serviceName, address, ttl); err!= nil {
		log.Fatal("Failed to register service: %v", err)
	}
	fmt.Println("Service registered successfully")
	
	//监听服务变化
	discover := etcd.NewDiscover(client)
	services, err := discover.DiscoverServices(serviceName)
	if err!= nil {
		log.Fatal("Failed to discover services: %v", err)
	}
	fmt.Println("Services: ", services)
	//监听服务变化
	discover.Watch(serviceName, func(updateServices []string){
		fmt.Println("Updated services: ", updateServices)
	})

	//模拟服务下线
	go func() {
		time.Sleep(5*time.Second)
		registry.Deregister(serviceName, address)
		fmt.Println("Service deregistered successfully")
	}()
	
	select {}
}