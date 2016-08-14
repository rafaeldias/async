# async

async is a utility library for manipulating synchronous and asynchronous flows.

### Install

```bash
$ go get github.com/rafaeldias/async
```

### Waterfall
Waterfall will execute all the functions in sequence, each returning their results to the next. If the last returning value of the function is of type `error`, then this value will not be passed to the next function, see [returning error](#returning-error) .

Signature:
```go
Waterfall(async.Tasks, ...interface{}) ([]interface[}, error)
```
- `async.Tasks` is a slice of functions that will be executed in series.
- `...interface{}` is optional parameters that will be passed to the first task.

Waterfall returns the results of the last task as a `[]interface{}` and `error`.



### <a name="returning-error"></a>Returning error

If an error occur in any of the functions to be executed, the next function will not be executed, and the error will be returned to the caller.

In order for async to identify if an error occured, the error **must** be the last returning value of the function:

```go
_, err := async.Waterfall(async.Tasks{
        func Task() (int, error) {
                return 1, nil
        },
        func TaskWithError(i int) (string, error) {
                if i > 0 {
                    // This line will interrupt the execution flow
                    return "", errors.New("Error occurred")
                }
                return "Ok", nil
        },
        // This function will not be executed.
        func TaskNeverReached(s string) {
            return
        }
});

if err != nil {
      fmt.Println(err.Error()); // "Error occurred"
}
```

### Example

```go
import (
        "fmt"
        "github.com/rafaeldias/async"
)

type test struct {
        ID uint
}

// execution in series.
res, e := async.Waterfall(async.Tasks{
        func(t *test) (int, error) {
                fmt.Println(t)
                return return 1, nil
        },
        func(n int) (int, string, error) {
                fmt.Println(n)
                return return 2, "String", nil
        },
        func(n2 int, s string) error {
                fmt.Println(n2, s)
                return nil
        },
}, &test{20})

if e != nil {
      fmt.Printf("Error executing a Waterfall (%v)\n", e)
}

```


# Todo :
- Implement `Parallel` with channels for concurrent executions.
