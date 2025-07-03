package tournament

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	. "github.comPhantomvv1/SwissPairAPI/internal/auth"
)

type Tournament struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	OwnerID   int       `json:"owner_id"`
	Status    int       `json:"status"`
	Start     time.Time `json:"start"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const (
	PendingStatus = iota + 1
	ActiveStatus
	FinishedStatus
)

func CreateTournamentsTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists tournaments "+
		"(id serial primary key, name text, owner_id int references authentication (id) "+
		"status int check(status in (1, 2, 3)), start timestamp, created_at timestamp, updated_at timestamp)")
	if err != nil {
		return err
	}

	return nil
}

func CreateTournament(c *gin.Context) {
	var information map[string]string
	json.NewDecoder(c.Request.Body).Decode(&information) // name && start && token

	token, ok := information["token"]
	if !ok {
		log.Println("Incorrectly provided token")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error incorrectly provided token"})
		return
	}

	id, _, err := ValidateJWT(token)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error invalid token"})
		return
	}

	name, ok := information["name"]
	if !ok {
		log.Println("Incorrectly provided name of the tournament")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error incorrectly provided name of the tournament"})
		return
	}

	start, ok := information["start"]
	if !ok {
		log.Println("Incorrectly provided start date and time of the tournament")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error incorrectly provided start date and time of the tournament"})
		return
	}

	startTS, err := time.Parse(time.RFC3339, start)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to parse the date correctly"})
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	if err = CreateTournamentsTable(conn); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to create the table for tournaments"})
		return
	}

	_, err = conn.Exec(context.Background(), "insert into tournaments (name, owner_id, status, start, created_at, updated_at) "+
		"values ($1, $2, $3, $4, current_timestamp, null)", name, id, PendingStatus, startTS)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error couldn't put the information about the tournament in the database"})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func UpdateTournament(c *gin.Context) {
	var information map[string]string
	json.NewDecoder(c.Request.Body).Decode(&information) // tournamentID && (name || start)

	useName := false
	name, ok := information["name"]
	if !ok {
		useName = true
	}

	start, ok := information["start"]
	if !ok && !useName {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error no information provided"})
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"errro": "Error unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	if useName {
		_
	}
}
