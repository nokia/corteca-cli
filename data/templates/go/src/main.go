package main

import (
	"fmt"
	"time"
{{if .app.options.use_libhlapi_module}}   "hlapi" {{end}}
)

// Written by {{.app.author}}

func main() {
	fmt.Printf("Hello from %v", "{{.app.name}}")
	for {
	    time.Sleep(1*time.Second)
	}
	
}
