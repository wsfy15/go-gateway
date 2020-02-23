# go实现网关

## 负载均衡

通过内置的`SingleHostReverseProxy`实现反向代理，实现了下列负载均衡算法：

- 随机法：`SelectByRand`

- 根据IP Hash确定节点：`SelectByIPHash`，保证来自同一个客户端的请求往同一台后台服务器发送，实现sesson sticky

- 加权随机法：切片实现：`SelectByWeightRand`，区间法实现：`SelectByWeightRand2`

- 简单轮询：`RoundRobin`

- 加权轮询：切片实现：`RoundRobinByWeight`  区间法实现：`RoundRobinByWeight2`，带健康检查

- 平滑加权轮询：`RoundRobinByWeight3`，不会把请求一下子都发送给同一台服务器。

  > ```
  > 例如三台服务器权重比a:b:c=3:1:1 则平滑后不会出现连续3次a之后才b、c，而是abaca的顺序
  > ```

- 平滑加权轮询，带健康检查：`RoundRobinByWeight4`

上述算法实现位于`src/LBProxy/LB.go`。



## 健康检查

实现位于`src/LBProxy/HttpChecker.go`。

通过定期向url发送head请求，根据响应码和响应时间判断服务是否可用。不使用get或post请求，因为head请求只返回头部，传输量小。

节点有两个状态：UP 和 DOWN。

如果异常，不会仅一次异常就直接标记不可用，而是有一个阈值`FailMax`。

**快速恢复**：节点处于DOWN状态时，连续m次成功则立即将其转为UP状态。通常m小与`FailMax`。

