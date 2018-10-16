package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/ikasamt/zapp/zapp"
	"google.golang.org/appengine"
)

// Auth Middleware
func UserAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {

		// ログインしてなかったらログイン画面に飛ばす
		currentUserID := zapp.GetSession(c, "user_id", 0).(int)
		if currentUserID == 0 {
			url := fmt.Sprintf("/user/login?next_url=%s", c.Request.URL)
			http.Redirect(c.Writer, c.Request, url, http.StatusFound)
			c.Abort()
			return
		}

		// DB接続を取得
		db, _ := NewGormDB(c)
		defer db.Close()

		// ユーザーを取得
		var currentUser User
		db.Debug().Where("id = ?", currentUserID).First(&currentUser)
		c.Set(`me`, currentUser)
	}
}

var zappEnvironment zapp.Environment

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	zappEnvironments, err := zapp.ReadEnvironments()
	if err != nil {
		log.Println(err)
		return
	}

	zappEnvironment = zappEnvironments[`production`]
	if appengine.IsDevAppServer() {
		zappEnvironment = zappEnvironments[`development`]
	}

	sessionSalt := zappEnvironment[`session_salt`].(string)
	sessionStore := cookie.NewStore([]byte(sessionSalt))

	root := gin.Default()

	root.Use(loggingMiddleware())
	root.Use(paramsMiddleware(``))
	root.Use(errorMiddleware(`layout_blank.jade`))
	root.Use(sessions.Sessions(`mysession`, sessionStore))

	// ログイン前の画面設定
	root.Use(
		pugMiddleware(`layout_before_login.jade`),
	)
	{
		root.GET("/", simplePugHandler)
		root.GET("/ping", pongHandler)
		root.GET("/dev/widgets", simplePugHandler)
		root.GET("/blank/index", blankHandler)

		root.GET("/user/login", userLoginHandler)
		root.POST("/user/login", userLoginHandler)
		root.GET("/password_reset/new", passwordResetCreateHandler)
		root.POST("/password_reset/new", passwordResetCreateHandler)
		root.GET("/password_reset/sent", passwordResetSentHandler)
		root.GET("/password_reset/change/:token", passwordResetChangeHandler)
		root.POST("/password_reset/change/:token", passwordResetChangeHandler)
	}

	// ログイン後の画面設定
	root.Use(
		UserAuthRequired(),
		pugMiddleware(`layout.jade`),
	)
	{
		root.GET("/user/logout", userLogoutHandler)
		root.GET("/user/search", userSearchHandler)
		root.GET("/user/show/1", userShowHandler)
		root.POST("/user/put_fulltext", userPutFulltext)

	}

	http.Handle("/", root)
}
