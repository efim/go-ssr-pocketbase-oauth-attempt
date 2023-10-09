package pages

import (
	"bytes"
	"embed"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)
// template files are bundled with binary
// for worry free deployment that needs to copy a single file

//go:embed templates
var templatesFS embed.FS

// static files are bundled into separate FS
// because full content of that embed.FS is available
// under http://127.0.0.1:8090/static/static/public/

//go:embed static
var staticFilesFS embed.FS

// registers site pages, to be served by pocketbase
// passes `app` to allow access to `DAO` and other apis
// each page will get auth data in request context
// and will be able to create all necessary info for page render:
// user data, external api calls, calculations
func AddPageRoutes(app *pocketbase.PocketBase) {
	app.OnBeforeServe().Add(getIndexPageRoute(app))
	app.OnBeforeServe().Add(getSomePageRoute(app))
	app.OnBeforeServe().Add(getErrorPageRoute(app))

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
			navInfoData := initNavInfoData(app, c)
			indexPageData := struct {
				BackendMessage string
				NavInfo        navInfo
			}{
				BackendMessage: "Hello from the backend!",
				NavInfo:        navInfoData,
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
func getSomePageRoute(app *pocketbase.PocketBase) func(*core.ServeEvent) error {
	return func(e *core.ServeEvent) error {
		e.Router.GET("/somepage", func(c echo.Context) error {
			// get data
			// and since i'm using 'base.gohtml' with Nav, i'll need Nav info
			navInfoData := initNavInfoData(app, c)

			somePageData := struct {
				RandomNumber int
				RandomString string
				NavInfo      navInfo
			}{
				RandomNumber: rand.Int(),
				RandomString: stringWithCharset(25, charset),
				NavInfo:      navInfoData,
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
}

func getErrorPageRoute(app *pocketbase.PocketBase) func(*core.ServeEvent) error {
	return func(e *core.ServeEvent) error {
		e.Router.GET("/error/:code", func(c echo.Context) error {
			// get data
			code := c.PathParam("code")
			codeNum, err := strconv.ParseInt(code, 10, 64)
			if err != nil {
				codeNum = 500
			}
			errorText := http.StatusText(int(codeNum))
			if errorText == "" {
				codeNum = 500
				errorText = http.StatusText(500)
			}

			// and since i'm using 'base.gohtml' with Nav, i'll need Nav info
			navInfoData := initNavInfoData(app, c)

			somePageData := struct {
				NavInfo   navInfo
				ErrorCode int64
				ErrorText string
			}{
				NavInfo:   navInfoData,
				ErrorCode: codeNum,
				ErrorText: errorText,
			}

			// then render template with it
			templateName := "templates/errors/error.gohtml"
			switch codeNum {
			case 404:
				templateName = "templates/errors/404.gohtml"
			case 401:
				templateName = "templates/errors/401.gohtml"
			}
			tmpl := template.Must(template.ParseFS(templatesFS, "templates/base.gohtml", templateName))
			var instantiatedTemplate bytes.Buffer
			if err := tmpl.Execute(&instantiatedTemplate, somePageData); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "error parsing template"})
			}

			return c.HTML(int(codeNum), instantiatedTemplate.String())
		})
		return nil
	}
}

// initializing data which is used by any page that has <nav> bar
// - whether user is already authenticated
// - which authentication methods are available
// this is used in templates/base.gohtml
func initNavInfoData(app *pocketbase.PocketBase, c echo.Context) navInfo {
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

	return navInfo{
		IsGuest:               isGuest,
		Username:              username,
		EnabledOauthProviders: oauthProviderNames,
	}
}
