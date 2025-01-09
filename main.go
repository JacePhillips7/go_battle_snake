package main

// Welcome to
// __________         __    __  .__                               __
// \______   \_____ _/  |__/  |_|  |   ____   ______ ____ _____  |  | __ ____
//  |    |  _/\__  \\   __\   __\  | _/ __ \ /  ___//    \\__  \ |  |/ // __ \
//  |    |   \ / __ \|  |  |  | |  |_\  ___/ \___ \|   |  \/ __ \|    <\  ___/
//  |________/(______/__|  |__| |____/\_____>______>___|__(______/__|__\\_____>
//
// This file can be a nice home for your Battlesnake logic and helper functions.
//
// To get you started we've included code to prevent your Battlesnake from moving backwards.
// For more info see docs.battlesnake.com

import (
	"log"
)

// info is called when you create your Battlesnake on play.battlesnake.com
// and controls your Battlesnake's appearance
// TIP: If you open your Battlesnake URL in a browser you should see this data
func info() BattlesnakeInfoResponse {
	log.Println("INFO")

	return BattlesnakeInfoResponse{
		APIVersion: "1",
		Author:     "TheOtherJace", // TODO: Your Battlesnake username
		Color:      "#0178D6",      // TODO: Choose color
		Head:       "earmuffs",     // TODO: Choose head
		Tail:       "pixel",        // TODO: Choose tail
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

func rankMove(move Coord, layers int, danger map[Coord]bool, reward map[Coord]bool, height int, width int, head Coord) int {
	if danger[move] {
		return 0
	}
	if move.X >= width || move.X < 0 {
		return 0
	}
	if move.Y >= height || move.Y < 0 {
		return 0
	}
	runningValue := 1

	if reward[move] {
		distance := int(Abs(int64(move.X)-int64(head.X)) + Abs(int64(move.Y)-int64(head.Y)))
		runningValue += 100

		if distance > runningValue {
			runningValue = 1
		}
	}

	if layers == 0 {
		return runningValue
	}
	layers--
	up := Coord{X: move.X, Y: move.Y + 1}
	down := Coord{X: move.X, Y: move.Y - 1}
	left := Coord{X: move.X - 1, Y: move.Y}
	right := Coord{X: move.X + 1, Y: move.Y}

	danger[move] = true

	return runningValue + rankMove(up, layers, danger, reward, height, width, head) +
		rankMove(down, layers, danger, reward, height, width, head) +
		rankMove(left, layers, danger, reward, height, width, head) +
		rankMove(right, layers, danger, reward, height, width, head)
}
func Abs(x int64) int64 {
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
	//add myself as a danger to society
	mybody := state.You.Body
	for _, v := range mybody {
		dangerSpots[v] = true
	}
	//add all enemy snakes as a danger
	opponents := state.Board.Snakes
	for _, o := range opponents {
		for _, b := range o.Body {
			dangerSpots[b] = true
		}
		//add all heads potential next move
		for _, h := range genNextMoves(o.Head) {
			dangerSpots[h] = true
		}
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

	nextMove := safeMoves[0]
	moveValue := 0
	layers := 10
	for _, move := range safeMoves {
		rank := 0
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
		rank = rankMove(runningCord, layers, dangerSpots, rewardMap, boardHeight, boardWidth, myHead)
		log.Printf("%s Scored: %d\n", move, rank)
		if rank > moveValue {
			nextMove = move
			moveValue = rank
		}

	}

	log.Printf("MOVE %d: %s SCORE: %d\n", state.Turn, nextMove, moveValue)
	return BattlesnakeMoveResponse{Move: nextMove}
}

func main() {
	RunServer()
}
