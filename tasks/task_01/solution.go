package main

import "strings"

func greet(name string) string {
	clearName := strings.TrimSpace(name) 
	if clearName == "" {
		clearName = "World"
	}
	return "Hello, " + clearName + "!"
}
