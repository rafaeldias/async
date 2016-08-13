package async

import (
	"fmt"
	"reflect"
)

var (
	emptyError  error
	emptyResult []interface{}
)

// Type used as a list of tasks
type Tasks []interface{}

// funcs is the struct used to control the stack
// of functions to be executed.
type tasks struct {
	Stack []reflect.Value
}

// executeInSeries executes recursively each task of the stack until it reachs
// the bottom of the stack or it is interrupted by an error
func (t *tasks) executeInSeries(args ...reflect.Value) ([]interface{}, error) {
	var (
		// true if function has the last return value is of type `error`
		returnsError bool
		// Get function to be executed
		fn reflect.Value
		// Get type of the function to be executed
		fnt reflect.Type
		// Parameters to be sent to the next function
		outArgs []reflect.Value
	)

	// end of stack, no need to proceed
	if len(t.Stack) == 0 {
		result := emptyResult
		if l := len(args); l > 0 {
			for i := 0; i < l; i++ {
				result = append(result, args[i].Interface())
			}
		}
		return result, nil
	}

	// Get function to be executed
	fn = t.Stack[0]
	// Get type of the function to be executed
	fnt = fn.Type()

	// If function expect any argument
	if l := fnt.NumOut(); l > 0 {
		// Get last argument of the function
		lastArg := fnt.Out(l - 1)

		// Check if the last argument is a error
		returnsError = reflect.Zero(lastArg).Interface() == emptyError
	}

	// Remove current function from the stack
	t.Stack = t.Stack[1:len(t.Stack)]

	outArgs = fn.Call(args)

	lr := len(outArgs)

	if lr > 0 {
		// If function is expecting an `error`
		if returnsError {
			// check if the error occured, if so returns the error and break the execution
			if e, ok := outArgs[lr-1].Interface().(error); ok {
				return emptyResult, e
			}
			lr = lr - 1
		}
	}
	return t.executeInSeries(outArgs[:lr]...)
}

// Waterfall executes every task sequencially.
// The execution flow may be interrupted by returning an error.
// `firstArgs` is a slice of parameters to be passed to the first task of the stack.
func Waterfall(stack Tasks, firstArgs ...interface{}) ([]interface{}, error) {
	var (
		i    int
		args []reflect.Value
		t    *tasks = &tasks{}
	)
	// Checks if arguments passed are valid functions.
	for ; i < len(stack); i++ {
		v := reflect.Indirect(reflect.ValueOf(stack[i]))

		if v.Kind() != reflect.Func {
			return emptyResult, fmt.Errorf("%T must be a Function ", v)
		}

		t.Stack = append(t.Stack, v)
	}

	// transform interface{} to reflect.Value for execution
	for i = 0; i < len(firstArgs); i++ {
		args = append(args, reflect.ValueOf(firstArgs[i]))
	}

	return t.executeInSeries(args...)
}
