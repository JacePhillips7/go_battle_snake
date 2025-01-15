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
		for _, b := range o.Body {
			dangerSpots[b] = true
		}
		dangerSpots[o.Head] = true
		if o.Head != myHead {
			head_cords := genNextMoves(o.Head)
			for _, h := range head_cords {
				dangerSpots[h] = true
			}
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

	if len(safeMoves) == 1 {
		log.Printf("MOVE %s: Only 1 option to move\n", safeMoves[0])
		return BattlesnakeMoveResponse{Move: safeMoves[0]}
	}
	nextMove := safeMoves[0]
	moveValue := 0
	// this is where we make a maps of all squares and rank them by what they are worth
	rankMap := make(map[Coord]int)
	// we will rank all danger spots as a 0
	for key := range dangerSpots {
		rankMap[key] = -1
	}
	//give spots value by distance from head
	calcMapFromDistance(myHead, &rankMap)
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

		rank = rankMap[runningCord]
		if rank > moveValue {
			nextMove = move
			moveValue = rank
		}

	}

	log.Printf("MOVE %d: %s SCORE: %d\n", state.Turn, nextMove, moveValue)
	return BattlesnakeMoveResponse{Move: nextMove}
}

func calcMapFromDistance(cor Coord, m *map[Coord]int) {
	for c, v := range *m {
		if v == -1 { //point is dead so we don't evaluate it
			continue
		}
		dis := calcDistance(c, cor)
		value := 10 - dis
		if value < 0 {
			value = 0
		}
		v += value
	}
}
func calcDistance(c1 Coord, c2 Coord) int {
	dis := Abs((c1.X) - (c2.X))
	dis += Abs((c1.Y) - (c2.Y))
	return dis
}
func main() {
	RunServer()
}
