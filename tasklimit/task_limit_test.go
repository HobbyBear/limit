package tasklimit

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

type Trainer struct {
	Name string
	Age  int
	City string
}

func Test02(t *testing.T) {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	//ash := Trainer{"Ash", 10, "Pallet Town"}
	collection := client.Database("test").Collection("trainers")
	//insertResult, err := collection.InsertOne(context.TODO(), ash)
	//if err != nil {
	//	log.Fatal(err)
	//}
	filter := bson.D{{"name", "Ash"}}
	var result Trainer

	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found a single document: %+v\n", result)
}
