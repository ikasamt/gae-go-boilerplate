package clefs

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/iancoleman/strcase"
	"github.com/ikasamt/zapp/zapp"
	"github.com/jinzhu/gorm"
)

// select All
func selectAllAnythings(selects string) (instances []*Anything) {
	db, _ := NewGormDBSimple()
	defer db.Close()
	db.Debug().Select(selects).Find(&instances)
	return
}

// fetch One
func fetchAnything(anyID int) (any Anything) {
	db, _ := NewGormDBSimple()
	defer db.Close()
	db.Debug().Where("id = ?", anyID).First(&any)
	any.beforeJSON = zapp.CallMethod(any, `AsJSON`, gin.H{}).(gin.H)
	return any
}

func getAnything(c *gin.Context) (instance Anything, e error) {
	// 対象IDを取得
	ID, err := zapp.GetID(c)
	if err != nil {
		log.Println(err)
		return instance, err
	}
	// DBから取得
	instance = fetchAnything(ID)
	return instance, nil
}

func (any *Anything) Set(c *gin.Context) {
	r := reflect.ValueOf(any)
	method := r.MethodByName("Setter") // これは個別に定義したいので回避
	if method.IsValid() {              // メソッドがある場合のみ実行
		method.Call([]reflect.Value{reflect.ValueOf(c)})
	}
}

func (any *Anything) Validate() {
	r := reflect.ValueOf(any)
	method := r.MethodByName("Validations") // これは個別に定義したいので回避
	if method.IsValid() {                   // メソッドがある場合のみ実行
		response := method.Call(nil)[0]
		// nilじゃなければ
		if reflect.ValueOf(response.Interface()).IsValid() {
			any.errors = response.Interface().(error)
		}
	}
}

func (any Anything) Diff() gin.H {
	// 現在と元のデータの違い
	var afterJSON gin.H
	j := zapp.CallMethod(any, `AsJSON`, nil)
	if j != nil {
		afterJSON = j.(gin.H)
	}
	if any.ID == 0 {
		return afterJSON
	}

	diff := make(gin.H)
	for k, before := range any.beforeJSON {
		after := afterJSON[k]
		if !reflect.DeepEqual(before, after) {
			diff[k] = []interface{}{before, after}
		}
	}
	//
	return diff
}

func (any Anything) GetErrors(key string) []string {
	if any.errors == nil {
		return []string{}
	}

	// エラー文字列を分割する
	_errors := make(map[string][]string)
	for _, e := range strings.Split(any.errors.Error(), `;`) {
		tmp := strings.Split(e, `:`)
		k := strings.TrimLeft(tmp[0], " ")
		v := strings.TrimLeft(tmp[1], " ")
		v = strings.TrimRight(v, ".")
		switch v {
		case `must be in a valid format`:
			v = `正しい形式ではありません`
		case `cannot be blank`:
			v = `必須項目です`
		}
		if _errors[k] == nil {
			_errors[k] = []string{}
		}
		_errors[k] = append(_errors[k], v)
	}
	return _errors[key]
}

// GetValueAndName
func (any Anything) GetValueAndName(key string) (value reflect.Value, name string) {
	value = reflect.ValueOf(&any).Elem().FieldByName(key)
	name = strcase.ToSnake(strings.Replace(key, "ID", "Id", 1))
	return
}

// CheckboxField
func (any Anything) CheckboxField(key string) template.HTML {
	value, name := any.GetValueAndName(key)

	if value.Interface().(bool) {
		return template.HTML(fmt.Sprintf("<input type='checkbox' name='%s' checked='checked' />", name))
	}
	return template.HTML(fmt.Sprintf("<input type='checkbox' name='%s' />", name))
}

// TextField
func (any Anything) TextField(key string) template.HTML {
	value, name := any.GetValueAndName(key)

	inputText := fmt.Sprintf("<input type='text' name='%s' value='%v' />", name, value)
	if messages := any.GetErrors(key); len(messages) > 0 {
		retval := "<div class='field_with_errors'>"
		retval += inputText
		retval += "<br/>"
		retval += fmt.Sprintf("<span class='field_with_errors_message'>%s</span>", strings.Join(messages, ","))
		retval += "</div>"
		return template.HTML(retval)
	}
	return template.HTML(inputText)
}

// SelectField
func (any Anything) SelectField(key string, options string) template.HTML {
	_, name := any.GetValueAndName(key)

	inputText := ``
	inputText += fmt.Sprintf("<select class='select2' name='%s' style='width: 100%%'>", name)
	inputText += options
	inputText += "<br/>"
	inputText += "</select>"
	if messages := any.GetErrors(key); len(messages) > 0 {
		retval := "<div class='field_with_errors'>"
		retval += inputText
		retval += fmt.Sprintf("<span class='field_with_errors_message'>%s</span>", strings.Join(messages, ","))
		retval += "</div>"
		return template.HTML(retval)
	}
	return template.HTML(inputText)
}

