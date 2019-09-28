// +build darwin linux
// +build amd64
package tid

// getgid restrieves goroutine id from TSL.
func gettid() int64

// GoID gets current goroutine ID.
func ThreadID() int64 {
	return gettid()
}
