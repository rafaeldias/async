package async

import (
	"fmt"
	"reflect"
	"sync"
)

// Interface fot async to handle user user functions
type taskier interface {
	GetFuncs() (interface{}, error)
}

// Type used as a slisce of tasks
type Tasks []interface{}

func (t Tasks) GetFuncs() (interface{}, error) {
	var (
		l   = len(t)
		fns = make([]reflect.Value, l)
	)

	for i := 0; i < l; i++ {
		f := reflect.Indirect(reflect.ValueOf(t[i]))

		if f.Kind() != reflect.Func {
			return fns, fmt.Errorf("%T must be a Function ", f)
		}

		fns[i] = f
	}

	return fns, nil
}

func (t Tasks) IndexOf(val interface{}) int {

	for k, v := range t {
		if v == val {
			return k
		}
	}

	return -1
}

// Type used as a map of tasks
type MapTasks map[string]interface{}

func (mt MapTasks) GetFuncs() (interface{}, error) {
	var fns = map[string]reflect.Value{}

	for k, v := range mt {
		f := reflect.Indirect(reflect.ValueOf(v))

		if f.Kind() != reflect.Func {
			return fns, fmt.Errorf("%T must be a Function ", f)
		}

		fns[k] = f
	}

	return fns, nil
}

// Waterfall executes every task sequencially.
// The execution flow may be interrupted by returning an error.
// `firstArgs` is a slice of parameters to be passed to the first task of the stack.
func Waterfall(stack Tasks, firstArgs ...interface{}) ([]interface{}, error) {
	var (
		err  error
		args []reflect.Value
		f    = &funcs{}
	)
	// Checks if the Tasks passed are valid functions.
	f.Stack, err = stack.GetFuncs()

	if err != nil {
		panic(err.Error())
	}

	// transform interface{} to reflect.Value for execution
	for i := 0; i < len(firstArgs); i++ {
		args = append(args, reflect.ValueOf(firstArgs[i]))
	}

	return f.ExecInSeries(args...)
}

type autoTask struct {
	HasDependency bool
	Fn            reflect.Value
	Key           string
}

func Auto(stack MapTasks) (Results, error) {
	var (
		counter int
		err     error

		dependents     = map[string][]*sync.WaitGroup{}
		mutex          = &sync.Mutex{}
		readyTasks     = make(chan autoTask, len(stack))
		readyToCheck   []string
		remainingTasks = len(stack)
		results        = mapResults{}
		uncheckedDeps  = map[string]int{}
	)

	for k, v := range stack {
		fn := reflect.Indirect(reflect.ValueOf(v))

		if t, ok := v.(Tasks); ok {
			var (
				wg = &sync.WaitGroup{}

				l    = len(t)
				deps = t[:l-1]
				ld   = len(deps)
				f    = reflect.Indirect(reflect.ValueOf(t[l-1]))
			)

			// No dependencies for this task
			if ld == 0 {
				if f.Kind() != reflect.Func {
					panic("Invalid type sent to Auto")
				}
				readyToCheck = append(readyToCheck, k)
				readyTasks <- autoTask{Fn: f, Key: k}
			}

			wg.Add(ld)

			// Checks if dependencies exist in tasks list
			for _, d := range deps {
				depName := d.(string)
				if _, ok := stack[depName]; !ok {
					panic(fmt.Sprintf("Task '%s' has non-existent dependency in %v", k, deps))
				}
				dependents[depName] = append(dependents[depName], wg)
			}

			uncheckedDeps[k] = ld

			go func(w *sync.WaitGroup, task reflect.Value, key string) {
				w.Wait()

				mutex.Lock()
				if err != nil {
					mutex.Unlock()
					return
				}
				mutex.Unlock()

				readyTasks <- autoTask{Fn: task, HasDependency: true, Key: key}
			}(wg, f, k)
		} else if fn.Kind() == reflect.Func { // There's no dependecies
			readyToCheck = append(readyToCheck, k)
			readyTasks <- autoTask{Fn: fn, Key: k}
		} else {
			panic("Invalid type sent to Auto")
		}
	}

	// Checks for deadlocks.
	for len(readyToCheck) > 0 {
		var (
			tasks       []string
			currentTask string
		)

		lrtc := len(readyToCheck)
		currentTask, readyToCheck = readyToCheck[lrtc-1], readyToCheck[:lrtc-1]
		counter = counter + 1

		for k, t := range stack {
			if ts, ok := t.(Tasks); ok && ts.IndexOf(currentTask) != -1 {
				tasks = append(tasks, k)
			}
		}

		for _, t := range tasks {
			if uncheckedDeps[t] = uncheckedDeps[t] - 1; uncheckedDeps[t] == 0 {
				readyToCheck = append(readyToCheck, t)
			}
		}
	}

	if counter != remainingTasks {
		panic("Auto cannot execute tasks due to recursive dependency")
	}

	for task := range readyTasks {

		go func(at autoTask) {
			var (
				args []reflect.Value
				res  []reflect.Value
			)

			defer mutex.Unlock()

			if at.HasDependency {
				args = append(args, reflect.ValueOf(results))
			}

			res = at.Fn.Call(args)

			// Get type of the function to be executed
			fnt := at.Fn.Type()

			mutex.Lock()

			// Checks is error occurred
			if err == nil {

				// Check if function returns any value
				if l := fnt.NumOut(); l > 0 {
					// Gets last return value type of the function
					lastArg := fnt.Out(l - 1)

					// Check if the last return value is error
					if reflect.Zero(lastArg).Interface() == emptyError {
						// Check if error occurred and interrupt the flow
						if e, ok := res[l-1].Interface().(error); ok {
							err = e
							for _, d := range dependents {
								for _, wg := range d {
									wg.Done()
								}
							}
							close(readyTasks)
							return
						}
						// Decrements l so the results returned doesn't contain the error
						l = l - 1
					}

					// If no error occurred, fills the exr.results
					if l > 0 {
						for _, v := range res[:l] {
							results[at.Key] = append(results[at.Key], v.Interface())
						}
					}
				}

				if wgs, ok := dependents[at.Key]; ok {
					for _, wg := range wgs {
						wg.Done()
					}
					delete(dependents, at.Key)
				}

				remainingTasks = remainingTasks - 1

				if remainingTasks == 0 {
					close(readyTasks)
				}

			}
		}(task)

	}

	return results, err
}

func Concurrent(stack taskier) (Results, error) {
	return execConcurrentStack(stack, false, false)
}

func Parallel(stack taskier) (Results, error) {
	return execConcurrentStack(stack, true, false)
}

func Race(stack taskier) (Results, error) {
	return execConcurrentStack(stack, false, true)
}

func execConcurrentStack(stack taskier, parallel, race bool) (Results, error) {
	var (
		err error
		f   = &funcs{}
	)

	// Checks if the Tasks passed are valid functions.
	f.Stack, err = stack.GetFuncs()

	if err != nil {
		panic(err)
	}
	return f.ExecConcurrent(parallel, race)
}
