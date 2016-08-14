package async

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

var (
	emptyStr    string
	emptyError  error
	emptyResult []interface{}
	emptyArgs   []reflect.Value
)

// Errors is a type of []error
// This is used to pass multiple errors when using parallel or concurrent methods
// and yet subscribe to the error interface
type Errors []error

func (e Errors) Error() string {
	b := bytes.NewBufferString(emptyStr)

	for _, err := range e {
		b.WriteString(err.Error())
		b.WriteString(" ")
	}

	return strings.TrimSpace(b.String())
}

// Type used as a list of tasks
type Tasks []interface{}

// funcs is the struct used to control the stack
// of functions to be executed.
type tasks struct {
	Stack []reflect.Value
}

// ExecInSeries executes recursively each task of the stack until it reachs
// the bottom of the stack or it is interrupted by an error
func (t *tasks) ExecInSeries(args ...reflect.Value) ([]interface{}, error) {
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

	// If function is expecting an `error`
	if lr > 0 && returnsError {
		// check if the error occured, if so returns the error and break the execution
		if e, ok := outArgs[lr-1].Interface().(error); ok {
			return emptyResult, e
		}
		lr = lr - 1
	}
	return t.ExecInSeries(outArgs[:lr]...)
}

// ExecInParallel executes all functions in the stack in Parallel.
func (t *tasks) ExecInParallel() error {
	var (
		errs Errors
		// Length of the tasks to execute
		ls = len(t.Stack)
		// Creates buffered channel for errors
		ce = make(chan error, ls)
		// Creates bufferd channel for controlling CPU usage and guarantee Paralellism
		climit = make(chan int, runtime.GOMAXPROCS(0))
	)

	for i := 0; i < ls; i++ {
		climit <- 1
		go execRoutine(t.Stack[i], ce)
		<-climit // discards the message, make space in buffer
	}

	// Consumes the errors from the channel
	for i := 0; i < ls; i++ {
		if e := <-ce; e != nil {
			errs = append(errs, e)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

func execRoutine(f reflect.Value, c chan error) {
	var (
		resErr error
		res    = f.Call(emptyArgs)
	)

	// Check if an error occurred
	if lr := len(res); lr > 0 {
		if e, ok := res[lr-1].Interface().(error); ok {
			resErr = e
		}
	}
	// Sends message to the channel
	c <- resErr
}

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
	var (
		err error
		t   *tasks = &tasks{}
	)

	// Checks if the Tasks passed are valid functions.
	t.Stack, err = validFuncs(stack)

	if err != nil {
		return err
	}
	return t.ExecInParallel()
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