// SaveAnything 履歴保存
func SaveAnything(db *gorm.DB, any *Anything) *gorm.DB {
	history := zapp.History{}
	history.Model = reflect.TypeOf(any).Elem().Name()
	history.InstanceID = any.ID
	bytes, _ := json.Marshal(any.Diff())
	history.Data = string(bytes)
	if history.Data != `{}` {
		db.Debug().Save(&history)
	}
	db.Debug().Save(&any)
	return db
}

func searchAnythings(c *gin.Context, any Anything) (count int, instances []Anything) {
	// DB接続を取得
	db, _ := NewGormDB(c)
	defer db.Close()

	// 検索条件
	db = db.Debug()

	method := zapp.GetMethod(any, `Search`)
	if method.IsValid() {
		db = method.Call([]reflect.Value{reflect.ValueOf(db)})[0].Interface().(*gorm.DB)
	}
	perPage, err := strconv.Atoi(zapp.GetParams(c, `per_page`))
	if err != nil {
		perPage = zapp.DefaultPerPage
	}
	page, err := strconv.Atoi(zapp.GetParams(c, `page`))
	if err != nil {
		page = 0
	}

	// 件数
	db.Model(any).Count(&count)
	// 検索実行
	db.Offset(page * perPage).Limit(perPage).Order(`id DESC`).Find(&instances)
	return count, instances
}

// List
func adminAnythingListHandler(c *gin.Context) {
	instance := Anything{}
	instance.Set(c)
	totalCount, instances := searchAnythings(c, instance)
	context := map[string]interface{}{"instances": instances, "instance": instance, "total_count": totalCount}
	zapp.Render(c, `admin`, context)
}

// New
func adminAnythingNewHandler(c *gin.Context) {
	instance := Anything{}
	context := map[string]interface{}{"instance": instance}
	zapp.Render(c, `admin`, context)
}

// Edit
func adminAnythingEditHandler(c *gin.Context) {
	instance, _ := getAnything(c)
	context := map[string]interface{}{"instance": instance}
	zapp.Render(c, `admin`, context)
}

// Show
func adminAnythingShowHandler(c *gin.Context) {
	instance, err := getAnything(c)
	if err != nil || instance.ID == 0 {
		log.Println(err)
	}
	context := map[string]interface{}{"instance": instance}
	zapp.Render(c, `admin`, context)
}

// Create
func adminAnythingCreateHandler(c *gin.Context) {
	// DB接続を取得
	db, _ := NewGormDB(c)
	defer db.Close()

	instance := &Anything{}
	instance.Set(c)
	instance.Validate()
	if instance.errors != nil {
		context := map[string]interface{}{"instance": instance}
		zapp.Render(c, `admin`, context, `new`)
		return
	}

	// 値の更新 ------------------------------------------>
	SaveAnything(db, instance)
	// <-------------------------------------------------

	// 完了ページへリダイレクト
	message := fmt.Sprintf("%v 追加しました", instance)
	zapp.SetFlashMessage(c, message)
	adminPrefix := `admin`
	controllerName, _ := zapp.ExtractControllerActionName(c.Request.URL.Path, adminPrefix)
	backURL := fmt.Sprintf("/%s/%s", adminPrefix, controllerName)
	c.Redirect(http.StatusFound, backURL)
}

// Update
func adminAnythingUpdateHandler(c *gin.Context) {
	// DB接続を取得
	db, _ := NewGormDB(c)
	defer db.Close()

	a, err := getAnything(c)
	instance := &a
	if err != nil {
		zapp.RenderDirect(c, `admin/500`, map[string]interface{}{"message": err})
		return
	}

	instance.Set(c)
	instance.Validate()
	if instance.errors != nil {
		context := map[string]interface{}{"instance": instance}
		zapp.Render(c, `admin`, context, `edit`)
		return
	}

	// 値の更新 ------------------------------------------>
	SaveAnything(db, instance)
	// <-------------------------------------------------

	// 完了ページへリダイレクト
	adminPrefix := `admin`
	controllerName, _ := zapp.ExtractControllerActionName(c.Request.URL.Path, adminPrefix)
	backURL := fmt.Sprintf("/%s/%s/show/%d", adminPrefix, controllerName, instance.ID)
	c.Redirect(http.StatusFound, backURL)
}

func AppendAnythingResources(group *gin.RouterGroup) {
	structName := zapp.GetType(Anything{})
	controllerName := strcase.ToSnake(structName)
	log.Println(structName)
	log.Println(controllerName)
	group.GET(fmt.Sprintf("/%s/", controllerName), adminAnythingListHandler)
	group.GET(fmt.Sprintf("/%s/new", controllerName), adminAnythingNewHandler)
	group.GET(fmt.Sprintf("/%s/edit/:id", controllerName), adminAnythingEditHandler)
	group.GET(fmt.Sprintf("/%s/show/:id", controllerName), adminAnythingShowHandler)
	group.POST(fmt.Sprintf("/%s/create", controllerName), adminAnythingCreateHandler)
	group.POST(fmt.Sprintf("/%s/update", controllerName), adminAnythingUpdateHandler)
}
