package async

import (
	"fmt"
	"reflect"
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
		return emptyResult, err
	}

	// transform interface{} to reflect.Value for execution
	for i := 0; i < len(firstArgs); i++ {
		args = append(args, reflect.ValueOf(firstArgs[i]))
	}

	return f.ExecInSeries(args...)
}

func Parallel(stack taskier) (Results, error) {
	return execConcurrentStack(stack, true)
}

func Concurrent(stack taskier) (Results, error) {
	return execConcurrentStack(stack, false)
}

func execConcurrentStack(stack taskier, parallel bool) (Results, error) {
	var (
		err error
		f   = &funcs{}
	)

	// Checks if the Tasks passed are valid functions.
	f.Stack, err = stack.GetFuncs()

	if err != nil {
		return nil, err
	}
	return f.ExecConcurrent(parallel)
}

// Loop through the stack of Tasks and check if they are valid functions.
// Returns the functions as []reflect.Value and error
func validFuncs(stack Tasks) ([]reflect.Value, error) {
	var (
		l  = len(stack)
		rf = make([]reflect.Value, l)
	)
	// Checks if arguments passed are valid functions.
	for i := 0; i < l; i++ {
		v := reflect.Indirect(reflect.ValueOf(stack[i]))

		if v.Kind() != reflect.Func {
			return rf, fmt.Errorf("%T must be a Function ", v)
		}

		rf = append(rf, v)
	}
	return rf, nil
}
