package main

import (
	"fmt"
	"os"
)

func say1() {
	fmt.Println("我是lab-test2")
}

func show(args ...interface{}) {
	for _, cont := range(args) {
		fmt.Printf("%T: ", cont)
		fmt.Println(cont)
	}
	// fmt.Println(args)
	os.Exit(0)
}
