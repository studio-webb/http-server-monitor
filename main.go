package main

import (
	"fmt"
)

func main() {
	fmt.Println("Print system monitor")
	go func() {
		for {
			GetSystemSection()
		}
	}()
}
