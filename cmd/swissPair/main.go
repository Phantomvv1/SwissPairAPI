package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	. "github.comPhantomvv1/SwissPairAPI/internal/auth"
	. "github.comPhantomvv1/SwissPairAPI/internal/players"
	. "github.comPhantomvv1/SwissPairAPI/internal/tournament"
)

func main() {
	r := gin.Default()

	r.Any("/", func(c *gin.Context) { c.JSON(http.StatusOK, nil) })
	r.POST("/signup", SignUp)
	r.POST("/login", LogIn)
	r.POST("/profile", GetCurrentProfile)
	r.DELETE("/account", DeleteAccount)

	t := r.Group("/tournament")
	t.GET("/", GetAllTournaments)
	t.POST("/", CreateTournament)
	t.POST("/get", GetTournament)
	t.PUT("/", UpdateTournament)
	t.DELETE("/", DeleteTournament)
	t.POST("/status", GetTournamentsWithStatus)

	p := r.Group("/player")
	p.POST("/", CreatePlayer)
	p.POST("/get", GetPlayersForTournament)
	p.DELETE("/", RemoveUserFromTournament)

	r.Run(":42069")
}
