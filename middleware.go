package app

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func errorMiddleware(layoutName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Writer.Status()

		switch code {
		case 401, 402, 403, 404:
			RenderPug(c, layoutName, `404.jade`)
		case 500:
			RenderPug(c, layoutName, `500.jade`)
		}
		c.Next()
	}
}

func pugMiddleware(layoutName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		_, ok := c.Get("rendered")
		if !ok {
			controller := c.MustGet("controller").(string)
			action := c.MustGet("action").(string)
			templateName := fmt.Sprintf("%s/%s.jade", controller, action)
			RenderPug(c, layoutName, templateName)
		}

	}
}

func paramsMiddleware(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		params := map[string]interface{}{}

		// controller and action
		tmp := strings.TrimPrefix(c.Request.URL.Path, prefix)
		paths := strings.Split(tmp, `/`)

		controller := ``
		action := ``
		if tmp == `/` {
			controller = `top`
			action = `index`
		}
		if len(paths) >= 2 {
			if paths[1] != `` {
				controller = paths[1]
			}
		}
		if len(paths) >= 3 {
			if paths[2] != `` {
				action = paths[2]
			}
		}

		c.Set("controller", controller)
		c.Set("action", action)
		params[`controller`] = c.MustGet("controller")
		params[`action`] = c.MustGet("action")

		// set parameters

		// Query
		for key, value := range c.Request.URL.Query() {
			params[key] = value
		}
		// Params
		for _, p := range c.Params {
			params[p.Key] = p.Value
		}
		// PostForm
		req := c.Request
		req.ParseForm()
		req.ParseMultipartForm(8 << 20)
		for key, value := range req.PostForm {
			params[key] = value
		}
		c.Set("params", params)

		c.Next()
	}
}

func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()
		finishedAt := time.Now()
		delta := finishedAt.Sub(startedAt)
		log.Println(fmt.Sprintf(`[delta:%s]`, delta))
	}
}
