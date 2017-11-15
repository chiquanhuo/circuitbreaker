## circuitbreaker

Circuitbreaker 使用`Golang`实现退避算法。

Circuitbreaker 可以在项目中需要依赖第三方库时使用。项目一般会因为链接第三方库超时或者报错而无法使用。
如果你的项目中需要请求多次第三方库，Circuitbreaker可以监听第三方库是否报错或者超时.
当达到一定数量的错误时，Circuitbreaker会跳闸，将来的执行将避免远程请求第三方库并直接返回错误。
同时，Circuitbreaker 将定期允许一些呼叫再次尝试请求第三方库，如果这些请求成功，则关闭Circuitbreaker。

### Installation

```bash
git clone git.github.com/chiquanhuo/circuitbreaker
```

### Example

以下是一个简单的例子，也可以参考`handle.go`

```Go
/*
 * error_rate(错误概率):     0.1
 * minSample(最小测试集):100
 * consecFails(连续错误数): 5
 * interval(尝试请求时间):   5 sec
 */
 
breaker := circuit.NewBreaker(0.1, 100, 5, time.Duration(5 * time.Second))


breaker.Subscribe() // 监听错误


if breaker.GetStatus() {
        // 继续进行请求第三方库
}


breaker.Call(bool) // 写入请求第三方库结果
```

