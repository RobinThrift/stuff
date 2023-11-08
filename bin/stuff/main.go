package main

import (
	"fmt"
	"os"

	"github.com/RobinThrift/stuff/app"
)

func main() {
	if err := app.Start(); err != nil {
		fmt.Println("error starting stuff", err)
		os.Exit(1)
	}
}
