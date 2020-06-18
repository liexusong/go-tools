package golimiter

import (
	"fmt"
	"testing"
	"time"
)

func TestLimiter_Go(t *testing.T) {
	limiter := New(1000)

	for i := 0; i < 100; i++ {
		limiter.Go(func(args ...interface{}) {
			for i := 0; i < 5; i++ {
				fmt.Println(args)
				time.Sleep(time.Second)
			}
		}, i)
	}

	limiter.WaitAllJobsExited()
}
