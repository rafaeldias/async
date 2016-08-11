# async

async is a utility library for manipulating synchronous and asynchronous flows.

### Install

```bash
$ go get github.com/rafaeldias/async
```

### Waterfall
Waterfal function will execute all the functions in series. Usage:

Signature:
```go
Waterfall(async.Tasks, ...interface{}) error
```
- `async.Tasks` is a list of tasks that will be executed in series.
- `...interface{}` is optional parameters that will be passed to the first task.

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
e := async.Waterfall(async.Tasks{
        func(s *test, cb async.Callback) error {
                fmt.Println(s)
                return cb(1)
        },
        func(n int, cb async.Callback) error {
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
        async.Waterfall(async.Tasks{
                func(cb async.Callback) error {
                        return cb(1)
                },
                func(n int, cb async.Callback) error {
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
        async.Waterfall(async.Tasks{
                func(cb async.Callback) error {
                        return cb(1)
                },
                func(n int, cb async.Callback) error {
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

