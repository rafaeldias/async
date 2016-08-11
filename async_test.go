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
	e := Waterfall(Tasks{
		func(s *test, cb Callback) error {
			fmt.Println(s)
			return cb(1)
		},
		func(n int, cb Callback) error {
			fmt.Println(n)
			return cb(2, "String")
		},
		func(n2 int, s2 string) error {
			fmt.Println(n2, s2)
			return nil
		},
	}, &test{20})

	if e != nil {
		t.Errorf("Error executing a Waterfall (%q)", e)
	}
}

func TestAsyncError(t *testing.T) {
	e := Waterfall(Tasks{
		func(cb Callback) error {
			return cb(1)
		},
		func(n int, cb Callback) error {
			if n > 0 {
				return errors.New("Error on second function")
			}
			return cb()
		},
		func() error {
			fmt.Println("Function never reached")
			return nil
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
			func(cb Callback) error {
				return cb(1)
			},
			func(n int, cb Callback) error {
				fmt.Println(n)
				return cb()
			},
			func() error {
				fmt.Println("Last function")
				done <- true
				return nil
			},
		})
	}()

	go func() {
		Waterfall(Tasks{
			func(cb Callback) error {
				return cb(1)
			},
			func(n int, cb Callback) error {
				fmt.Println(n)
				time.Sleep(3 * time.Second)
				return cb()
			},
			func() error {
				fmt.Println("Last function 2")
				done <- true
				return nil
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
