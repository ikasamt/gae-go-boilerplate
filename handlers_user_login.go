package app

import (
	"github.com/gin-gonic/gin"
	"github.com/ikasamt/zapp/zapp"
)

func userLoginHandler(c *gin.Context) {
	db, _ := NewGormDB(c)
	defer db.Close()

	nextURL := zapp.GetParams(c, `next_url`)
	if nextURL == `` {
		nextURL = `/`
	}

	variables := map[string]interface{}{}

	var user User
	email := zapp.GetParams(c, "email")
	password := zapp.GetParams(c, "password")
	if c.Request.Method == `POST` {
		salt := zappEnvironment[`password_salt`].(string)
		hashedPassword := zapp.HashPassword(salt, password)
		db.Debug().Where(`email = ? AND hashed_password = ?`, email, hashedPassword).First(&user)
		// ログイン成功
		if user.ID != 0 {
			zapp.SetSession(c, `user_id`, user.ID)
			c.Redirect(301, nextURL)
			return
		}
		// ログイン失敗
		variables[`error_message`] = `Emailもしくはパスワードが間違っています`
	}
	//　GET|Error時表示
	variables[`next_url`] = nextURL
	variables[`email`] = email
	c.Set(`variables`, variables)
}

func userLogoutHandler(c *gin.Context) {
	zapp.SetSession(c, `user_id`, nil)
	c.Redirect(301, `/`)
}
