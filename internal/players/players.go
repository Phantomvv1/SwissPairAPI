package players

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	. "github.comPhantomvv1/SwissPairAPI/internal/auth"
	. "github.comPhantomvv1/SwissPairAPI/internal/emails"
	. "github.comPhantomvv1/SwissPairAPI/internal/tournament"
)

func GetPlayersForTournamentFromDB(conn *pgx.Conn, tournamentID int) ([]Profile, error) {
	rows, err := conn.Query(context.Background(), "select user_id from players where tournament_id = $1 order by user_id", tournamentID)
	if err != nil {
		return nil, err
	}

	var ids []interface{}
	for rows.Next() {
		id := 0
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	query := "select name, email from authentication where id in ("
	for i := range ids {
		if i == len(ids)-1 {
			query += "$" + fmt.Sprintf("%d", i+1)
		} else {
			query += "$" + fmt.Sprintf("%d", i+1) + ", "
		}
	}
	query += ")"

	rows, err = conn.Query(context.Background(), query, ids...)
	if err != nil {
		return nil, err
	}

	var users []Profile
	i := 0
	for rows.Next() {
		p := Profile{}
		err = rows.Scan(&p.Name, &p.Email)
		if err != nil {
			return nil, err
		}

		p.ID = ids[i].(int)
		i++
		users = append(users, p)
	}

	return users, nil
}

func CreatePlayersTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists players (tournament_id int, user_id int)")
	return err
}

func CreatePlayer(c *gin.Context) {
	var information map[string]any
	json.NewDecoder(c.Request.Body).Decode(&information) // token && playerID && tournamentID

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

	userIDFl, ok := information["playerID"].(float64)
	if !ok {
		log.Println("Incorrectly provided id of the user")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error Incorrectly provided id of the user"})
		return
	}
	userID := int(userIDFl)

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	if err = CreatePlayersTable(conn); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error couldn't create the table for the players"})
		return
	}

	ownerID, err := GetTournamentOwnerID(conn, tournamentID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error couldn't get the owner of the tournament from the database"})
		return
	}

	if accountType != Admin && id != ownerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Error only admins and owners can add players"})
		return
	}

	_, err = conn.Exec(context.Background(), "insert into players (tournament_id, user_id) values ($1, $2)", tournamentID, userID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to register the user as a player for your tournament"})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func GetPlayersForTournament(c *gin.Context) {
	tournamentIDS := c.Param("tournamentID")
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

	users, err := GetPlayersForTournamentFromDB(conn, tournamentID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to get the users who play in this tournament"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func RemoveUserFromTournament(c *gin.Context) {
	var information map[string]interface{}
	json.NewDecoder(c.Request.Body).Decode(&information) // token && tournamentID && userID

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

	userIDFl, ok := information["userID"].(float64)
	if !ok {
		log.Println("Incorrectly provided id of the user")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error incorrectly provided id of the user"})
		return
	}
	userID := int(userIDFl)

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error couldn't get the owner from the database"})
		return
	}

	if accountType != Admin && id != ownerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Error only admins and the owner of the tournament can remove users from the tournament"})
		return
	}

	check := 0
	err = conn.QueryRow(context.Background(), "delete from players where tournament_id = $1 and user_id = $2",
		tournamentID, userID).Scan(&check)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Error there is no user with this id playing in this tournament"})
			return
		}
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to remove the user from this tournament"})
		return
	}

	userEmail := ""
	err = conn.QueryRow(context.Background(), "select email from authentication a where a.id = $1", userID).Scan(&userEmail)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to get the email of the user you removed from the database"})
		return
	}

	tournamentName := ""
	err = conn.QueryRow(context.Background(), "select name from tournaments t where t.id = $1", tournamentID).Scan(&tournamentName)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to get the name of the tournament from the database"})
		return
	}

	err = RemoveEmail(userEmail, tournamentName)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending an email to the user"})
		return
	}

	c.JSON(http.StatusOK, nil)
}
