package hystrix

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/afex/hystrix-go/hystrix"
)

func Test01(t *testing.T) {
	err := hystrix.Do("my_command", func() error {
		// talk to other services
		return errors.New("测试断路器")
	}, func(e error) error {
		return errors.New("断路器打开")
	})
	if err != nil {
		t.Fatal(err)
	}

}

func Test02(t *testing.T) {
	var one sync.Once
	one.Do(func() {
		fmt.Println("hello")
	})
	one.Do(func() {
		fmt.Println("hello2")
	})
}

type Counter struct {
	buckets map[int64]*bucket
	Rw      sync.RWMutex
}

type bucket struct {
	Value int
}

func (c *Counter) Increment(i int) {
	if i == 0 {
		return
	}
	c.Rw.Lock()
	defer c.Rw.Unlock()
	bucket := c.getCurrentBucket()
	bucket.Value += i
	c.removeOldBucket()
}

// 过去10s内最新的计数值
func (c *Counter) Sum() int {
	now := time.Now().Unix() - 10
	sum := 0
	c.Rw.RLock()
	defer c.Rw.Unlock()
	for timestamp, bucket := range c.buckets {
		if timestamp >= now {
			sum += bucket.Value
		}
	}
	return sum
}

func (c *Counter) getCurrentBucket() *bucket {
	now := time.Now().Unix()
	var ok bool
	var b *bucket
	if b, ok = c.buckets[now]; !ok {
		b = new(bucket)
		c.buckets[now] = b
	}
	return b

}

func (c *Counter) removeOldBucket() {
	now := time.Now().Unix() - 10
	for timestamp := range c.buckets {
		if timestamp <= now {
			delete(c.buckets, timestamp)
		}
	}
}
