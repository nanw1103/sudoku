package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func main() {

	if len(os.Args) > 1 {
		solveFile(os.Args[1], true)
		return
	}

	files, err := ioutil.ReadDir("./")
	if err != nil {
		panic(err)
	}

	fmt.Println("Name		Solution	Complexity	Time	Gaps	FastPath	Deduce")
	for _, f := range files {
		name := f.Name()
		if strings.HasSuffix(name, ".txt") {
			solveFile(name, false)
		}
	}
}

func solveFile(name string, verbose bool) {

	matrix := readFile(name)

	if !ValidateInput(matrix) {
		fmt.Printf("%10s	", name)
		fmt.Println(" <Invalid input>")
		return
	}

	if verbose {
		printMatrix(matrix)
		success, stat := Solve(matrix)
		printMatrix(matrix)
		printStat(stat)
		if !success {
			fmt.Println("No solution found")
		}
	} else {
		fmt.Printf("%10s	", name)
		success, stat := Solve(matrix)

		if success {
			fmt.Print("V		")
		} else {
			fmt.Print("X		")
		}
		fmt.Print(stat.backward, "		", stat.timeCost, "	", stat.gaps, "	", stat.fastPathFill, "		", stat.deducePrune)
		fmt.Println()
	}
}

func printMatrix(array []byte) {
	fmt.Println("-----------")
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if c == 3 || c == 6 {
				fmt.Print(" ")
			}
			fmt.Print(array[r*9+c])
		}
		fmt.Println()
		if r == 2 || r == 5 {
			fmt.Println()
		}
	}
}

func printStat(stat Statistics) {
	fmt.Println()
	fmt.Println("Complexity (#backward):", stat.backward)
	fmt.Println("	gaps:", stat.gaps)
	fmt.Println("	fastPathFill:", stat.fastPathFill)
	fmt.Println("	fastPathPrune:", stat.fastPathPrune, ",", stat.fastPathPruneBranches)
	fmt.Println("	deducePrune:", stat.deducePrune, ",", stat.deducePruneBranches)
	fmt.Println("	failedAttempt:", stat.failedAttempt)
	fmt.Println("Time ms:", stat.timeCost)
}

func readFile(file string) []byte {
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	text := string(dat)

	text = strings.Replace(text, " ", "", -1)
	text = strings.Replace(text, "\r", "", -1)
	lines := strings.Split(text, "\n")

	matrix := make([]byte, 81)
	d := 0
	for r := 0; r < 9; r++ {
		line := lines[r+d]
		for len(line) == 0 {
			d++
			line = lines[r+d]
		}
		for c := 0; c < 9; c++ {
			v := line[c]
			matrix[r*9+c] = v - '0'
		}
	}
	return matrix
}
