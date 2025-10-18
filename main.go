package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/template"
)

var views embed.FS

type Comment struct {
	Comment string
}

func main() {
	app := pocketbase.New()

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {

		registry := template.NewRegistry()

		// serves static files from the provided public dir (if exists)
		se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))

		se.Router.GET("/hello", func(e *core.RequestEvent) error {
			html, err := registry.LoadFiles(
				"views/layout.html",
				"views/hello_title.html",
				"views/hello_body.html",
			).Render(map[string]any{})

			if err != nil {
				// or redirect to a dedicated 404 HTML page
				return e.NotFoundError("", err)
			}

			return e.HTML(http.StatusOK, html)
		})

		se.Router.GET("/hello/{name}", func(e *core.RequestEvent) error {
			name := e.Request.PathValue("name")

			html, err := registry.LoadFiles(
				"views/hello.html",
			).Render(map[string]any{"name": name})

			if err != nil {
				// or redirect to a dedicated 404 HTML page
				return e.NotFoundError("", err)
			}

			return e.HTML(http.StatusOK, html)
		})

		se.Router.GET("/htmx/comments", func(e *core.RequestEvent) error {
			comments := []Comment{}
			app.DB().NewQuery("select comment from comments").All(&comments)

			html, err := registry.LoadFiles(
				"views/comments.html",
			).Render(map[string]any{"Comments": comments})

			if err != nil {
				// or redirect to a dedicated 404 HTML page
				return e.NotFoundError("", err)
			}

			return e.HTML(http.StatusOK, html)
		})

		app.OnRecordCreateRequest("comments").BindFunc(func(e *core.RecordRequestEvent) error {

			fmt.Println(e.Auth.Get("id"))

			e.Record.Set("user", e.Auth.Get("id"))

			return e.Next()
		})

		// app.OnRecordAuthRequest().BindFunc(func(e *core.RecordAuthRequestEvent) error {
		// 	if e.Token == "" || e.Record == nil {
		// 		return nil // nothing to do
		// 	}

		// 	cookieName := "pb_auth"
		// 	cookieValue := e.Token
		// 	maxAge := int((7 * 24 * time.Hour).Seconds()) // 7 days

		// 	cookie := fmt.Sprintf(
		// 		"%s=%s; Path=/; HttpOnly; SameSite=Strict; Max-Age=%d; Secure",
		// 		cookieName,
		// 		cookieValue,
		// 		maxAge,
		// 	)

		// 	e.Response.Header().Add("Set-Cookie", cookie)
		// 	return nil
		// })

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
