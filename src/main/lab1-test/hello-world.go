package main

import (
	"fmt"
	"math/big"
	// "test/lab1-test2"
)

func main() {
	max := big.NewInt(1000)
	fmt.Println(max)
	say1()
	var a [3]int = [3]int{5, 2, 1}
	show(1, 2, max, a)
}
