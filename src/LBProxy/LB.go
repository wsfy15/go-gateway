package main

import (
	"fmt"
	"hash/crc32"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"time"
)

type LoadBalancer struct {
	Servers       HttpServers
	ServerIndices []int	// 使用切片的方式表示加权随机、加权轮询
	CurIndex	int	// 轮询算法中 下一个使用的服务器下标
	SumWeight int // 总权重 用于平滑
}

type HttpServer struct {
	Host string
	Weight int
	CurWeight int // 当前权重 用于平滑
	FailWeight int // 失败权重 由Weight-FailWeight 得到加权轮询下的权重
	Status string // 节点状态UP DOWN
	FailCount int // 失败次数
	SuccessCount int // 连续成功次数，用于快速恢复
}

type HttpServers []*HttpServer
func (p HttpServers) Len() int           { return len(p) }
func (p HttpServers) Less(i, j int) bool { return p[i].CurWeight > p[j].CurWeight } // 降序排序
func (p HttpServers) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }


func NewHttpServer(host string, weight int) *HttpServer{
	return &HttpServer{Host: host, Weight: weight, CurWeight:0, Status: "UP"}
}

func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{Servers: make([]*HttpServer, 0)}
}

func(this *LoadBalancer) AddServer(host string, weight int) {
	server := NewHttpServer(host, weight)
	this.Servers = append(this.Servers, server)
}

// 判断所有节点是否都DOWN掉
func (this *LoadBalancer) IsAllDown() bool {
	for _, s := range this.Servers {
		if s.Status == "UP" {
			return false
		}
	}
	return true
}

// 负载均衡算法
func(this *LoadBalancer) SelectByRand() *HttpServer {
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(this.Servers))
	return this.Servers[index]
}

func(this *LoadBalancer) SelectByIPHash(ip string) *HttpServer{
	index := crc32.ChecksumIEEE([]byte(ip)) % uint32(len(this.Servers))
	return this.Servers[index]
}

// 加权随机
func(this *LoadBalancer) SelectByWeightRand() *HttpServer{
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(LB.ServerIndices))
	return LB.Servers[LB.ServerIndices[index]]
}

// 加权随机 改良版
func(this *LoadBalancer) SelectByWeightRand2() *HttpServer{
	// 假设A:B:C = 5:2:1 则划分为三个区间[0,5) [5,7) [7,8) 直接计算rand.Intn(8) 判断随机数落在哪个区间即可
	sumList := make([]int, len(LB.Servers))
	sum := 0
	for i, server := range LB.Servers {
		sum += server.Weight
		sumList[i] = sum
	}

	rand.Seed(time.Now().UnixNano())
	rad := rand.Intn(sum)
	for i, v := range sumList {
		if rad < v {
			return LB.Servers[i]
		}
	}
	return LB.Servers[0]
}

// 简单轮询
func (this *LoadBalancer) RoundRobin() *HttpServer{
	server := this.Servers[this.CurIndex]
	this.CurIndex = (this.CurIndex + 1) % len(this.Servers)
	// 如果全部都down掉了，就按原计划返回节点，依然照样轮询所有
	// 可以考虑返回一个错误页面
	if server.Status == "DOWN" && !this.IsAllDown() {
		return this.RoundRobin()
	}
	return server
}

// 加权轮询
func (this *LoadBalancer) RoundRobinByWeight() *HttpServer{
	server := this.Servers[this.ServerIndices[this.CurIndex]]
	this.CurIndex = (this.CurIndex + 1) % len(this.ServerIndices)
	return server
}

// 加权轮询 改良版 带健康检查
func (this *LoadBalancer) RoundRobinByWeight2() *HttpServer{
	server := this.Servers[0]
	sum := 0
	for i := 0; i < len(this.Servers); i++ {
		realWeight := this.Servers[i].Weight - this.Servers[i].FailWeight
		if realWeight == 0 {
			continue
		}

		sum += realWeight
		if this.CurIndex < sum {
			server = this.Servers[i]
			if this.CurIndex == sum - 1 && i == len(this.Servers) - 1 {
				this.CurIndex = 0
			} else {
				this.CurIndex++
			}
			break
		}
	}
	return server
}

// 平滑加权轮询 a:b:c=3:1:1 则平滑后不会出现连续3次a之后才b、c，而是abaca的顺序
func (this *LoadBalancer) RoundRobinByWeight3() *HttpServer {
	for _, s := range this.Servers {
		s.CurWeight += s.Weight	// 把当前权重加上原始权重
	}
	sort.Sort(this.Servers)
	maxServer := this.Servers[0]
	maxServer.CurWeight -= this.SumWeight	// 把命中节点的当前权重减去初始 总权重
	return maxServer
}

func (this *LoadBalancer) getTotalRealWeight() int {
	sum := 0
	for _, s := range this.Servers {
		sum += s.Weight - s.FailWeight
	}
	return sum
}

// 平滑加权轮询 带健康检查
func (this *LoadBalancer) RoundRobinByWeight4() *HttpServer {
	for _, s := range this.Servers {
		s.CurWeight += s.Weight - s.FailWeight	// 把当前权重加上真实权重
	}
	sort.Sort(this.Servers)
	maxServer := this.Servers[0]

	maxServer.CurWeight -= this.getTotalRealWeight() // 把命中节点的当前权重减去 真实总权重
	return maxServer
}

func(this LoadBalancer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
		}
	}()

	// 对于chrome浏览器，由于其会请求favicon，导致轮询算法效果与预期不同，所以无视该请求
	if request.URL.Path == "/favicon.ico" {
		return
	}

	//server := LB.SelectByRand()
	//server := LB.SelectByIPHash(request.RemoteAddr)
	//server := LB.SelectByWeightRand()
	//server := LB.SelectByWeightRand2()
	server := LB.RoundRobin()
	//server := LB.RoundRobinByWeight()
	//server := LB.RoundRobinByWeight2()
	//server := LB.RoundRobinByWeight3()

	parseUrl, _ := url.Parse(server.Host)
	proxy := httputil.NewSingleHostReverseProxy(parseUrl)
	proxy.ServeHTTP(writer, request)
}

func(this *LoadBalancer) CheckServers() {
	checker := NewHttpChecker(this.Servers, 4, 2, 5.0)
	t := time.NewTicker(3 * time.Second) // 每3秒做一次检查
	for {
		select {
		case <- t.C:
			checker.Check()
		}
		for _, s := range this.Servers {
			fmt.Println(s.Host, s.Status, s.FailCount, s.SuccessCount)
		}
		fmt.Println("---------------------------------")
	}
}