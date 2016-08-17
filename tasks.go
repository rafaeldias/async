package async

import (
	"bytes"
	"reflect"
	"runtime"
	"strings"
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
func (t *tasks) ExecConcurrent(parallel bool) error {
	var (
		errs Errors
		// Length of the tasks to execute
		ls = len(t.Stack)
		// Creates buffered channel for errors
		ce = make(chan error, ls)
		// Creates bufferd channel for controlling CPU usage and guarantee Paralellism
		sem = make(chan int, runtime.GOMAXPROCS(0))
	)

	// If parallel, tries to distribute the go routines among the cores, creating
	// at most `runtime.GOMAXPROCS` go routine.
	if parallel {
		for i := 0; i < ls; i++ {
			// Fill the buffered channel, if it gets full, go will block the execution
			// until any routine frees the channel
			sem <- 1 // the value doesn't matters
			go execRoutineParallel(t.Stack[i], ce, sem)
		}
	} else {
		for i := 0; i < ls; i++ {
			go execRoutine(t.Stack[i], ce)
		}
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

// Executes the task and consumes the message of `sem` channel
func execRoutineParallel(f reflect.Value, c chan error, sem chan int) {
	// execute routine
	execRoutine(f, c)

	// Once the task has done its job, consumes message from channel `sem`
	<-sem
}

// Executes the task and sends error to the `c` channel
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
	// Sends message to the error channel
	c <- resErr
}
