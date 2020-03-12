package main

import "testing"

func checkSolution(matrix []byte) bool {
	success, stat := Solve(matrix)
	_ = stat
	return success
}

func TestMatrixSimple(t *testing.T) {
	matrix, err := ReadMatrixFile("s01.txt")
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	success := checkSolution(matrix)
	if !success {
		t.Errorf("Failed finding solution")
	}
}

/*
func TestMatrixInvalid(t *testing.T) {
	matrix, err := ReadMatrixFile("s10-invalid.txt")
	if err == nil {
		t.Errorf("Fail identifying invalid input")
		return
	}
	_ = matrix
}

func TestMatrixNoSolution(t *testing.T) {
	matrix, err := ReadMatrixFile("s13-no-sol.txt")
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	success := checkSolution(matrix)
	if success {
		t.Errorf("Expect to find no solution, but found solution.")
	}
}

func TestMatrixEmpty(t *testing.T) {
	matrix, err := ReadMatrixFile("s12-empty.txt")
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	success := checkSolution(matrix)
	if !success {
		t.Errorf("Expect to find solution, but Failed.")
	}
}

func TestMatrixFull(t *testing.T) {
	matrix, err := ReadMatrixFile("s11-full.txt")
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	success := checkSolution(matrix)
	if !success {
		t.Errorf("Failed finding solution")
	}
}

func TestMatrixNormal(t *testing.T) {
	matrix, err := ReadMatrixFile("s20.txt")
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	success := checkSolution(matrix)
	if !success {
		t.Errorf("Failed finding solution")
	}
}
*/
