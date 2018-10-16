package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/ikasamt/zapp/zapp"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/mail"
)

func passwordResetSentHandler(c *gin.Context) {}

func passwordResetCreateHandler(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)
	db, _ := NewGormDB(c)
	defer db.Close()

	variables := map[string]interface{}{}

	var passwordReset PasswordReset
	if c.Request.Method == `POST` {

		// 対象ユーザー取得
		var user User
		email := c.PostForm("email")
		db.Debug().Where("email = ?", email).First(&user)
		passwordReset.UserID = user.ID
		passwordReset.Token = zapp.RandomString(10)
		passwordReset.Validate()

		// エラーあり
		if passwordReset.errors != nil {
			log.Debugf(ctx, "passwordReset.errors: %v", passwordReset.errors)
			variables[`error_message`] = `メールアドレスが見つからないか形式が間違えています。半角で正しい形式で入力してください`
			variables[`email`] = email
			c.Set(`variables`, variables)
			return
		}

		// エラーなし
		SavePasswordReset(db, &passwordReset)
		// メールを送信する処理
		subject := "[Teame] パスワードの再設定"
		body := fmt.Sprintf(`
			Teame　事務局です。
			パスワードの再設定はこちらのこのURLをクリックしてください。パスワードを設定します。
			https://%s/password_reset/change/%s
			`, c.Request.Host, passwordReset.Token)
		log.Debugf(ctx, "body: %v", body)
		toEmail := passwordReset.User().Email
		sendGridFromAddress := zappEnvironment[`SendGridFromAddress`].(string)
		msg := &mail.Message{
			Sender:  sendGridFromAddress,
			To:      []string{toEmail},
			Subject: subject,
			Body:    body,
		}
		if err := mail.Send(ctx, msg); err != nil {
			log.Errorf(ctx, "Couldn't send email: %v", err)
		}

		c.Redirect(301, `/password_reset/sent`)
	}
}

func passwordResetChangeHandler(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)
	db, _ := NewGormDB(c)
	defer db.Close()
	variables := map[string]interface{}{}

	// passwordResetレコードを取得する
	token := c.Param(`token`)
	var passwordReset PasswordReset
	db.Debug().Where(`token = ?`, token).First(&passwordReset)

	log.Debugf(ctx, "%v", passwordReset)
	log.Debugf(ctx, "%v", token)

	// Statusを見る
	if passwordReset.Status != 0 {
		// Error表示
		variables[`instance`] = passwordReset
		variables[`error`] = `パスワードはすでに変更されています。もう一度パスワードリセットの手順を行ってください`
		c.Set(`variables`, variables)
		return
	}

	if c.Request.Method == `POST` {
		user := passwordReset.User()
		newPassword := c.PostForm("new_password")
		if !user.ValidPassword(newPassword) {
			// 間違ってたら表示
			variables[`instance`] = passwordReset
			variables[`error`] = `パスワードの形式が正しくありません`
			c.Set(`variables`, variables)
			return
		}

		// ユーザーを保存する
		salt := zappEnvironment[`password_salt`].(string)
		user.HashedPassword = zapp.HashPassword(salt, newPassword)
		SaveUser(db, &user)

		// パスワードリセットのステータスを変える(使えなくなる)
		db.Debug().Model(&passwordReset).Update("status", 1)

		// ログインする
		zapp.SetSession(c, `user_id`, user.ID)
		c.Redirect(301, `/`)
		return
	}

	//　GET
	variables[`instance`] = passwordReset
	c.Set(`variables`, variables)
}
