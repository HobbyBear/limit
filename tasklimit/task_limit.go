package tasklimit

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"sync"
	"time"
)

type Handler func(x interface{}) error

type TaskLimit struct {
	client      *redis.Client
	rw          sync.RWMutex
	exitsWorker bool
	handler     Handler
	taskQueue   string
	limitQueue  string
	rate        int64
	lastTime    time.Time
	cleanTime   time.Duration
}

func NewLimitTask(client *redis.Client, taskName string, handler Handler, rate int64, clean time.Duration) *TaskLimit {
	return &TaskLimit{
		exitsWorker: false,
		handler:     handler,
		client:      client,
		taskQueue:   fmt.Sprintf("%s:task", taskName),
		limitQueue:  fmt.Sprintf("%s:limit", taskName),
		rate:        rate,
		cleanTime:   clean,
	}
}

func (t *TaskLimit) Do(task interface{}) {
	t.rw.RLock()
	if t.exitsWorker {
		t.rw.RUnlock()
		goto exitsWorker
	}
	t.rw.RUnlock()
	t.rw.Lock()
	t.exitsWorker = true
	t.lastTime = time.Now()
	t.rw.Unlock()
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if limit, _ := t.client.LLen(t.limitQueue).Result(); limit > t.rate {
				break
			}
			count, _ := t.client.LPush(t.limitQueue, struct{}{}).Result()
			if count == 1 {
				t.client.Expire(t.limitQueue, 1*time.Second)
			}
			dataStr, _ := t.client.RPop(t.taskQueue).Result()
			t.rw.RLock()
			lastRunningTime := t.lastTime
			t.rw.RUnlock()
			if dataStr == "" {
				if t.cleanTime < time.Since(lastRunningTime) {
					t.rw.Lock()
					t.exitsWorker = false
					t.rw.Unlock()
					return
				}
				break
			}
			t.rw.Lock()
			t.lastTime = time.Now()
			t.rw.Unlock()
			go func(dataStr string) {
				var x interface{}
				json.Unmarshal([]byte(dataStr), &x)
				t.handler(x)
			}(dataStr)
		}
	}()
exitsWorker:
	data, _ := json.Marshal(task)
	count, _ := t.client.LPush(t.taskQueue, data).Result()
	if count == 1 {
		t.client.Expire(t.taskQueue, 1*time.Second)
	}
}
