package mr

import (
	"fmt"
	"os"
)

func wait(a int) int{
	if a == 0 {
		os.Exit(3)
	}
	a--
	return a
}

func show(args ...interface{}) {
	showAndContinue(args...)
	os.Exit(2)
}

func showAndContinue(args ...interface{}) {
	for idx, cont := range args {
		if idx == 0 {
			fmt.Println("**********************************")
			fmt.Println("            ", cont)
			fmt.Println("----------------------------------")
		} else {
			// fmt.Printf("%T: ", cont)
			fmt.Println(cont)
		}
	}
	fmt.Println("----------------------------------")
	fmt.Println("              End")
	fmt.Println("**********************************")
}
