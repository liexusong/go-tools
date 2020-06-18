package golimiter

import (
	"container/list"
	"sync"
	"sync/atomic"
)

type Fn func(args ...interface{})

type Job struct {
	fn   Fn
	args []interface{}
}

type Limiter struct {
	cond    *sync.Cond
	jobs    *list.List
	maxSize int64
	counter int64
}

func New(maxSize int64) *Limiter {
	limiter := &Limiter{
		cond:    sync.NewCond(&sync.Mutex{}),
		jobs:    list.New(),
		maxSize: maxSize,
	}

	var notify = make(chan struct{}, 1)

	go limiter.runJobs(notify)

	<-notify // Waiting for runJobs routine running

	return limiter
}

func (l *Limiter) runJobs(c chan struct{}) {
	c <- struct{}{}

	for {
	again:
		l.cond.L.Lock()

		if l.counter < l.maxSize && l.jobs.Len() > 0 {
			elm := l.jobs.Front()
			job := elm.Value.(*Job)

			l.jobs.Remove(elm)

			if job != nil {
				atomic.AddInt64(&l.counter, 1)

				go func() {
					job.fn(job.args...) // Call job function

					l.cond.L.Lock()
					atomic.AddInt64(&l.counter, -1)
					l.cond.Broadcast()
					l.cond.L.Unlock()
				}()
			}

			l.cond.L.Unlock()

			goto again
		}

		l.cond.Wait() // Release locker and waiting for signal

		l.cond.L.Unlock()
	}
}

func (l *Limiter) Go(fn Fn, args ...interface{}) {
	job := &Job{
		fn:   fn,
		args: args,
	}

	l.cond.L.Lock()
	l.jobs.PushBack(job)
	l.cond.Broadcast()
	l.cond.L.Unlock()
}

func (l *Limiter) WaitAllJobsExited() {
	l.cond.L.Lock()
	for atomic.LoadInt64(&l.counter) > 0 || l.jobs.Len() > 0 {
		l.cond.Wait()
	}
	l.cond.L.Unlock()
}
