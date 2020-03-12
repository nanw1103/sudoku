package main

import (
	"fmt"
	"time"
)

var matrix []byte

const enableDeducePrune = true
const enableFastPathPrune = true

// const enableTranspositionPrune = true

////////////////////////////////////////////////////////////////////////////
//	Entry
////////////////////////////////////////////////////////////////////////////

// Solve the Sudoku problem. The input array matrix is a
// 2D matrix in 1D representation: each cell is matrix[row*9+column].
// The result will be filled in-place.
// Return success or not and Statistics.
func Solve(m []byte) (bool, Statistics) {

	matrix = m
	stat = Statistics{}

	startTime := time.Now().UnixNano() / 1e6

	initSets()
	// transpositionPruneInit()
	openNodes.init(matrix)
	stat.gaps = openNodes.size

	//optional fast path solver
	if enableFastPathPrune {
		fastPathSolver()
	}

	//printMatrix(matrix)
	//openNodes._debugPrintOpenNodes()

	success := solveImpl(0)

	endTime := time.Now().UnixNano() / 1e6
	stat.timeCost = endTime - startTime

	if success && !ValidateSolved(matrix) {
		fmt.Println("Wrong algorithm")
		panic(11)
	}

	return success, stat
}

////////////////////////////////////////////////////////////////////////////
//	Statistics
////////////////////////////////////////////////////////////////////////////

// Statistics of the calculation
type Statistics struct {
	backward              int
	failedAttempt         int
	deducePrune           int
	deducePruneBranches   int
	fastPathFill          int
	fastPathPrune         int
	fastPathPruneBranches int
	// transpositionPrune    int

	timeCost int64
	gaps     int
}

var stat Statistics

////////////////////////////////////////////////////////////////////////////
//	Impl
////////////////////////////////////////////////////////////////////////////

func solveImpl(depth int) bool {

	if openNodes.size == 0 {
		return true
	}

	loc := openNodes.first()
	row := loc / 9
	col := loc % 9
	a := deducedAvailableSet(row, col)

	//fmt.Println("LOC ", row, ",", col, " pc=", PopCount(a))
	for _, v := range bitSet2Numbers(a) {

		//transposition prune
		// alreadyVisited := transpositionPruneAdd(depth, loc, v)

		// if enableTranspositionPrune && alreadyVisited {
		// 	stat.transpositionPrune++
		// 	transpositionPruneRemove(loc, v)
		// 	continue
		// }

		matrix[loc] = v     //try filling one
		onFill(row, col, v) //keep sets up-to-date
		openNodes.removeFirst()
		//fmt.Println("  Fill ", row, " ", col, " ", i)

		//optional optimization: update open nodes, in the same row/col/box
		valid := recalculateAffectedNodes(row, col)

		//solve more
		if valid && solveImpl(depth+1) {
			return true //find single solution and out
		}

		//fmt.Println("Unfill ", row, " ", col, " ", i)
		//backward. undo the fill
		stat.backward++
		matrix[loc] = 0
		onUnfill(row, col, v)
		openNodes.unremoveFirst()

		// transpositionPruneRemove(loc, v)

	}
	return false
}

func fastPathSolver() bool {

	for openNodes.size > 0 {
		loc := openNodes.first()
		row := loc / 9
		col := loc % 9
		a := deducedAvailableSet(row, col)
		v := singleBitSet2Value(a)
		if v != 255 { //if its a single possibility value
			matrix[loc] = v     //a sure fill
			onFill(row, col, v) //keep sets up-to-date
			openNodes.removeFirst()
			stat.fastPathFill++

			//mark all affected gaps as dirty
			fastPathOnChange(loc)

			continue
		}

		//no more confirmed ones. try deep calculation
		changed := false
		for {
			loc := openNodes._fastPathPoll()
			if loc == -1 {
				break
			}
			changed = openNodes.recalculate(loc)
			if changed {
				fastPathOnChange(loc)
				if openNodes.meta[loc].choices == 1 {
					break
				}
			}
		}

		if !changed {
			return false
		}

	}
	return true
}

func fastPathOnChange(loc int) {
	r := loc / 9
	c := loc % 9
	slots := getEmptySlotsInTheSameRowOrColOrBox(r, c)
	for _, loc := range slots {
		openNodes._fastPathAdd(loc)
	}
}

func recalculateAffectedNodes(row, col int) bool {
	slots := getEmptySlotsInTheSameRowOrColOrBox(row, col)
	for _, loc := range slots {
		if !recalculateOpenNode(loc) {
			return false
		}
	}
	return true
}

