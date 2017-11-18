// Package circuit 实现了 Circuit-breaker pattern.
package circuit

import (
	"math/rand"
	"sync/atomic"
	"time"
	"sync"
)

// Breaker ...
type Breaker struct {
	option         *Options

	bucket         *Bucket
	TripperTime    int64 // 上次tripper时间
	HalfopenFail   int64 // 上次halfopen时失败次数
	consecFailures int64 // 连续错误次数

			     // breaker 状态
	tripper        int32 // 跳闸
	halfopen       int32

	mutex          sync.Mutex
}

// state breaker 状态
type state int

// state 状态值
const (
	tripper  state = iota
	halfopen state = iota
	closed   state = iota
)

// Options         配置项
// ErrRate         错误概率
// Sample          测试集
// ConsecFailTimes 连续失败次数
// Interval        Halfopen时间间隔(s)
// BucketTimeout   bucket失效时间(s)
type Options struct {
	ErrRate         float64
	Sample          int64
	ConsecFailTimes int64
	Interval        int64
	BucketTimeout   int64
}

// NewBreaker with default options
func NewBreaker() *Breaker{
	return NewBreakerWithOptions(nil)
}

// NewBreaker ...
func NewBreakerWithOptions(options *Options) *Breaker {
	if options == nil {
		options = &Options{
			ErrRate:         0.1,
			Sample:          100,
			ConsecFailTimes: 5,
			Interval:        5,
			BucketTimeout:   60,
		}
	}

	rand.Seed(time.Now().UnixNano())
	bucket := NewBucket(options.BucketTimeout)
	breaker := &Breaker{
		option: options,
		bucket: bucket,
	}
	breaker.Reset()
	return breaker
}

// Reset 重置所有计数器
func (b *Breaker) Reset() {
	atomic.StoreInt64(&b.HalfopenFail, 0)
	atomic.StoreInt32(&b.tripper, 0)
	atomic.StoreInt32(&b.halfopen, 0)
	b.bucket.Reset()
}

// Call 发送事件
func (b *Breaker) Call(val bool) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if val {
		b.Success()
	} else {
		b.Fail()
	}
	return
}

// Subscribe 每次请求判断status
func (b *Breaker) Subscribe() (status bool) {
	status = true
	// 判断是否为true 或者 是否halfopen,如果halfopen则50%可以请求
	b.Halfopen()
	if b.Halfopened() {
		x := rand.Intn(100)
		if x > 50 {
			status = false
			return
		} else {
			return
		}
	}
	if b.Tripped() {
		status = false
	}
	return
}

// Halfopened 是否开了一半
func (b *Breaker) Halfopened() bool {
	return atomic.LoadInt32(&b.halfopen) == 1
}

// Trip 跳闸
func (b *Breaker) Trip() {
	atomic.StoreInt32(&b.tripper, 1)
	now := time.Now()
	atomic.StoreInt64(&b.TripperTime, now.UnixNano())
}

// Tripped 是否跳闸
func (b *Breaker) Tripped() bool {
	return atomic.LoadInt32(&b.tripper) == 1
}

// Fail 记录失败
func (b *Breaker) Fail() {
	// 记录失败
	b.bucket.Fail()
	atomic.AddInt64(&b.consecFailures, 1)
	if b.ShouldTrip() {
		b.Trip()
	}

	// 试错阶段又失败，halfopen重新关闭
	if b.Tripped() && b.Halfopened() {
		atomic.AddInt64(&b.HalfopenFail, 1)
		atomic.StoreInt32(&b.halfopen, 0)
	}
}

//ShouldTrip 是否需要跳闸
func (b *Breaker) ShouldTrip() bool {
	// 失败概率大于x
	total := b.Successes() + b.Failures()
	if total >= b.option.Sample && b.bucket.ErrorRate() >= b.option.ErrRate {
		return true
	}

	// 连续失败大于10次
	if b.GetconsecFailures() >= b.option.ConsecFailTimes {
		return true
	}

	return false
}

// GetconsecFailures 获取consecFailures
func (b *Breaker) GetconsecFailures() int64 {
	return atomic.LoadInt64(&b.consecFailures)
}

// Successes 获取成功次数
func (b *Breaker) Successes() int64 {
	return b.bucket.Successes()
}

// Failures 获取成功次数
func (b *Breaker) Failures() int64 {
	return b.bucket.Failures()
}

// Success 计数success
func (b *Breaker) Success() {
	state := b.state()
	if state == halfopen {
		b.Reset()
	} else if state == closed {
		// 防止数据太多
		total := b.Successes() + b.Failures()
		if total >= b.option.Sample {
			b.Reset()
		}
	}
	atomic.StoreInt64(&b.consecFailures, 0)
	b.bucket.Success()
}

func (b *Breaker) state() state {
	if b.Halfopened() {
		return halfopen
	}

	if b.Tripped() {
		return tripper
	}
	return closed
}

// Halfopen ...
func (b *Breaker) Halfopen() {
	tripped := b.Tripped()
	if tripped {
		last := atomic.LoadInt64(&b.TripperTime)
		now := time.Now().UnixNano()
		// 已经经过了interval时间
		var alpha int64
		halfopenfails := atomic.LoadInt64(&b.HalfopenFail)
		if halfopenfails != 0 {
			alpha = halfopenfails
		}
		alpha += 1
		expire := time.Duration(b.option.Interval) * time.Second
		if now-last > int64(expire.Nanoseconds()*alpha) {
			atomic.StoreInt32(&b.halfopen, 1)
		}
	}
}

// ConsecFailures ...
func (b *Breaker) ConsecFailures() int64 {
	return b.consecFailures
}

// ErrorRate ...
func (b *Breaker) ErrorRate() float64 {
	return b.bucket.ErrorRate()
}
