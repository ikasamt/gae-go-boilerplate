package app

import (
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

var (
	salt      = "T9rrLADK"
	mysession = "mysession"
)

func pongHandler(c *gin.Context) {
	c.Set("rendered", true)
	c.String(http.StatusOK, "pong")
}

func blankHandler(c *gin.Context) {
	c.Set("controller", `blank`)
	c.Set("action", `index`)
}

func simplePugHandler(c *gin.Context) {}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	sessionStore := cookie.NewStore([]byte(salt))

	root := gin.Default()

	root.Use(loggingMiddleware())
	root.Use(paramsMiddleware(``))
	root.Use(pugMiddleware(`layout.jade`))
	root.Use(errorMiddleware(`layout_blank.jade`))

	root.Use(sessions.Sessions(mysession, sessionStore))

	root.GET("/blank/index", blankHandler)
	root.GET("/ping", pongHandler)
	root.GET("/", simplePugHandler)

	http.Handle("/", root)
}
