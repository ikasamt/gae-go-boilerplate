package clefs

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func NewGormDB(c *gin.Context) (db *gorm.DB, err error) { //generic.Type
	db, err = gorm.Open("mysql", ``) //generic.Type
	return db, err                   //generic.Type
} //generic.Type

func NewGormDBSimple() (db *gorm.DB, err error) { //generic.Type
	db, err = gorm.Open("mysql", ``) //generic.Type
	return db, err                   //generic.Type
} //generic.Type

// Anything is defined as base Struct for genny
type Anything struct { //generic.Type
	ID         int       //generic.Type
	UserID     int       //generic.Type
	CreatedAt  time.Time //generic.Type
	UpdatedAt  time.Time //generic.Type
	errors     error     //generic.Type
	beforeJSON gin.H     //generic.Type
} //generic.Type

func (any *Anything) AsJSON() gin.H { //generic.Type
	return gin.H{ //generic.Type
		"id":         any.ID,        //generic.Type
		"created_at": any.CreatedAt, //generic.Type
		"updated_at": any.UpdatedAt, //generic.Type
	} //generic.Type
} //generic.Type
