package main

import (
	"fmt"
)

func longestString(matrix [][]string) int {
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return 0
	}
	x := len(matrix[0][0])
	for _, v := range matrix {
		for _, z := range v {
			if len(z) > x {
				x = len(z)
			}
		}
	}
	return x
}

func printPaddedStr(s string, lm int) {
	left := lm - len(s)
	fmt.Printf(s)
	for i := 0; i < left; i++ {
		fmt.Printf(" ")
	}
}



func PrintGrid(matrix [][]string) {
	startSpaces := "  "
	lmax := longestString(matrix) + 4
	for i, row := range matrix {
		fmt.Printf(startSpaces)
		for j, _ := range row {
			printPaddedStr(matrix[i][j], lmax)
		}
		fmt.Printf("\n")
	}
}
