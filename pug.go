package app

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Joker/jade"
	humanize "github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/golang/go/src/html/template"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

var funcMap = template.FuncMap{
	"Comma": func(value interface{}) string {
		switch t := value.(type) {
		case int:
			v := value.(int)
			return humanize.Comma(int64(v))
		case int64:
			v := value.(int64)
			return humanize.Comma(v)
		case float64:
			v := value.(float64)
			return humanize.Comma(int64(v))
		default:
			return fmt.Sprintf(`unknown comma error(%s)`, t)
		}
	},
}

func RenderError(c *gin.Context, err error) {
	msg := `ERROR`
	if strings.Contains(fmt.Sprint(err), `no such file or directory`) {
		msg = `404 NOT FOUND`
	}
	c.String(http.StatusInternalServerError, msg)
	return
}

func respondWithError(c *gin.Context, code int, message string) {
	c.AbortWithStatus(code)
	c.Set(`errorMessage`, message)
}

func RenderPug(c *gin.Context, layoutName string, templateName string) {
	ctx := appengine.NewContext(c.Request)

	layoutHTML, err := jade.ParseFile("views/" + layoutName)
	if err != nil {
		log.Errorf(ctx, "Error: %v", err)
		respondWithError(c, 500, err.Error())
		return
	}

	contentHTML, err := jade.ParseFile("views/" + templateName)
	if err != nil {
		log.Errorf(ctx, "Error: %v", err)
		respondWithError(c, 500, err.Error())
		return
	}

	params := c.MustGet("params").(map[string]interface{})
	variables := c.MustGet("variables").(map[string]interface{})
	for k, v := range variables {
		params[k] = v
	}

	t, err := template.New("layout").Funcs(funcMap).Parse(layoutHTML)
	t.New("tmpl").Parse(contentHTML)
	if err := t.Execute(c.Writer, params); err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprint(err))
	}
}
