# limit
限流

tokenBucket 是go对令牌桶的实现
taskLimit是对需要异步执行的任务做限速处理
## taskLimit
```go
var taskLimiter = new(tasklimit.Limit)

func A(c *gin.Context){
// ...
client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	type people struct {
		Name string
	}
 
  // 需要异步处理的任务
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
		WithRate(100),  // 一秒最多执行100次任务
		WithCleanDuration(3*time.Second),  // 3秒后没有新任务到达 则释放工作go程
	)
  taskLimiter.Do(people{Name: c.Param("name")})
  // ....
}
```
