package swiss

type Player struct {
	Id       int64
	Score    int64
	Opponent map[int64]struct{} // Opponents encountered before
}

// pickTablePlayer calculates the arrangement of players in a Swiss Tournament
func pickTablePlayer(players []int64, playerOpponentMap map[int64]map[int64]struct{}) ([]int64, bool) {
	if len(players) < 2 {
		return players, true
	}
	whitePlayer := players[0]
	opponentMap, _ := playerOpponentMap[whitePlayer]
	for i := range players {
		if i != 0 {
			// Check if already played against
			if _, has := opponentMap[players[i]]; !has {
				// Select
				res := make([]int64, 2)
				res[0] = whitePlayer
				res[1] = players[i]

				// Assemble the remaining sorted data
				var nextRound []int64
				nextRound = append(nextRound, players[1:i]...)
				nextRound = append(nextRound, players[i+1:]...)
				pick, ok := pickTablePlayer(nextRound, playerOpponentMap) // Proceed to the next round of sorting
				if ok {
					return append(res, pick...), true // Success, result floats up
				}
			}
		}
	}
	return nil, false // Failure, recalculate in the upper level
}

func CreateSwissRound(players []Player) (playerBattleList [][]int64, emptyPlayer int64, ok bool) {
	ok = true

	// Determine the bye player
	total := len(players)
	if total%2 != 0 {
		emptyPlayer = players[total-1].Id
		players = players[:total]
	}

	// Convert data structure
	var playerIds []int64
	var playerOpponentMap = make(map[int64]map[int64]struct{})
	for _, v := range players {
		playerIds = append(playerIds, v.Id)
		if _, has := playerOpponentMap[v.Id]; !has {
			playerOpponentMap[v.Id] = v.Opponent
		}
	}

	// Calculate the match order
	playerList, ok := pickTablePlayer(playerIds, playerOpponentMap)
	if !ok {
		return playerBattleList, emptyPlayer, ok
	}

	// Convert to a two-dimensional array
	for i := 0; i < len(playerList)/2; i++ {
		playerBattleList = append(playerBattleList, []int64{
			playerList[i*2],
			playerList[i*2+1],
		})
	}
	return
}
