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

// Prints all errors from asynchronous tasks separated by space
func (e Errors) Error() string {
	b := bytes.NewBufferString(emptyStr)

	for _, err := range e {
		b.WriteString(err.Error())
		b.WriteString(" ")
	}

	return strings.TrimSpace(b.String())
}

// Results is an interface used to return sliceResult or mapResults
// from asynchronous tasks. It has methods that should be used to
// get data from the results.
type Results interface {
	Index(int) []interface{}  // Get value by index
	Key(string) []interface{} // Get value by key
	Len() int                 // Get the length of the result
	Keys() []string           // Get the keys of the result
}

// sliceResults is a slice of slice of interface used to return
// results from asynchronous tasks that were passed as slice.
type sliceResults [][]interface{}

// Returns the values returned from ith task
func (s sliceResults) Index(i int) []interface{} {
	return s[i]
}

// Returns the length of the results
func (s sliceResults) Len() int {
	return len(s)
}

// Not supported by sliceResults
func (s sliceResults) Keys() []string {
	panic("Cannot get map keys from Slice")
}

// Not supported by sliceResults
func (s sliceResults) Key(k string) []interface{} {
	panic("Cannot get map key from Slice")
}

// sliceResults is a map of string of slice of interface used to return
// results from asynchronous tasks that were passed as map of string.
type mapResults map[string][]interface{}

// Not supported by mapResults
func (m mapResults) Index(i int) []interface{} {
	panic("Cannot get index from Map")
}

// Returns the length of the results
func (m mapResults) Len() int {
	return len(m)
}

// Returns the keys of the result map
func (m mapResults) Keys() []string {
	var keys = make([]string, len(m))

	i := 0
	for k := range m {
		keys[i] = k
		i++
	}

	return keys
}

// Returns the result value by key
func (m mapResults) Key(k string) []interface{} {
	return m[k]
}

// Internal usage to gather results from tasks
type execResult struct {
	err     error
	results []reflect.Value
	key     string
}

// funcs is the struct used to control the stack
// of functions to be executed.
type funcs struct {
	Stack interface{}
}

// ExecInSeries executes recursively each task of the stack until it reachs
// the bottom of the stack or it is interrupted by an error
func (f *funcs) ExecInSeries(args ...reflect.Value) ([]interface{}, error) {
	var (
		fns          = f.Stack.([]reflect.Value)
		fnl          = len(fns)
		returnsError bool            // true if function has the last return value is of type `error`
		fn           reflect.Value   // Get function to be executed
		fnt          reflect.Type    // Get type of the function to be executed
		outArgs      []reflect.Value // Parameters to be sent to the next function
	)

	// end of stack, no need to proceed
	if fnl == 0 {
		result := emptyResult
		if l := len(args); l > 0 {
			for i := 0; i < l; i++ {
				result = append(result, args[i].Interface())
			}
		}
		return result, nil
	}

	// Get function to be executed
	fn = fns[0]
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
	f.Stack = fns[1:fnl]

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
	return f.ExecInSeries(outArgs[:lr]...)
}

// ExecInParallel executes all functions in the stack in Parallel.
func (f *funcs) ExecConcurrent(parallel bool) (Results, error) {
	var (
		results Results
		errs    Errors
	)

	if funcs, ok := f.Stack.([]reflect.Value); ok {
		results, errs = execSlice(funcs, parallel)
	} else if mapFuncs, mapOk := f.Stack.(map[string]reflect.Value); mapOk {
		results, errs = execMap(mapFuncs, parallel)
	} else {
		// Incorret t.Stack type
		panic("Stack type must be of type []reflect.Value or map[string]reflect.Value.")
	}

	if len(errs) == 0 {
		return results, nil
	}

	return results, errs
}

func execSlice(funcs []reflect.Value, parallel bool) (sliceResults, Errors) {
	var (
		errs    Errors
		results = sliceResults{}
		ls      = len(funcs)                // Length of the functions to execute
		cr      = make(chan execResult, ls) // Creates buffered channel for errors
	)
	// If parallel, tries to distribute the go routines among the cores, creating
	// at most `runtime.GOMAXPROCS` go routine.
	if parallel {
		sem := make(chan int, runtime.GOMAXPROCS(0)) // Creates bufferd channel for controlling CPU usage and guarantee Paralellism
		for i := 0; i < ls; i++ {
			// Fill the buffered channel, if it gets full, go will block the execution
			// until any routine frees the channel
			sem <- 1 // the value doesn't matter
			go execRoutineParallel(funcs[i], cr, sem, emptyStr)
		}
	} else {
		for i := 0; i < ls; i++ {
			go execRoutine(funcs[i], cr, emptyStr)
		}
	}

	// Consumes the results from the channel
	for i := 0; i < ls; i++ {
		r := <-cr

		if r.err != nil {
			errs = append(errs, r.err)
		} else if lcr := len(r.results); lcr > 0 {
			res := make([]interface{}, lcr)
			for j, v := range r.results {
				res[j] = v.Interface()
			}
			results = append(results, res)
		}
	}

	return results, errs
}

func execMap(funcs map[string]reflect.Value, parallel bool) (mapResults, Errors) {
	var (
		errs    Errors
		results = mapResults{}
		ls      = len(funcs)                // Length of the functions to execute
		cr      = make(chan execResult, ls) // Creates buffered channel for errors
	)

	// If parallel, tries to distribute the go routines among the cores, creating
	// at most `runtime.GOMAXPROCS` go routine.
	if parallel {
		sem := make(chan int, runtime.GOMAXPROCS(0)) // Creates bufferd channel for controlling CPU usage and guarantee Paralellism
		for k, f := range funcs {
			// Fill the buffered channel, if it gets full, go will block the execution
			// until any routine frees the channel
			sem <- 1 // the value doesn't matter
			go execRoutineParallel(f, cr, sem, k)
		}
	} else {
		for k, f := range funcs {
			go execRoutine(f, cr, k)
		}
	}

	for i := 0; i < ls; i++ {
		r := <-cr

		if r.err != nil {
			errs = append(errs, r.err)
		} else if lcr := len(r.results); lcr > 0 {
			res := make([]interface{}, lcr)
			for j, v := range r.results {
				res[j] = v.Interface()
			}
			results[r.key] = res
		}
	}

	return results, errs
}

// Executes the task and consumes the message of `sem` channel
func execRoutineParallel(f reflect.Value, c chan execResult, sem chan int, k string) {
	// execute routine
	execRoutine(f, c, k)

	// Once the task has done its job, consumes message from channel `sem`
	<-sem
}

// Executes the task and sends error to the `c` channel
func execRoutine(f reflect.Value, c chan execResult, key string) {
	var (
		exr = execResult{}      // Result
		res = f.Call(emptyArgs) // Calls the function
	)

	// Get type of the function to be executed
	fnt := f.Type()

	// Check if function returns any value
	if l := fnt.NumOut(); l > 0 {
		// Gets last return value type of the function
		lastArg := fnt.Out(l - 1)

		// Check if the last return value is error
		if reflect.Zero(lastArg).Interface() == emptyError {
			// If so and an error occured, set the execResult.error to the occurred error
			if e, ok := res[l-1].Interface().(error); ok {
				exr.err = e
			}
			// Decrements l so the results returned doesn't contain the error
			l = l - 1
		}

		// If no error occurred, fills the exr.results
		if exr.err == nil && l > 0 {
			exr.results = res[:l]
			// If result has a key
			if key != "" {
				exr.key = key
			}
		}
	}
	// Sends message to the error channel
	c <- exr
}
