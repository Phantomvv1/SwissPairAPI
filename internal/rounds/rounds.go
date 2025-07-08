package rounds

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	. "github.comPhantomvv1/SwissPairAPI/internal/auth"
	. "github.comPhantomvv1/SwissPairAPI/internal/swiss"
	. "github.comPhantomvv1/SwissPairAPI/internal/tournament"
)

const (
	ResultPlayer1Win = iota + 1
	ResultPlayer2Win
	ResultDraw
)

func CreateRoundsTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists rounds (id serial primary key, pl_1 int "+
		"references authentication(id), pl_2 references authentication(id), result int check(result in (1, 2, 3)")

	return err
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
