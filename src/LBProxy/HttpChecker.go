package main

import (
	"math"
	"net/http"
)

// 通过定期向url发送head请求，根据响应码和响应时间判断服务是否可用
// 如果异常，不能仅一次异常就直接标记不可用，而是有一个阈值
// 节点恢复
type HttpChecker struct {
	servers HttpServers
	FailMax int // 失败阈值
	SuccessMax int // 连续成功阈值
	FailFactor float64 // 降权因子  默认为5.0
}

func NewHttpChecker(servers HttpServers, FailMax, SuccessMax int, FailFactor float64) *HttpChecker {
	return &HttpChecker{
		servers:      servers,
		FailMax:    FailMax,
		SuccessMax: SuccessMax,
		FailFactor: FailFactor,
	}
}

func(this *HttpChecker) Check() {
	client := http.Client{}
	for _, s := range this.servers {
		// 使用head请求，相比get、post更高效，因为只返回头部，传输量小
		res, err := client.Head(s.Host)
		if res != nil {
			defer res.Body.Close()
		}

		if err != nil {
			this.fail(s)
			continue
		}

		if res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusBadRequest {
			this.success(s)
		} else {
			this.fail(s)
		}
	}
}

func (this *HttpChecker) fail(s *HttpServer) {
	if s.FailCount == this.FailMax {
		s.Status = "DOWN"
	} else {
		s.FailCount++
	}
	s.SuccessCount = 0

	fw := int(math.Floor(float64(s.Weight)) / this.FailFactor)
	if fw == 0 {
		fw = 1
	}
	s.FailWeight += fw
	if s.FailWeight > s.Weight {
		s.FailWeight = s.Weight
	}
}

func (this *HttpChecker) success(s *HttpServer) {
	if s.FailCount > 0 {
		s.FailCount--
		s.SuccessCount++
	}

	if s.Status == "DOWN" {
		if s.FailCount == 0 || s.SuccessCount == this.SuccessMax {
			s.FailCount = 0
			s.Status = "UP"
			s.SuccessCount = 0
		}
	}

	s.FailWeight = 0 // 一旦成功 直接设为0 简单
}
