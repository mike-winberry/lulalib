package message

import "fmt"

// Generic is used to implement the io.Writer interface for generic messages.
type Generic struct{}

func (g *Generic) Write(p []byte) (n int, err error) {
	text := string(p)
	fmt.Println(text)
	return len(p), nil
}
