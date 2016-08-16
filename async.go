package async

import (
	"fmt"
	"reflect"
)

// Waterfall executes every task sequencially.
// The execution flow may be interrupted by returning an error.
// `firstArgs` is a slice of parameters to be passed to the first task of the stack.
func Waterfall(stack Tasks, firstArgs ...interface{}) ([]interface{}, error) {
	var (
		err  error
		args []reflect.Value
		t    = &tasks{}
	)
	// Checks if the Tasks passed are valid functions.
	t.Stack, err = validFuncs(stack)

	if err != nil {
		return emptyResult, err
	}

	// transform interface{} to reflect.Value for execution
	for i := 0; i < len(firstArgs); i++ {
		args = append(args, reflect.ValueOf(firstArgs[i]))
	}

	return t.ExecInSeries(args...)
}

func Parallel(stack Tasks) error {
	return execConcurrentStack(stack, true)
}

func Concurrent(stack Tasks) error {
	return execConcurrentStack(stack, false)
}

func execConcurrentStack(stack Tasks, parallel bool) error {
	var (
		err error
		t   *tasks = &tasks{}
	)

	// Checks if the Tasks passed are valid functions.
	t.Stack, err = validFuncs(stack)

	if err != nil {
		return err
	}
	return t.ExecConcurrent(parallel)
}

// Loop through the stack of Tasks and check if they are valid functions.
// Returns the functions as []reflect.Value and error
func validFuncs(stack Tasks) ([]reflect.Value, error) {
	var rf []reflect.Value
	// Checks if arguments passed are valid functions.
	for i := 0; i < len(stack); i++ {
		v := reflect.Indirect(reflect.ValueOf(stack[i]))

		if v.Kind() != reflect.Func {
			return rf, fmt.Errorf("%T must be a Function ", v)
		}

		rf = append(rf, v)
	}
	return rf, nil
}