func recalculateOpenNode(loc int) bool {
	r := loc / 9
	c := loc % 9
	numChoices := popCount(deducedAvailableSet(r, c))

	if numChoices == 0 {
		stat.failedAttempt++
		return false
	}

	if numChoices == openNodes.meta[loc].choices { //if its not changed
		return true
	}

	openNodes.meta[loc].choices = numChoices
	openNodes.reorder(loc)
	return true
}

////////////////////////////////////////////////////////////////////////////
//	Sets to help available number calculation
////////////////////////////////////////////////////////////////////////////

var rowSet = make([]uint16, 9) //rowSet[rowId] is a bit set, each 1 at index N means number N+1 is not seen in the row
var colSet = make([]uint16, 9) //similar to rowSet
var boxSet = make([]uint16, 9) //similar to rowSet

func onFill(r, c int, n byte) {
	bit := ^(uint16(1) << uint(n-1))
	rowSet[r] &= bit
	colSet[c] &= bit
	boxSet[rowCol2Box(r, c)] &= bit
}

func onUnfill(r, c int, n byte) {
	bit := uint16(1) << uint(n-1)
	rowSet[r] |= bit
	colSet[c] |= bit
	boxSet[rowCol2Box(r, c)] |= bit
}

func availableSet(r, c int) uint16 {
	return rowSet[r] & colSet[c] & boxSet[rowCol2Box(r, c)] & openNodes.meta[r*9+c].available
}

func initSets() {
	//generate row/col/box sets
	for i := 0; i < 9; i++ {
		rowSet[i] = 0x01FF
		colSet[i] = 0x01FF
		boxSet[i] = 0x01FF
	}
	//make sure all the sets have correct state
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			v := matrix[r*9+c]
			if v == 0 {
				continue
			}
			onFill(r, c, v)
		}
	}
}

func deducedAvailableSet(row, col int) uint16 {

	mySet := availableSet(row, col)
	if mySet == 0 {
		return mySet
	}

	if !enableDeducePrune {
		return mySet
	}

	deducedRowSet := mySet
	deducedColSet := mySet
	deducedBoxSet := mySet

	//the rationale here:
	//for each perspective (e.g. a row), for position A, remove availabilities of all other gaps from availability set of A,
	//if the result is not empty, then it means only A is possible to place the values left.
	//e.g. if gap A, B and C are in the same row. There are no other gaps in the row. Availability sets are:
	//A [1, 2, 3]
	//B [2, 3]
	//C [2, 3]
	//Then only A is possible to be 1. So availability of A is reduced to [1], from this perspective (the row)

	locsInBox := boxID2Locations[rowCol2Box(row, col)]
	me := row*9 + col
	for j := 0; j < 9; j++ {
		loc := row*9 + j
		if matrix[loc] == 0 && j != col {
			t := availableSet(row, j)
			deducedRowSet &^= t
		}
		loc = j*9 + col
		if matrix[loc] == 0 && j != row {
			t := availableSet(j, col)
			deducedColSet &^= t
		}
		loc = locsInBox[j]
		rr := loc / 9
		cc := loc % 9
		if matrix[loc] == 0 && loc != me {
			t := availableSet(rr, cc)
			deducedBoxSet &^= t
		}
	}

	deducedSet := mySet
	if deducedRowSet != 0 {
		deducedSet &= deducedRowSet
	}
	if deducedColSet != 0 {
		deducedSet &= deducedColSet
	}
	if deducedBoxSet != 0 {
		deducedSet &= deducedBoxSet
	}

	if deducedSet != mySet {
		stat.deducePrune++
		stat.deducePruneBranches += popCount(mySet) - popCount(deducedSet)

		//fmt.Println("deduced from ", PopCount(mySet), " to ", PopCount(deducedSet))
	}

	return deducedSet
}

func singleBitSet2Value(a uint16) byte {
	switch a {
	case 1:
		return 1
	case 2:
		return 2
	case 4:
		return 3
	case 8:
		return 4
	case 16:
		return 5
	case 32:
		return 6
	case 64:
		return 7
	case 128:
		return 8
	case 256:
		return 9
	}
	return 255
}

func bitSet2Numbers(a uint16) []byte {
	size := 0
	ret := [9]byte{}

	for i := byte(0); i < 9; i++ {
		if a&(uint16(1)<<uint(i)) == 0 {
			continue
		}
		ret[size] = i + 1
		size++
	}

	return ret[:size]
}

////////////////////////////////////////////////////////////////////////////
//	Box related operations
////////////////////////////////////////////////////////////////////////////

var boxID2Locations = make([][]int, 9) //boxID2Locations[boxId] contains 9 locations (matrix coordination) in the box area

