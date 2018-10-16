package app

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/ikasamt/zapp/zapp"
	"github.com/jinzhu/gorm"
)

func (x User) Ngrams() []string {
	ngrams := []string{}
	ngrams = append(ngrams, zapp.SplitNgramsRange(x.Email, 3)...)
	ngrams = append(ngrams, zapp.SplitNgramsRange(x.Name, 3)...)
	return ngrams
}

// users
type User struct {
	ID             int
	OrganizationID int
	Name           string
	NameKana       string
	Email          string
	HashedPassword string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	beforeJSON     gin.H
	errors         error
}

// String
func (x *User) String() string { return fmt.Sprintf(`%d: %s`, x.ID, x.Name) }

// 制約条件
func (x User) Validations() error {
	return validation.ValidateStruct(&x,
		validation.Field(&x.OrganizationID, validation.Required),
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.Email, validation.Required),
	)
}

// Setter
func (x *User) Setter(c *gin.Context) {
	x.OrganizationID, _ = strconv.Atoi(zapp.GetParams(c, "organization_id"))
	x.Name = zapp.GetParams(c, "name")
	x.NameKana = zapp.GetParams(c, "name_kana")
	x.Email = zapp.GetParams(c, "email")
	if zapp.GetParams(c, "password") != `` {
		salt := zappEnvironment[`password_salt`].(string)
		x.HashedPassword = zapp.HashPassword(salt, zapp.GetParams(c, "password"))
	}
}

// Search
func (x User) Search(q *gorm.DB) *gorm.DB {
	if x.Email != `` {
		q = q.Where("email LIKE ? ", "%"+x.Email+"%")
	}
	if x.Name != `` {
		q = q.Where("name LIKE ? ", "%"+x.Name+"%")
	}
	if x.OrganizationID != 0 {
		q = q.Where("organization_id = ?", x.OrganizationID)
	}
	return q
}

func (x User) ValidPassword(password string) bool {
	// メールアドレスとパスワードが同じは　ダメ
	if x.Email == password {
		return false
	}

	// ７文字以下なら　ダメ
	if len(password) <= 7 {
		return false
	}

	// ３種類すべて含む
	if regexp.MustCompile(`[a-z]`).MatchString(password) && // 英小文字　含むか
		regexp.MustCompile(`[A-Z]`).MatchString(password) && // 英大文字　含むか
		regexp.MustCompile(`[0-9]`).MatchString(password) { // 数字　含むか
		return true
	}
	return false
}
