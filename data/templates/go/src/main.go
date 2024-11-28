package main

import (
	"fmt"
	"time"
)

// Written by {{.app.author}}

func main() {

	fmt.Printf("Hello from %v", "{{.app.name}}")
	for {
		time.Sleep(1 * time.Second)
	}

}