func init() {
	//boxId2Location
	boxSetSize := make([]int, 9)
	for i := 0; i < 9; i++ {
		boxID2Locations[i] = make([]int, 9)
	}
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			boxID := rowCol2Box(r, c)
			loc := r*9 + c
			size := boxSetSize[boxID]
			boxID2Locations[boxID][size] = loc
			boxSetSize[boxID] = size + 1
		}
	}
}

func rowCol2Box(r, c int) int {
	rr := r / 3
	cc := c / 3
	return rr*3 + cc
}

func getEmptySlotsInTheSameRowOrColOrBox(row, col int) []int {
	ret := make([]int, 9-1+9-1+4)
	size := 0

	locsInBox := boxID2Locations[rowCol2Box(row, col)]
	for j := 0; j < 9; j++ {
		loc := row*9 + j
		if matrix[loc] == 0 && j != col {
			ret[size] = loc
			size++
		}
		loc = j*9 + col
		if matrix[loc] == 0 && j != row {
			ret[size] = loc
			size++
		}
		loc = locsInBox[j]
		rr := loc / 9
		cc := loc % 9
		if matrix[loc] == 0 && rr != row && cc != col {
			ret[size] = loc
			size++
		}
	}
	return ret[:size]
}

////////////////////////////////////////////////////////////////////////////
//	sortedNodes: A structure to keep open locations in sorted manner
////////////////////////////////////////////////////////////////////////////
type metaType struct {
	choices   int
	idx       int    //index in sortedLocations
	available uint16 //calculated bit set of location loc, like rowSet/colSet/boxSet. [for fast path test only]
	dirty     bool   //whether availableSet is dirty.  [for fast path test only]
}
type sortedNodes struct {
	meta            []metaType //meta[loc]: meta data of location loc
	sortedLocations []int      //an sorted array, content range [start, start+size), consists of locations.

	size  int //size of open nodes
	start int //start index of open nodes in sortedLocations

	_dirtySet     []int //contains locations which meta.available should be recalculated
	_dirtySetSize int
}

var openNodes sortedNodes

func init() {
	openNodes.meta = make([]metaType, 81)
	openNodes.sortedLocations = make([]int, 81)
	openNodes._dirtySet = make([]int, 81)
}

func (me *sortedNodes) init(matrix []byte) {
	me.size = 0
	me.start = 0
	me._dirtySetSize = 0

	for loc := range matrix {
		me.meta[loc] = metaType{
			0,
			-1,
			0x01FF, //make sure all default available sets contains 9 numbers
			false,
		}
	}

	for loc, v := range matrix {
		if v != 0 {
			continue
		}
		r := loc / 9
		c := loc % 9
		a := deducedAvailableSet(r, c)
		me.meta[loc] = metaType{
			popCount(a),
			me.size,
			a,
			true,
		}
		me.sortedLocations[me.size] = loc
		me.size++
		me.reorder(loc)

		me._dirtySet[me._dirtySetSize] = loc
		me._dirtySetSize++
	}
}

func (me *sortedNodes) first() int {
	return me.sortedLocations[me.start]
}

func (me *sortedNodes) removeFirst() {
	me.size--
	me.start++
}

func (me *sortedNodes) unremoveFirst() {
	me.size++
	me.start--
}

func (me *sortedNodes) reorder(loc int) {

	m := &me.meta[loc]
	idx := m.idx

	//try moving left
	exchange := -1
	choices := m.choices
	for i := idx - 1; i >= me.start; i-- {
		if me.meta[me.sortedLocations[i]].choices <= choices {
			break
		}
		exchange = i
	}

	if exchange == -1 { //if can not move left
		//try moving right
		for i := idx + 1; i < me.start+me.size; i++ {
			if me.meta[me.sortedLocations[i]].choices >= choices {
				break
			}
			exchange = i
		}
	}

	if exchange != -1 { //if moved
		//exchange node in the priority queue
		otherLoc := me.sortedLocations[exchange]
		me.sortedLocations[m.idx] = otherLoc
		me.sortedLocations[exchange] = loc
		me.meta[loc].idx = exchange
		me.meta[otherLoc].idx = idx

		//rearrange the moved one to make sure its correct position
		me.reorder(otherLoc)
	}
}

func (me *sortedNodes) recalculate(loc int) bool {
	m := &me.meta[loc]

	m.dirty = false

	r := loc / 9
	c := loc % 9
	a := deducedAvailableSet(r, c)
	if a == m.available {
		return false
	}

	n := popCount(a)
	// if n >= openNodes.meta[loc].choices {
	// 	panic(111)
	// }
	//fmt.Println("reduced:", openNodes.meta[loc].choices, n)
	stat.fastPathPrune++
	stat.fastPathPruneBranches += m.choices - n

	m.choices = n
	m.available = a
	me.reorder(loc)
	return true
}

