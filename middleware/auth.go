package middleware

import (
	"fmt"
	"log"
	"net/http"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tokens"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/spf13/cast"
)

const AuthCookieName = "Auth"

func AddCookieSessionMiddleware(app *pocketbase.PocketBase) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.Use(loadAuthContextFromCookie(app))

		return nil
	})

	// fires for every auth collection
    app.OnRecordAuthRequest().Add(func(e *core.RecordAuthEvent) error {
        log.Println(e.HttpContext)
        log.Println(e.Record)
        log.Println(e.Token)
        log.Println(e.Meta)
		e.HttpContext.SetCookie(&http.Cookie{
			Name: AuthCookieName,
			Value: e.Token,
			Path: "/",
		})
		e.HttpContext.SetCookie(&http.Cookie{
			Name: "username",
			Value: e.Record.Username(),
		})
        return nil
    })
	app.OnAdminAuthRequest().Add(func(e *core.AdminAuthEvent) error {
        log.Println(e.HttpContext)
        log.Println(e.Admin)
        log.Println(e.Token)
		e.HttpContext.SetCookie(&http.Cookie{
			Name: AuthCookieName,
			Value: e.Token,
			Path: "/",
		})
        return nil
    })

}

func loadAuthContextFromCookie(app core.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenCookie, err := c.Request().Cookie(AuthCookieName)
			if err != nil || tokenCookie.Value == "" {
				return next(c) // no token cookie
			}

			token := tokenCookie.Value

			claims, _ := security.ParseUnverifiedJWT(token)
			tokenType := cast.ToString(claims["type"])

			switch tokenType {
			case tokens.TypeAdmin:
				admin, err := app.Dao().FindAdminByToken(
					token,
					app.Settings().AdminAuthToken.Secret,
				)
				if err == nil && admin != nil {
					// "authenticate" the admin
					c.Set(apis.ContextAdminKey, admin)
					someData := struct {
						username string
						email string
					} {
						admin.Email,
						admin.Created.String(),
					}
					fmt.Printf("triggering the middlewar for cookie %v and err %v\n", someData, err)
				}

			case tokens.TypeAuthRecord:
				record, err := app.Dao().FindAuthRecordByToken(
					token,
					app.Settings().RecordAuthToken.Secret,
				)
				if err == nil && record != nil {
					// "authenticate" the app user
					c.Set(apis.ContextAuthRecordKey, record)
					someData := struct {
						username string
						email string
					} {
						record.Username(),
						record.Email(),
					}
					fmt.Printf("triggering the middlewar for cookie %v and err %v\n", someData, err)

				}
			}

			return next(c)
		}
	}
}
