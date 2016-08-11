package async

import (
	"fmt"
	"reflect"
)

// Type used for callback
type Callback func(...interface{}) error

// Type used as a list of functions
type Functions []interface{}

// funcs is the struct used to control the stack
// of functions to be executed.
type funcs struct {
	Stack []reflect.Value
}

// executeNext executes recursively each function of
// the stack until it reachs the bottom of the stack or
// it is interrupted by an error or isn't called by one of the
// functions of the stack
func (f *funcs) executeNext(args ...interface{}) error {
	// end of stack, no need to proceed
	if len(f.Stack) == 0 {
		return nil
	}

	var (
		// true if function has the last argument of type `Callback`
		expectCallback bool
		// Prepare callback that will be passed to function if `expectCallback` is true
		next Callback = Callback(f.executeNext)
		// Get function to be executed
		fn reflect.Value = f.Stack[0]
		// Get type of the function to be executed
		fnt    reflect.Type = fn.Type()
                // Arguments to be sent to the function
		inArgs []reflect.Value
	)

	// If function expect any argument
	if l := fnt.NumIn(); l > 0 {
		// Get last argument of the function
		lastArg := fnt.In(l - 1)

		// Check if the last argument is a Callback
		expectCallback = lastArg.AssignableTo(reflect.TypeOf(next))
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
	f.Stack = f.Stack[1:len(f.Stack)]

	resArgs := fn.Call(inArgs)

	if lr := len(resArgs); lr > 0 {
		if e, ok := resArgs[lr-1].Interface().(error); ok {
			return e
		}
	}
	return nil
}

// Waterfall executes every function sequencially.
// The execution flow may be interrupted by not calling the  callback or returning an error.
// `firstArgs` is a slice of parameters to be passed to the first function of the stack.
func Waterfall(stack Functions, firstArgs ...interface{}) error {
	// Init stack of functions
	f := &funcs{}
	// Checks if arguments passed are valid functions.
	// If so, appends functions to `f.Stack`.
	for i := 0; i < len(stack); i++ {
		v := reflect.Indirect(reflect.ValueOf(stack[i]))

		if v.Kind() != reflect.Func {
			return fmt.Errorf("%T must be a Function ", v)
		}

		f.Stack = append(f.Stack, v)
	}

	return f.executeNext(firstArgs...)
}
