package tasklimit

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"testing"
	"time"
)

func TestNewLimitTask(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	type people struct {
		Name string
	}
	handler := func(data []byte) error {
		p := new(people)
		json.Unmarshal(data, p)
		fmt.Println(p.Name)
		return nil
	}
	taskLimiter := new(TaskLimit)
	taskLimiter.Init(
		WithRedisClient(client),
		WithTaskName("test"),
		WithHandler(handler),
		WithRate(10),
		WithCleanDuration(3*time.Second),
	)
	for i := 0; i < 1000; i++ {
		go func() {
			taskLimiter.Do(people{Name: "xch"})
		}()
	}
	c := make(chan int64)
	<-c
}

func Test01(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	count, err := client.LPush("name", struct{}{}).Result()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(count)
}
