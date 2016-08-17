# async

async is a utility library for manipulating synchronous and asynchronous flows.

### Install

```bash
$ go get github.com/rafaeldias/async
```
### Concurrent
```go
async.Concurrent(tasks async.Tasks) error
```

Concurrent will execute all the functions concurrently using goroutines.

All errors ocurred in the functions will be returned. See [returning error](#returning-error) and [consuming error](#consuming-error).

- `tasks` is a slice of functions that will be executed concurrently.

### Parallel
```go
async.Parallel(tasks async.Tasks) error
```

Parallel will execute all the functions in parallel. It creates multiple goroutines and distributes the functions execution among them.
The number of goroutines  defaults to `runtime.GOMAXPROCS`. If the number of goroutines is equal to `GOMAXPROCS` and there're more functions to execute, these functions will wait until one of functions being executed finishes its job.

All errors occured in the functions will be returned. See [returning error](#returning-error) and [consuming error](#consuming-error).

- `tasks` is a slice of functions that will be executed in parallel. 

### Waterfall
```go
async.Waterfall(tasks async.Tasks, args ...interface{}) ([]interface[}, error)
```

Waterfall will execute all the functions in sequence, each returning their results to the next. If the last returning value of the function is of type `error`, then this value will not be passed to the next function, see [returning error](#returning-error) .

- `tasks` is a slice of functions that will be executed in series.
- `args` are optional parameters that will be passed to the first task.

Waterfall returns the results of the last task as a `[]interface{}` and `error`. 

### <a name="returning-error"></a>Returning error

If an error occur in any of the functions to be executed, the next function will not be executed, and the error will be returned to the caller.

In order for async to identify if an error occured, the error **must** be the last returning value of the function:

```go
_, err := async.Waterfall(async.Tasks{
        func () (int, error) {
                return 1, nil
        },
        // Function with error
        func (i int) (string, error) {
                if i > 0 {
                    // This line will interrupt the execution flow
                    return "", errors.New("Error occurred")
                }
                return "Ok", nil
        },
        // This function will not be executed.
        func (s string) {
            return
        }
});

if err != nil {
      fmt.Println(err.Error()); // "Error occurred"
}
```

### <a name="consuming-error"></a>Consuming errors

If errors occur in any function executed by `Concurrent` or `Parallel` an instance of `async.Errors` will be returned.
`async.Errors` implements the `error` interface, so in order to test if an error occurred, check if the returned error is not nil,
if it's not type cast it to `async.Errors`:

```go
err : = async.Parallel(func1, func2, ...funcN)

if err != nil {
        parallelErrors := err.(async.Errors)

        for _, e := range parallelErrors {
                fmt.Println(e.Error())
        }
}
```


# Examples 

### Waterfall

```go
import (
        "fmt"

        "github.com/rafaeldias/async"
)

func fib(p, c int) (int, int) {
  return c, p + c
}

func main() {

        // execution in series.
        res, e := async.Waterfall(async.Tasks{
                fib,
                fib,
                fib,
                func(p, c int) (int, error) {
                        return c, nil
                },
        }, 0, 1)

        if e != nil {
              fmt.Printf("Error executing a Waterfall (%s)\n", e.Error())
        }

        fmt.Println(res[0].(int)) // Prints 3
}

```

### Parallel

```go
import (
        "fmt"
        "time"

        "github.com/rafaeldias/async"
)

func main() {

        e = async.Parallel(async.Tasks{
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
                fmt.Printf("Errors [%s]\n", e.Error())
        }
}
```

### Concurrent

```go
import (
        "errors"
        "fmt"

        "github.com/rafaeldias/async"
)

func main() {

        e = async.Concurrent(async.Tasks{
                func() error {
                        for i := 'a'; i < 'a'+26; i++ {
                                fmt.Printf("%c ", i)
                        }
                        return nil
                },
                func() error {
                        time.Sleep(3 * time.Microsecond)
                        for i := 0; i < 27; i++ {
                                fmt.Printf("%d ", i)
                        }
                        return errors.New("Error executing concurently")
                },
        })

        if e != nil {
                fmt.Printf("Errors [%s]\n", e.Error()) // output errors separated by space
        }
}
```

# License :
Distributed under MIT License. See LICENSE file for more details.
