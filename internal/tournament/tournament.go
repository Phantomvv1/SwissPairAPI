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
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

const (
	StatusPending = iota + 1
	StatusActive
	FinishedStatus
)

func (t *Tournament) GetTournament(conn *pgx.Conn) error {
	err := conn.QueryRow(context.Background(), "select name, owner_id, status, start from tournaments where id = $1", t.ID).Scan(
		&t.Name, &t.OwnerID, &t.Status, &t.Start)
	return err
}

func (t *Tournament) GetTournamentAdmin(conn *pgx.Conn) error {
	err := conn.QueryRow(context.Background(), "select name, owner_id, status, start, created_at, updated_at from tournaments where id = $1",
		t.ID).Scan(&t.Name, &t.OwnerID, &t.Status, &t.Start, &t.CreatedAt, &t.UpdatedAt)
	return err
}

func GetTournamentOwnerID(conn *pgx.Conn, tournamentID int) (int, error) {
	ownerID := 0
	err := conn.QueryRow(context.Background(), "select owner_id from tournaments where tournament_id = $1", tournamentID).Scan(&ownerID)
	if err != nil {
		return 0, err
	}

	return ownerID, err
}

func CreateTournamentsTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists tournaments "+
		"(id serial primary key, name text, owner_id int references authentication (id) "+
		"status int check(status in (1, 2, 3)), start timestamp, created_at timestamp, updated_at timestamp)")
	if err != nil {
		return err
	}

	return nil
}

func CreateTournament(c *gin.Context) { // test
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
		"values ($1, $2, $3, $4, current_timestamp, null)", name, id, StatusPending, startTS)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error couldn't put the information about the tournament in the database"})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func UpdateTournament(c *gin.Context) { // test
	var information map[string]interface{}
	json.NewDecoder(c.Request.Body).Decode(&information) // token && tournamentID && (name || start)

	token, ok := information["token"].(string)
	if !ok {
		log.Println("Inorrectly provided token")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error inorrectly provided token"})
		return
	}

	id, accountType, err := ValidateJWT(token)
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

	useName := false
	name, ok := information["name"].(string)
	if !ok {
		useName = true
	}

	start, ok := information["start"].(string)
	if !ok && !useName {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error no information provided"})
		return
	}

	startTS, err := time.Parse(time.RFC3339, start)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to parse the date and time"})
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"errro": "Error unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	ownerID, err := GetTournamentOwnerID(conn, tournamentID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error couldn't check the owner of the tournament"})
		return
	}

	if accountType != Admin && id != ownerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Error only admins and owners can edit the tournaments"})
		return
	}

	if useName && ok {
		_, err = conn.Exec(context.Background(), "update tournaments set name = $1, start = $2, updated_at = current_timestamp",
			name, startTS)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to update the tournament"})
			return
		}
	} else if useName {
		_, err = conn.Exec(context.Background(), "update tournaments set name = $1, updated_at = current_timestamp",
			name)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to update the tournament"})
			return
		}
	} else if ok {
		_, err = conn.Exec(context.Background(), "update tournaments set start = $1, updated_at = current_timestamp",
			name)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to update the tournament"})
			return
		}
	}
}

func GetTournament(c *gin.Context) { // test
	var information map[string]interface{}
	json.NewDecoder(c.Request.Body).Decode(&information) // token && tournamentID

	token, ok := information["token"].(string)
	if !ok {
		log.Println("Incorrectly provided token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error incorrectly provided token"})
		return
	}

	id, accountType, err := ValidateJWT(token)
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

	ownerID, err := GetTournamentOwnerID(conn, tournamentID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error couldn't check the owner of the tournament"})
		return
	}

	tournament := Tournament{}

	if accountType != Admin && id != ownerID {
		tournament.GetTournament(conn)
	} else {
		tournament.GetTournamentAdmin(conn)
	}

	c.JSON(http.StatusOK, gin.H{"tournament": tournament})
}

func DeleteTournament(c *gin.Context) {
	var information map[string]interface{}
	json.NewDecoder(c.Request.Body).Decode(&information) // token && tournamentID

	token, ok := information["token"].(string)
	if !ok {
		log.Println("Incorrectly provided token")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error incorrectly provided token"})
		return
	}

	id, accountType, err := ValidateJWT(token)
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

	ownerID, err := GetTournamentOwnerID(conn, tournamentID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to get the owner of the tournament from the database"})
		return
	}

	if accountType != Admin && id != ownerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Error only admins and owners can delete tournaments"})
	}

	check := 0
	err = conn.QueryRow(context.Background(), "delete from tournaments where id = $1 returning id", tournamentID).Scan(&check)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Error a tournament with this id doesn't exists"})
			return
		}

		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error couldn't delete the tournament from the database"})
		return
	}

	c.JSON(http.StatusOK, nil)
}
