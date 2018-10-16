package clefs

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ikasamt/zapp/zapp"
)

type Something struct { //generic.Type
	ID         int   //generic.Type
	UserID     int   //generic.Type
	AnythingID int   //generic.Type
	errors     error //generic.Type
} //generic.Type

func (x Anything) Somethings() (instances []Something) {
	db, _ := NewGormDBSimple()
	defer db.Close()
	lower := strings.ToLower(reflect.TypeOf(x).Name())
	db.Debug().Where(lower+`_id = ?`, x.ID).Find(&instances)
	return
}

//
func (x Something) AnythingOptions() string {
	tmp := ""
	tmp += "<option>&nbsp;</option>"
	for _, any := range selectAllAnythings("id, name") {
		ID_ := reflect.ValueOf(x.AnythingID)
		var ID int
		if ID_.Type().Kind() == reflect.Ptr {
			if !ID_.IsNil() {
				ID = int(ID_.Elem().Int())
			}
		} else {
			ID = int(ID_.Int())
		}

		label := zapp.CallMethod(any, `String`, fmt.Sprintf(`[%04d]`, any.ID))

		if ID == any.ID {
			tmp += fmt.Sprintf("<option value='%d' selected>%s</option>", any.ID, label)
		} else {
			tmp += fmt.Sprintf("<option value='%d'>%s</option>", any.ID, label)
		}
	}
	return tmp
}
