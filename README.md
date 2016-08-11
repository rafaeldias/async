# async

async is a utility library for manipulating synchronous and asynchronous flows.

### Install

```bash
$ go get github.com/rafaeldias/async
```

### Waterfall
Waterfal function will execute all the functions in series. Usage:

Signature:
```
Waterfall(funcs async.Functions, args ...interface{}) error
```
- ***funcs*** is a list of `func` that will be executed in series.
- ***args*** is a list of optional parameters that will be passed to the first function.

### Synchronous Usage:
```go
import (
        "fmt"
        "github.com/rafaeldias/async"
)

type test struct {
  ID uint
}

// Syncrhonous execution in series.
e := async.Waterfall(async.Functions{
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
    fmt.Printf("Error executing a Waterfall (%q)\n", e)
}
```

### Ascynchronous usage:
```go
import (
        "fmt"
        "github.com/rafaeldias/async"
)
var done = make(chan bool, 2)

go func() {
        async.Waterfall(async.Functions{
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
        async.Waterfall(async.Functions{
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

```
# TODO :
- Implement `Parallel` with channels for concurrent executions.

