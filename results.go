package async

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
