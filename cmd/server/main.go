package main

import "fmt"

func Run() error {
	fmt.Println("running server...")
	return nil
}

func main() {
	if err := Run(); err != nil {
		fmt.Println("error running server:", err)
	}
}
