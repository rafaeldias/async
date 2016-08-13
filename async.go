package async

import (
	"fmt"
	"reflect"
)

var emptyResult []interface{}

// Type used for callback
type Callback func(error, ...interface{})

// Type used as a list of tasks
type Tasks []interface{}

// tasks is the struct used to control the stack
// of functions to be executed.
type tasks struct {
	Stack      []reflect.Value
	lastResult []interface{}
	execError  error
}

func (t *tasks) GetError() error {
	return t.execError
}

func (t *tasks) GetLastResult() []interface{} {
	return t.lastResult

}

// executeNext executes recursively each task of
// the stack until it reachs the bottom of the stack or
// it is interrupted by an error or isn't called by one of the
// tasks of the stack
func (t *tasks) executeNext(e error, args ...interface{}) {
	var (
		// true if function has the last argument of type `Callback`
		expectCallback bool
		// callback that will be passed to function if `expectCallback` is true
		next Callback
		// function to be executed
		fn reflect.Value
		// type of the function to be executed
		fnt reflect.Type
		// arguments to be sent to the function
		inArgs []reflect.Value
	)

	// if and error occurred, stop calling the tasks of the stack
	// and returns to the caller
	if e != nil {
		t.execError = e
		return
	}

	// end of stack, no need to proceed
	if len(t.Stack) == 0 {
		t.lastResult = args
		return
	}

	// Prepare callback that will be passed to function if `expectCallback` is true
	next = Callback(t.executeNext)
	// Get function to be executed
	fn = t.Stack[0]
	// Get type of the function to be executed
	fnt = fn.Type()

	// If function expect any argument
	if l := fnt.NumIn(); l > 0 {
		// Get last argument of the function
		lastArg := fnt.In(l - 1)

		// Check if the last argument is a Callback
		_, expectCallback = reflect.Zero(lastArg).Interface().(Callback)
	}

	// Transform arguments from interface{} to reflect.Value
	for i := 0; i < len(args); i++ {
		inArgs = append(inArgs, reflect.ValueOf(args[i]))
	}

	// If function is expecting a `Callback`, append next to the function arguments
	if expectCallback {
		inArgs = append(inArgs, reflect.ValueOf(next))
	}

	// Remove current function from the stack
	t.Stack = t.Stack[1:len(t.Stack)]

	fn.Call(inArgs)
}

// Waterfall executes every task sequencially.
// The execution flow may be interrupted by not calling the  callback or returning an error.
// `firstArgs` is a slice of parameters to be passed to the first task of the stack.
func Waterfall(stack Tasks, firstArgs ...interface{}) ([]interface{}, error) {
	// Init stack of tasks
	t := &tasks{}
	// Checks if arguments passed are valid functions.
	// If so, appends functions to `t.Stack`.
	for i := 0; i < len(stack); i++ {
		v := reflect.Indirect(reflect.ValueOf(stack[i]))

		if v.Kind() != reflect.Func {
			return emptyResult, fmt.Errorf("%T must be a Function ", v)
		}

		t.Stack = append(t.Stack, v)
	}

	t.executeNext(nil, firstArgs...)

	if e := t.GetError(); e != nil {
		return emptyResult, e
	}

	return t.GetLastResult(), nil
}
