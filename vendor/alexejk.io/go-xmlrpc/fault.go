package xmlrpc

import "fmt"

// Fault is a wrapper for XML-RPC fault object
type Fault struct {
	// Code provides numerical failure code
	Code int
	// String includes more detailed information about the fault, such as error name and cause
	String string
}

func (f *Fault) Error() string {
	return fmt.Sprintf("%d: %s", f.Code, f.String)
}
