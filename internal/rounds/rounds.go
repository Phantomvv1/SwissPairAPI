package rounds

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	. "github.comPhantomvv1/SwissPairAPI/internal/auth"
	. "github.comPhantomvv1/SwissPairAPI/internal/swiss"
	. "github.comPhantomvv1/SwissPairAPI/internal/tournament"
)

type Round struct {
	ID           int `json:"id"`
	Player1ID    int `json:"player_1_id"`
	Player2ID    int `json:"player_2_id"`
	Result       int `json:"result"`
	TournamentID int `json:"tournament_id"`
}

const (
	ResultPlayer1Win = iota + 1
	ResultPlayer2Win
	ResultDraw
)

func CreateRoundsTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists rounds (id serial primary key, pl_1 int "+
		"references authentication(id), pl_2 int references authentication(id), result int check(result in (1, 2, 3)) "+
		"tournament_id int references tournaments(id))")

	return err
}

func GetRounds(conn *pgx.Conn, tournamentID int) ([]Round, []Player, error) {
	rows, err := conn.Query(context.Background(), "select id, pl_1, pl_2, result from rounds where tournament_id = $1", tournamentID)
	if err != nil {
		return nil, nil, err
	}

	rounds := make([]Round, 0)
	players := make([]Player, 0)
	for rows.Next() {
		id, pl1, pl2, result := 0, 0, 0, 0
		err = rows.Scan(&id, &pl1, &pl2, &result)
		if err != nil {
			return nil, nil, err
		}

		rounds = append(rounds, Round{ID: id, Player1ID: pl1, Player2ID: pl2, Result: result, TournamentID: tournamentID})

		index := GetIndexOfPlayer(players, pl1)
		if index == -1 {
			players = append(players, Player{Id: int64(pl1)})
		}
		players[index].Opponent[int64(pl2)] = struct{}{}

		index = GetIndexOfPlayer(players, pl2)
		if index == -1 {
			players = append(players, Player{Id: int64(pl2)})
		}
		players[index].Opponent[int64(pl1)] = struct{}{}

		switch result {
		case ResultPlayer1Win:
			players[GetIndexOfPlayer(players, pl1)].Score += 1.0
		case ResultPlayer2Win:
			players[index].Score += 1.0
		case ResultDraw:
			players[GetIndexOfPlayer(players, pl1)].Score += 0.5
			players[index].Score += 0.5
		}
	}

	return rounds, players, nil
}

func CreateRounds(c *gin.Context) {
	var information map[string]any // token && tournamentID
	json.NewDecoder(c.Request.Body).Decode(&information)

	token, ok := information["token"].(string)
	if !ok {
		log.Println("Incorrectly provided token")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error incorrectly provided token"})
		return
	}

	id, accType, err := ValidateJWT(token)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error invalid token"})
		return
	}

	tournamentIDFl, ok := information["tournamentID"].(float64)
	if !ok {
		log.Println("Incorrectly provided id of the tournament")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error incorrectly provided id of the tournament"})
		return
	}
	tournamentID := int(tournamentIDFl)

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	ownerId, err := GetTournamentOwnerID(conn, tournamentID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to get the id of the owner of the tournament"})
		return
	}

	if accType != Admin && id != ownerId {
		c.JSON(http.StatusForbidden, gin.H{"error": "Error only admins and the owner of the tournament can start the next rounds"})
		return
	}

	rows, err := conn.Query(context.Background(), "select user_id from players where tournament_id = $1", tournamentID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the ids of the players in the tournamet"})
		return
	}

	ids := make([]int, 0)
	for rows.Next() {
		id := 0
		err = rows.Scan(&id)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error working with the ids of the users"})
			return
		}

		ids = append(ids, id)
	}

	if rows.Err() != nil {
		log.Println(rows.Err())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error working with the ids of the users"})
		return
	}

	players := make([]Player, 0)
	for _, id := range ids {
		players = append(players, Player{Id: int64(id)})
	}

	rounds, emptyPlayer, ok := CreateSwissRound(players)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to create the new rounds"})
		return
	}

	log.Println(rounds)
	log.Println(emptyPlayer)
}

func GetAllRounds(c *gin.Context) {
	tournamentIDS := c.Param("tournamentID")
	if tournamentIDS == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error incorrectly provided id of the tournament"})
		return
	}

	tournamentID, err := strconv.Atoi(tournamentIDS)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to parse the id of the tournament"})
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	rounds, _, err := GetRounds(conn, tournamentID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to get the rounds"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rounds": rounds})
}
