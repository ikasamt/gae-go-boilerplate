package app

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/ikasamt/zapp/zapp"
	"github.com/jinzhu/gorm"
)

type Organization struct {
	ID         int
	Name       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	beforeJSON gin.H
	errors     error
}

func (x Organization) String() string {
	return fmt.Sprintf(`[%d] %s`, x.ID, x.Name)
}

func (x Organization) Validations() error {
	return validation.ValidateStruct(&x,
		validation.Field(&x.Name, validation.Required),
	)
}

func (x *Organization) Setter(c *gin.Context) {
	x.Name = zapp.GetParams(c, "name")
}

// 検索条件
func (x Organization) Search(q *gorm.DB) *gorm.DB {
	if x.Name != `` {
		q = q.Where("name LIKE ?", "%"+x.Name+"%")
	}
	return q
}
