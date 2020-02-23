package main

import "net/http"

var LB *LoadBalancer
func init() {
	LB = NewLoadBalancer()
	LB.AddServer("http://localhost:9000", 3)
	LB.AddServer("http://localhost:9001", 1)
	LB.AddServer("http://localhost:9002", 1)

	for index, server := range LB.Servers {
		LB.SumWeight += server.Weight
		for i := 0; i < server.Weight; i++ {
			LB.ServerIndices = append(LB.ServerIndices, index)
		}
	}

	// 健康检查
	go func() {
		LB.CheckServers()
	}()
}

func main() {
	http.ListenAndServe(":8000", LoadBalancer{})
}
