package tasklimit

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis"
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
	taskLimiter := new(Limit)
	taskLimiter.Init(
		WithRedisClient(client),
		WithTaskName("test"),
		WithHandler(handler),
		WithRate(100),
		WithCleanDuration(3*time.Second),
	)
	var wg sync.WaitGroup
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			wg.Done()
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
	type person struct {
		Name string
		Age  int
	}
	//p := person{
	//	Name: "xch",
	//	Age:  16,
	//}

	count, err := client.BRPop(1*time.Second, "source", "source1").Result()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(count)
}
