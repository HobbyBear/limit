// 令牌桶 channel的实现
package tokenbucket

import "time"

type LimiterChannel struct {
	Rate         int // send rate token every second
	currentToken chan struct{}
	Capacity     int // the size of bucket
}

func New() *LimiterChannel {

	l := new(LimiterChannel)
	l.currentToken = make(chan struct{}, 16)
	go func() {
		t := time.NewTicker(time.Second)
		for range t.C {
			for i := 0; i < l.Rate; i++ {
				<-l.currentToken
			}
		}
	}()
	return l
}

func (l *LimiterChannel) Allow() bool {
	select {
	case l.currentToken <- struct{}{}:
		return true
	default:
		return false

	}
}
