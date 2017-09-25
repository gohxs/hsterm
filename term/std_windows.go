// +build windows

package term

func init() {
	Stdin = NewRawReader()
	Stdout = New(Stdout)
}
