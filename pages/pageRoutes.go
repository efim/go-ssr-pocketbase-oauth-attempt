package pages

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

//go:embed templates
var templatesFS embed.FS

func AddPageRoutes(app *pocketbase.PocketBase) {
	app.OnBeforeServe().Add(getIndexPageRoute(app))
}

// render and return index page
func getIndexPageRoute(app *pocketbase.PocketBase)  func(*core.ServeEvent) error {
	return func (e *core.ServeEvent) error {
		e.Router.GET("/", func(c echo.Context) error {
			// first collect data
			info   := apis.RequestInfo(c)
			admin  := info.Admin      // nil if not authenticated as admin
			record := info.AuthRecord // nil if not authenticated as regular auth record

			isGuest := admin == nil && record == nil
			coolMessage := fmt.Sprintf("got admin %v and record %v. is guest: %t", admin, record, isGuest)
			fmt.Print(coolMessage)

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
			fmt.Printf(">> enabled providers names %+v\n", oauthProviderNames)

			indexPageData := struct {
				IsGuest, IsAdmin bool
				Username string
				EnabledOauthProviders []string
			}{
				IsGuest: isGuest,
				IsAdmin: admin != nil,
				Username: username,
				EnabledOauthProviders: oauthProviderNames,
			}

			// then render template with it
			templateName := "templates/index.gohtml"
			tmpl := template.Must(template.ParseFS(templatesFS, templateName))
			var instantiatedTemplate bytes.Buffer
			if err := tmpl.Execute(&instantiatedTemplate, indexPageData); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "error parsing template"})
			}

			return c.HTML(http.StatusOK, instantiatedTemplate.String())
		})
		return nil
	}
}
