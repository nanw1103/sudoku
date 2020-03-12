package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func main() {

	if len(os.Args) > 1 {
		success := solveFile(os.Args[1], true)
		if success {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
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

func solveFile(name string, verbose bool) bool {

	matrix, err := ReadMatrixFile(name)

	if err != nil {
		fmt.Printf("%-10s	", name)
		fmt.Println(" Error:", err.Error())
		return false
	}

	if verbose {
		printMatrix(matrix)
		success, stat := Solve(matrix)

		if success {
			printMatrix(matrix)
		} else {
			fmt.Println("\nNo solution found")
		}
		printStat(stat)
		return success
	}

	fmt.Printf("%-10s	", name)
	success, stat := Solve(matrix)

	if success {
		fmt.Print("V		")
	} else {
		fmt.Print("X		")
	}
	fmt.Print(stat.backward, "		", stat.timeCost, "	", stat.gaps, "	", stat.fastPathFill, "		", stat.deducePrune)
	fmt.Println()
	return success
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

// ReadMatrixFile reads the specified file as sudoku matrix
func ReadMatrixFile(file string) ([]byte, error) {
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		//panic(err)
		return nil, err
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

	if !ValidateInput(matrix) {
		return matrix, errors.New("Invalid Sudoku matrix")
	}

	return matrix, nil
}
