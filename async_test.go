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
		func(s *test) (int, error) {
			fmt.Println(s)
			return 1, nil
		},
		func(n int) (int, string, error) {
			fmt.Println(n)
			return 2, "string", nil
		},
		func(n2 int, s2 string) (string, error) {
			fmt.Println(n2, s2)
			return "done", nil
		},
	}, &test{20})

	if e != nil {
		t.Errorf("Error executing a Waterfall (%q)", e)
	}

	fmt.Println(res)
}

func TestAsyncError(t *testing.T) {
	res, e := Waterfall(Tasks{
		func() (int, error) {
			return 1, nil
		},
		func(n int) error {
			fmt.Printf("if %d > 0 then error\n", n)
			if n > 0 {
				return errors.New("Error on second function")
			}
			return nil
		},
		func() error {
			fmt.Println("Function never reached")
			return nil
		},
	})

	if e != nil {
		fmt.Println("Error executing a Waterfall (%q)", e)
	}

	// should be empty
	fmt.Println(res)
}

func TestAsyncRoutine(t *testing.T) {
	var done = make(chan bool, 2)

	go func() {
		Waterfall(Tasks{
			func() (int, error) {
				return 1, nil
			},
			func(n int) error {
				fmt.Println(n)
				return nil
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
			func() (int, error) {
				return 1, nil
			},
			func(n int) error {
				fmt.Println(n)
				time.Sleep(3 * time.Second)
				return nil
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
