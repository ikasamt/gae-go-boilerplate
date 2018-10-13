package app

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

type appengineLogger struct {
	context context.Context
}

func (l appengineLogger) Print(v ...interface{}) {
	log.Debugf(l.context, fmt.Sprintf("%v", v))
}

func NewGormDB(c *gin.Context) (db *gorm.DB, err error) {
	ctx := appengine.NewContext(c.Request)

	uri := productionCLOUDSQL
	if appengine.IsDevAppServer() {
		uri = localCLOUDSQL
	}
	db, err = gorm.Open("mysql", uri)
	if err != nil {
		log.Errorf(ctx, "NewGormDB: %v", err)
	}

	db.LogMode(true)
	db.SetLogger(&appengineLogger{context: ctx})
	return
}
