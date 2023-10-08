package pages

import (
	"bytes"
	"embed"
	"html/template"
	"math/rand"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

//go:embed templates
var templatesFS embed.FS

//go:embed static
var staticFilesFS embed.FS

func AddPageRoutes(app *pocketbase.PocketBase) {
	app.OnBeforeServe().Add(getIndexPageRoute(app))
	app.OnBeforeServe().Add(somePageRoute)

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.StaticFS("/static", staticFilesFS)
		// this path works : http://127.0.0.1:8090/static/static/public/htmx.min.js
		return nil
	})
}

type navInfo struct {
	Username              string
	IsGuest               bool
	EnabledOauthProviders []string
}

// render and return some page
func getIndexPageRoute(app *pocketbase.PocketBase) func(*core.ServeEvent) error {
	return func(e *core.ServeEvent) error {
		e.Router.GET("/", func(c echo.Context) error {
			// first collect data
			info := apis.RequestInfo(c)
			admin := info.Admin       // nil if not authenticated as admin
			record := info.AuthRecord // nil if not authenticated as regular auth record

			isGuest := admin == nil && record == nil

			username := ""
			switch {
			case admin != nil:
				username = admin.Email
			case record != nil:
				username = record.Username()
			}

			oauthProviders := app.Settings().NamedAuthProviderConfigs()
			oauthProviderNames := make([]string, 0, len(oauthProviders))
			for name, config := range oauthProviders {
				if config.Enabled {
					oauthProviderNames = append(oauthProviderNames, name)
				}
			}

			indexPageData := struct {
				IsGuest, IsAdmin      bool
				Username              string
				EnabledOauthProviders []string
				NavInfo               navInfo
			}{
				IsAdmin: admin != nil,
				NavInfo: navInfo{
					IsGuest:               isGuest,
					Username:              username,
					EnabledOauthProviders: oauthProviderNames,
				},
			}

			// then render template with it
			templateName := "templates/index.gohtml"
			tmpl := template.Must(template.ParseFS(templatesFS, "templates/base.gohtml", templateName))
			var instantiatedTemplate bytes.Buffer
			if err := tmpl.Execute(&instantiatedTemplate, indexPageData); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "error parsing template"})
			}

			return c.HTML(http.StatusOK, instantiatedTemplate.String())
		})
		return nil
	}
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
func somePageRoute(e *core.ServeEvent) error {
	e.Router.GET("/somepage", func(c echo.Context) error {
		// get data
		// and since i'm using 'base.gohtml' with Nav, i'll need Nav info

		info := apis.RequestInfo(c)
		admin := info.Admin       // nil if not authenticated as admin
		record := info.AuthRecord // nil if not authenticated as regular auth record

		username := ""
		switch {
		case admin != nil:
			username = admin.Email
		case record != nil:
			username = record.Username()
		}

		somePageData := struct {
			RandomNumber int
			RandomString string
			NavInfo      navInfo
		}{
			RandomNumber: rand.Int(),
			RandomString: stringWithCharset(25, charset),
			NavInfo: navInfo{
				Username: username,
			},
		}

		// then render template with it
		templateName := "templates/somepage.gohtml"
		tmpl := template.Must(template.ParseFS(templatesFS, "templates/base.gohtml", templateName))
		var instantiatedTemplate bytes.Buffer
		if err := tmpl.Execute(&instantiatedTemplate, somePageData); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "error parsing template"})
		}

		return c.HTML(http.StatusOK, instantiatedTemplate.String())
	}, apis.RequireAdminOrRecordAuth())
	return nil
}
