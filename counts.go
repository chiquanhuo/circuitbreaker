package circuit

import (
	"sync/atomic"
	"time"
)

type Bucket struct {
	lastAccess int64
	timeout    time.Duration
	failure    int64
	success    int64
}

func NewBucket(timeout time.Duration) *Bucket {
	return &Bucket{
		failure:    0,
		success:    0,
		lastAccess: time.Now().UnixNano(),
		timeout:    timeout,
	}
}

func (b *Bucket) Reset() {
	atomic.StoreInt64(&b.failure, 0)
	atomic.StoreInt64(&b.success, 0)
}

func (b *Bucket) Fail() {
	b.State()
	atomic.AddInt64(&b.failure, 1)
}

func (b *Bucket) Success() {
	b.State()
	atomic.AddInt64(&b.success, 1)
}

func (b *Bucket) ErrorRate() float64 {
	failure := atomic.LoadInt64(&b.failure)
	success := atomic.LoadInt64(&b.success)
	total := failure + success

	if total == 0 {
		return 0.0
	}

	return float64(failure) / float64(total)
}

func (b *Bucket) State() {
	elpased := time.Now().UnixNano() - b.lastAccess
	if elpased > b.timeout.Nanoseconds() {
		b.Reset()
	}
	b.lastAccess = time.Now().UnixNano()
}

func (b *Bucket) Failures() int64 {
	return atomic.LoadInt64(&b.failure)
}

func (b *Bucket) Successes() int64 {
	return atomic.LoadInt64(&b.success)
}
