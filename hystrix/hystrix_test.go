package hystrix

import (
	"errors"
	"github.com/afex/hystrix-go/hystrix"
	"testing"
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

type Counter struct {
	bucket map[int64]int
}
