package td

// Thing property value definition
//   credits: https://github.com/dravenk/webthing-go/blob/master/value.go

// Value A property value.
//
// This is used for communicating between the Thing representation and the
// actual physical thing implementation.
//
// Notifies all observers when the underlying value changes through an external
// update (command to turn the light off) or if the underlying sensor reports a
// new value.
type Value struct {
	lastValue      interface{}
	valueForwarder []func(interface{})
}
