package main

import (
	"log"
	"math"
	"sort"
)

func info() BattlesnakeInfoResponse {
	log.Println("INFO")

	return BattlesnakeInfoResponse{
		APIVersion: "1",
		Author:     "TheOtherJace",
		Color:      "#0178D6",
		Head:       "earmuffs",
		Tail:       "pixel",
	}
}

// start is called when your Battlesnake begins a game
func start(state GameState) {
	log.Println("GAME START")
}

// end is called when your Battlesnake finishes a game
func end(state GameState) {
	log.Printf("GAME OVER\n\n")
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func genNextMoves(c Coord) []Coord {
	moves := []Coord{
		{X: c.X + 1, Y: c.Y},
		{X: c.X - 1, Y: c.Y},
		{X: c.X, Y: c.Y + 1},
		{X: c.X, Y: c.Y - 1},
	}
	return moves
}

// move is called on every turn and returns your next move
// Valid moves are "up", "down", "left", or "right"
// See https://docs.battlesnake.com/api/example-move for available data
func move(state GameState) BattlesnakeMoveResponse {

	isMoveSafe := map[string]bool{
		"up":    true,
		"down":  true,
		"left":  true,
		"right": true,
	}

	// We've included code to prevent your Battlesnake from moving backwards
	myHead := state.You.Body[0] // Coordinates of your head
	myNeck := state.You.Body[1] // Coordinates of your "neck"

	if myNeck.X < myHead.X { // Neck is left of head, don't move left
		isMoveSafe["left"] = false

	} else if myNeck.X > myHead.X { // Neck is right of head, don't move right
		isMoveSafe["right"] = false

	} else if myNeck.Y < myHead.Y { // Neck is below head, don't move down
		isMoveSafe["down"] = false

	} else if myNeck.Y > myHead.Y { // Neck is above head, don't move up
		isMoveSafe["up"] = false
	}

	boardWidth := state.Board.Width
	boardHeight := state.Board.Height

	if myHead.X == boardWidth-1 {
		isMoveSafe["right"] = false
	}
	if myHead.X == 0 {
		isMoveSafe["left"] = false
	}

	if myHead.Y == boardHeight-1 {
		isMoveSafe["up"] = false
	}
	if myHead.Y == 0 {
		isMoveSafe["down"] = false
	}

	//prevent collision with self
	dangerSpots := make(map[Coord]bool, 0)
	//add all enemy snakes
	opponents := state.Board.Snakes
	for _, o := range opponents { /*  */
		for _, b := range o.Body[0 : len(o.Body)-1] { // this should mark tails as a safe spot
			dangerSpots[b] = true
		}
		dangerSpots[o.Head] = true
	}
	upCord := Coord{X: myHead.X, Y: myHead.Y + 1}
	downCord := Coord{X: myHead.X, Y: myHead.Y - 1}
	leftCord := Coord{X: myHead.X - 1, Y: myHead.Y}
	rightCord := Coord{X: myHead.X + 1, Y: myHead.Y}
	if dangerSpots[upCord] {
		isMoveSafe["up"] = false
	}
	if dangerSpots[downCord] {
		isMoveSafe["down"] = false
	}
	if dangerSpots[leftCord] {
		isMoveSafe["left"] = false
	}
	if dangerSpots[rightCord] {
		isMoveSafe["right"] = false
	}

	rewardMap := make(map[Coord]bool, 0)

	//add food to rewards
	for _, f := range state.Board.Food {
		rewardMap[f] = true
	}

	// Are there any safe moves left?
	safeMoves := []string{}
	for move, isSafe := range isMoveSafe {
		if isSafe {
			safeMoves = append(safeMoves, move)
		}
	}

	if len(safeMoves) == 0 {
		log.Printf("MOVE %d: No safe moves detected! Moving down\n", state.Turn)
		return BattlesnakeMoveResponse{Move: "down"}
	}

	if len(safeMoves) == 1 {
		log.Printf("MOVE %s: Only 1 option to move\n", safeMoves[0])
		return BattlesnakeMoveResponse{Move: safeMoves[0]}
	}
	moveValue := 0
	moveSafeLevel := 0
	// this is where we make a maps of all squares and rank them by what they are worth
	rankMap := make(map[Coord]int)
	for y := 0; y < boardHeight; y++ {
		for x := 0; x < boardWidth; x++ {
			boardCoord := Coord{
				X: x, Y: y,
			}
			rankMap[boardCoord] = 0
		}

	}
	// we will rank all danger spots as a 0
	for key := range dangerSpots {
		rankMap[key] = -1
	}
	//give spots value by distance from head
	calcMapFromDistance(myHead, rankMap, 2)
	nearFood := nearestFood(myHead, state.Board)
	if nearFood.X != -1 && nearFood.Y != -1 {
		calcMapFromDistance(nearFood, rankMap, 1)
	}
	for _, f := range state.Board.Food {
		calcMapFromDistance(f, rankMap, 1)
	}
	possibleOppsMoves := make(map[Coord]bool)
	for _, o := range state.Board.Snakes {
		if o.Head != myHead {
			calcMapFromDistance(o.Head, rankMap, -1)

			//gives extra rank if we can kill another snake, inverse is true
			head_cords := genNextMoves(o.Head)
			for _, h := range head_cords {
				if len(state.You.Body) > o.Length {
					rankMap[h] += 100
				} else {
					possibleOppsMoves[h] = true
					rankMap[h] -= 100
				}
			}
		}
	}
	WMoves := []WeightedMove{}
	for _, move := range safeMoves {

		var runningCord Coord
		switch move {
		case "down":
			runningCord = downCord
		case "up":
			runningCord = upCord
		case "left":
			runningCord = leftCord
		case "right":
			runningCord = rightCord
		}
		//deep copy danger to track past moves
		pastMoves := make(map[Coord]bool)
		for key, v := range dangerSpots {
			pastMoves[key] = v
		}
		safeLevel := numberOfSafeMoves(runningCord, pastMoves, state.Board, boardHeight*boardWidth)
		if possibleOppsMoves[runningCord] {
			safeLevel -= 100
		}
		rank := rankMap[runningCord]
		calculated := WeightedMove{
			Coord: runningCord,
			Move:  move,
			Safe:  safeLevel,
			Rank:  rank,
		}
		WMoves = append(WMoves, calculated)
	}
	bestMove := chooseMove(WMoves, state.You.Length-1)
	nextMove := bestMove.Move
	moveValue = bestMove.Rank
	moveSafeLevel = bestMove.Safe
	log.Printf("MOVE %d: %s SCORE: %d SAFE: %d\n", state.Turn, nextMove, moveValue, moveSafeLevel)
	return BattlesnakeMoveResponse{Move: nextMove}
}
func chooseMove(moves []WeightedMove, safeWanted int) WeightedMove {
	safeMoves := []WeightedMove{}
	for _, v := range moves {
		if v.Safe >= safeWanted {
			safeMoves = append(safeMoves, v)
		}
	}
	if len(safeMoves) == 0 { //out of safe moves from wanted, just choose the safest
		sort.Slice(moves, func(i, j int) bool {
			return moves[i].Safe > moves[j].Safe
		})
		return moves[0]
	}
	//sort by safe
	sort.Slice(safeMoves, func(i, j int) bool {
		return safeMoves[i].Safe > safeMoves[j].Safe
	})
	//now we can choose from rank
	sort.Slice(safeMoves, func(i, j int) bool {
		return safeMoves[i].Rank > safeMoves[j].Rank
	})
	return safeMoves[0]
}
func numberOfSafeMoves(move Coord, pastMoves map[Coord]bool, board Board, layers int) int {
	safe := 1
	if pastMoves[move] {
		return 0
	}
	if layers == 0 {
		return safe
	}
	pastMoves[move] = true
	possibleMoves := genNextMoves(move)
	for _, pm := range possibleMoves {
		if pm.X < 0 || pm.Y < 0 || pm.X >= board.Width || pm.Y >= board.Height {
			continue
		}
		if pastMoves[pm] {
			continue
		}
		safe += numberOfSafeMoves(pm, pastMoves, board, layers-1)
	}
	return safe
}
func calcMapFromDistance(cor Coord, m map[Coord]int, weight int) {
	for c, v := range m {
		if v == -1 { //point is dead so we don't evaluate it
			continue
		}
		dis := calcDistance(c, cor)
		value := 10 - dis
		if value < 0 {
			value = 0
		}
		m[c] = v + value*weight
	}
}
func nearestFood(head Coord, board Board) Coord {
	food := board.Food
	if len(food) == 0 {
		return Coord{X: -1, Y: -1}
	}
	dis := board.Width + board.Height
	closeFood := food[0]
	for _, f := range food {
		d := calcDistance(head, f)
		if d < dis {
			dis = d
			closeFood = f
		}
	}
	return closeFood
}
func calcDistance(c1 Coord, c2 Coord) int {
	dis := math.Sqrt(float64(square(c1.X-c2.X) + square(c1.Y-c2.Y)))
	return int(dis)
}
func square(i int) int {
	return i * i
}
func main() {
	RunServer()
}
