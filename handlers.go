package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func pongHandler(c *gin.Context) {
	c.Set("rendered", true)
	c.String(http.StatusOK, "pong")
}

func blankHandler(c *gin.Context) {
	c.Set("controller", `blank`)
	c.Set("action", `index`)
}

func simplePugHandler(c *gin.Context) {
	db, _ := NewGormDB(c)
	defer db.Close()

	var users []User
	db.Debug().Find(&users)

	variables := map[string]interface{}{}
	variables[`users`] = users
	c.Set(`variables`, variables)
}
