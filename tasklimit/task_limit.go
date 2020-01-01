package tasklimit

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"runtime/debug"
	"sync"
	"time"
)

type Handler func(data []byte) error

type Setter func(t *TaskLimit)

type TaskLimit struct {
	client        *redis.Client
	rw            sync.RWMutex
	once          sync.Once
	exitsWorker   bool
	handler       Handler
	taskQueue     string
	limitQueue    string
	rate          int64 // allow running rate every second
	lastTime      time.Time
	cleanDuration time.Duration
}

func (t *TaskLimit) Init(setters ...Setter) *TaskLimit {
	t.once.Do(func() {
		t.exitsWorker = false
		for _, setter := range setters {
			setter(t)
		}
	})
	return t
}

func WithTaskName(name string) Setter {
	return func(t *TaskLimit) {
		t.taskQueue = fmt.Sprintf("%s:task", name)
		t.limitQueue = fmt.Sprintf("%s:limit", name)
	}
}

func WithRate(rate int64) Setter {
	return func(t *TaskLimit) {
		t.rate = rate
	}
}

func WithCleanDuration(duration time.Duration) Setter {
	return func(t *TaskLimit) {
		t.cleanDuration = duration
	}
}

func WithRedisClient(client *redis.Client) Setter {
	return func(t *TaskLimit) {
		t.client = client
	}
}

func WithHandler(handler Handler) Setter {
	return func(t *TaskLimit) {
		t.handler = handler
	}
}

func (t *TaskLimit) Do(taskParam interface{}) error {
	data, err := json.Marshal(taskParam)
	if err != nil {
		return err
	}
	t.client.LPush(t.taskQueue, data)
	t.notifyWorker()
	return nil
}

func (t *TaskLimit) notifyWorker() {
	t.rw.RLock()
	if t.exitsWorker {
		t.rw.RUnlock()
		return
	}
	t.rw.RUnlock()
	t.rw.Lock()
	t.exitsWorker = true
	t.lastTime = time.Now()
	t.rw.Unlock()
	log.Println("create new go worker ...")
	goFunc(func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			for {
				limit, err := t.client.LLen(t.limitQueue).Result()
				if err != nil {
					break
				}
				if limit > t.rate {
					ttlTime, err := t.client.TTL(t.limitQueue).Result()
					if err != nil {
						break
					}
					if ttlTime < 0 {
						t.client.Expire(t.limitQueue, time.Second)
					}
					log.Println("tow many task todo ,please waiting ....")
					break
				}
				count, err := t.client.LPush(t.limitQueue, 0).Result()
				if err != nil {
					break
				}
				if count == 1 {
					t.client.Expire(t.limitQueue, time.Second)
				}
				dataStr, err := t.client.RPop(t.taskQueue).Result()
				if err != nil && err != redis.Nil {
					break
				}
				t.rw.RLock()
				lastRunningTime := t.lastTime
				t.rw.RUnlock()
				if dataStr == "" {
					if t.cleanDuration < time.Since(lastRunningTime) {
						log.Println("destroy go worker...")
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
				goFuncWithString(func(dataStr string) {
					err := t.handler([]byte(dataStr))
					if err != nil {
						log.Println("err ... ")
					}
				}, dataStr)
			}
		}

	})
}

func goFuncWithString(f func(data string), data string) {
	go func() {
		defer func() {
			if p := recover(); p != nil {
				stackInfo := debug.Stack()
				log.Println(string(stackInfo), p)
			}
		}()
		f(data)
	}()
}

func goFunc(f func()) {
	go func() {
		defer func() {
			if p := recover(); p != nil {
				stackInfo := debug.Stack()
				log.Println(string(stackInfo), p)
			}
		}()
		f()
	}()
}
