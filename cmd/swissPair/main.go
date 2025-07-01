package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	. "github.comPhantomvv1/SwissPairAPI/internal/auth"
)

func main() {
	r := gin.Default()

	r.Any("/", func(c *gin.Context) { c.JSON(http.StatusOK, nil) })
	r.POST("/signup", SignUp)
	r.POST("/login", LogIn)
	r.POST("/profile", GetCurrentProfile)
	r.DELETE("/account", DeleteAccount)
}