func (me *sortedNodes) _fastPathAdd(loc int) {
	if me.meta[loc].dirty {
		return
	}

	me.meta[loc].dirty = true
	me._dirtySet[me._dirtySetSize] = loc
	me._dirtySetSize++
}

func (me *sortedNodes) _fastPathPoll() int {
	if me._dirtySetSize == 0 {
		return -1
	}
	me._dirtySetSize--
	return me._dirtySet[me._dirtySetSize]
}

func (me *sortedNodes) _debugPrintOpenNodes() {
	for _, loc := range me.sortedLocations[me.start : me.start+me.size] {
		r := loc / 9
		c := loc % 9
		numbers := bitSet2Numbers(deducedAvailableSet(r, c))
		fmt.Printf("%d, (%d, %d): %v\n", loc, r, c, numbers)
	}
}

////////////////////////////////////////////////////////////////////////////
//	Validation
////////////////////////////////////////////////////////////////////////////

// ValidateInput validates the input of a Sudoku matrix
func ValidateInput(matrix []byte) bool {
	if len(matrix) != 81 {
		return false
	}
	for loc := 0; loc < 81; loc++ {
		if matrix[loc] > 9 {
			return false
		}

		if !validateOne(matrix, loc, false) {
			return false
		}
	}
	return true
}

// ValidateSolved validates the result of a solved Sudoku matrix
func ValidateSolved(array []byte) bool {
	for loc := 0; loc < 81; loc++ {
		if !validateOne(array, loc, true) {
			return false
		}
	}
	return true
}

func validateOne(array []byte, loc int, allowZero bool) bool {
	v := array[loc]
	if allowZero {
		if v == 0 {
			return false
		}
	} else {
		if v == 0 {
			return true
		}
	}

	r := loc / 9
	c := loc % 9
	for i := 0; i < 9; i++ {
		if i != c && array[r*9+i] == v {
			return false
		}
		if i != r && array[i*9+c] == v {
			return false
		}

		boxID := rowCol2Box(r, c)
		locations := boxID2Locations[boxID]
		for j := range locations {
			loc := locations[j]
			if r*9+c == loc {
				continue
			}

			if array[loc] == v {
				return false
			}
		}
	}
	return true
}

////////////////////////////////////////////////////////////////////////////
//	Zorbrist hashing & transposition table
////////////////////////////////////////////////////////////////////////////
// var zorbristNumbers = make([]uint64, 81*9)

// func init() {
// 	for i := range zorbristNumbers {
// 		zorbristNumbers[i] = rand.Uint64()
// 	}

// 	for i := 0; i < 81; i++ {
// 		transpositionDepthSets[i] = setType{}
// 	}
// }

// var transpositionCurrent uint64

// type setType map[uint64]bool

// var transpositionDepthSets = make([]setType, 81)

// func transpositionPruneInit() {
// 	transpositionCurrent = 0
// 	// for i := 0; i < 81; i++ {
// 	// 	transpositionDepthSets[i] = setType{}
// 	// }
// }

// func _transpositionPruneMask(loc int, v byte) {
// 	transpositionCurrent ^= zorbristNumbers[loc*9+int(v)-1]
// }

// func transpositionPruneAdd(depth, loc int, v byte) bool {
// 	_transpositionPruneMask(loc, v)

// 	if transpositionDepthSets[depth][transpositionCurrent] {
// 		return true
// 	}
// 	transpositionDepthSets[depth][transpositionCurrent] = true
// 	return false
// }

// func transpositionPruneRemove(loc int, v byte) {
// 	_transpositionPruneMask(loc, v)
// }

// func transpositionPruneDump() {
// 	for depth, s := range transpositionDepthSets {
// 		if len(s) == 0 {
// 			continue
// 		}
// 		fmt.Println(depth, len(s))
// 	}
// }

////////////////////////////////////////////////////////////////////////////
//	PopCount/Hamming weight
////////////////////////////////////////////////////////////////////////////
// pc[i] is the population count of i.
var pc [256]byte

func init() {
	popCountInit()
}

func popCountInit() {
	for i := range pc {
		pc[i] = pc[i>>1] + byte(i&1)
	}
}

// PopCount returns the population count (number of set bits) of x.
func popCount(x uint16) int {
	return int(pc[byte(x>>(0*8))] +
		pc[byte(x>>(1*8))])
	// pc[byte(x>>(2*8))] +
	// pc[byte(x>>(3*8))] +
	// pc[byte(x>>(4*8))] +
	// pc[byte(x>>(5*8))] +
	// pc[byte(x>>(6*8))] +
	// pc[byte(x>>(7*8))])
}
