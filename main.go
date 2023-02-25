package main

import (
	"fmt"
	"github.com/mgeri/udptunneler/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil && err.Error() != "" {
		fmt.Println(err)
	}
}
