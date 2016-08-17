package async

import (
	"errors"
	"fmt"
	"runtime"
	"testing"
	"time"
)

func fib(p, c int) (int, int) {
	return c, p + c
}

func TestAsync(t *testing.T) {
	var (
		e        error
		multiRes Results
		res      []interface{}
	)

	fmt.Println("Testing `Waterfall`")

	res, e = Waterfall(Tasks{
		fib, fib, fib,
		func(p, c int) int {
			return c
		},
	}, 0, 1)

	if e != nil {
		t.Errorf("Error executing a Waterfall (%s)", e.Error())
	}

	fmt.Println("Waterfall result :", res[0].(int))

	fmt.Printf("\nTesting `Parallel` with `runtime.GOMAXPROCS(2)`\n")

	runtime.GOMAXPROCS(2)

	multiRes, e = Parallel(Tasks{
		func() {
			for i := 'a'; i < 'a'+26; i++ {
				fmt.Printf("%c ", i)
			}
		},
		func() {
			time.Sleep(2 * time.Microsecond)
			for i := 0; i < 27; i++ {
				fmt.Printf("%d ", i)
			}
		},
		func() {
			for i := 'z'; i >= 'a'; i-- {
				fmt.Printf("%c ", i)
			}
		},
	})

	if e != nil {
		t.Errorf("Error executing a Waterfall (%s)", e.Error())
	}

	fmt.Println("Parallel Result", multiRes.Len())

	/*runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Printf("\nTesting `Concurrent`\n")

	e = Concurrent(Tasks{
		func() {
			for i := 'a'; i < 'a'+26; i++ {
				fmt.Printf("%c ", i)
			}
		},
		func() {
			time.Sleep(3 * time.Microsecond)
			for i := 0; i < 27; i++ {
				fmt.Printf("%d ", i)
			}
		},
		func() {
			for i := 'z'; i >= 'a'; i-- {
				fmt.Printf("%c ", i)
			}
		},
	})

	if e != nil {
		t.Errorf("Error executing a Waterfall (%s)", e.Error())
	}*/
}

func TestAsyncError(t *testing.T) {
	fmt.Printf("\nTesting `Waterfall` with error\n")

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
