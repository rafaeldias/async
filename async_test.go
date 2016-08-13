package async

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

type test struct {
	ID uint
}

func TestAsync(t *testing.T) {
	res, e := Waterfall(Tasks{
		func(s *test, cb Callback) {
			fmt.Println(s)
			cb(nil, 1)
		},
		func(n int, cb Callback) {
			fmt.Println(n)
			cb(nil, 2, "String")
		},
		func(n2 int, s2 string, cb Callback) {
			fmt.Println(n2, s2)
			cb(nil, n2, s2)
		},
	}, &test{20})

	if e != nil {
		t.Errorf("Error executing a Waterfall (%q)", e)
	}

	if len(res) > 0 {
		fmt.Println(res[0], res[1])
	}

}

func TestAsyncError(t *testing.T) {
	_, e := Waterfall(Tasks{
		func(cb Callback) {
			cb(nil, 1)
		},
		func(n int, cb Callback) {
			if n > 0 {
				cb(errors.New("Error on second function"))
				return
			}
			cb(nil)
		},
		func() {
			fmt.Println("Function never reached")
		},
	})

	if e != nil {
		fmt.Println("Error executing a Waterfall (%q)", e)
	}
}

func TestAsyncRoutine(t *testing.T) {
	var done = make(chan bool, 2)

	go func() {
		Waterfall(Tasks{
			func(cb Callback) {
				cb(nil, 1)
			},
			func(n int, cb Callback) {
				fmt.Println(nil, n)
				cb(nil)
			},
			func() {
				fmt.Println("Last function")
				done <- true
			},
		})
	}()

	go func() {
		Waterfall(Tasks{
			func(cb Callback) {
				cb(nil, 1)
			},
			func(n int, cb Callback) {
				fmt.Println(n)
				time.Sleep(3 * time.Second)
				cb(nil)
			},
			func() {
				fmt.Println("Last function 2")
				done <- true
			},
		})
	}()

	for i := 0; i < 2; i++ {
		select {
		case d := <-done:
			fmt.Println("done routine", d)
		}
	}
}
