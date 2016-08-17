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

type Results interface {
	Index(int) []interface{}
	Key(string) []interface{}
	Len() int
	Keys() []string
}

type sliceResults [][]interface{}

func (s sliceResults) Index(i int) []interface{} {
	return s[i]
}

func (s sliceResults) Len() int {
	return len(s)
}

func (s sliceResults) Keys() []string {
	panic("Cannot get map keys from Slice")
}

func (s sliceResults) Key(k string) []interface{} {
	panic("Cannot get map key from Slice")
}

type mapResults map[string][]interface{}

func (m mapResults) Index(i int) []interface{} {
	panic("Cannot get index from Map")
}

func (m mapResults) Len() int {
	return len(m)
}

func (m mapResults) Keys() []string {
	var keys = make([]string, len(m))

	i := 0
	for k := range m {
		keys[i] = k
		i++
	}

	return keys
}

func (m mapResults) Key(k string) []interface{} {
	return m[k]
}

// Type used as a list of tasks
type Tasks []interface{}

// Type used as a map of tasks
type MapTasks map[string]interface{}

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
func (t *tasks) ExecConcurrent(parallel bool) (Results, error) {
	var (
		results Results
		errs    Errors
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

	//results = sliceResults{[]interface{}{1}}
	results = mapResults{"test": []interface{}{1}}

	if len(errs) == 0 {
		return results, nil
	}

	return results, errs
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
