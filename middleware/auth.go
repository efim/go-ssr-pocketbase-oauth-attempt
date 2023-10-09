package middleware

import (
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
		e.HttpContext.SetCookie(&http.Cookie{
			Name: AuthCookieName,
			Value: e.Token,
			Path: "/",
			Secure: true,
			HttpOnly: true,
		})
		e.HttpContext.SetCookie(&http.Cookie{
			Name: "username",
			Value: e.Record.Username(),
		})
        return nil
    })
	app.OnAdminAuthRequest().Add(func(e *core.AdminAuthEvent) error {
		e.HttpContext.SetCookie(&http.Cookie{
			Name: AuthCookieName,
			Value: e.Token,
			Path: "/",
			Secure: true,
			HttpOnly: true,
		})
        return nil
    })
	app.OnBeforeServe().Add(getLogoutRoute(app))
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
				}

			case tokens.TypeAuthRecord:
				record, err := app.Dao().FindAuthRecordByToken(
					token,
					app.Settings().RecordAuthToken.Secret,
				)
				if err == nil && record != nil {
					// "authenticate" the app user
					c.Set(apis.ContextAuthRecordKey, record)
				}
			}

			return next(c)
		}
	}
}

// render and return login page with configured oauth providers
func getLogoutRoute(app *pocketbase.PocketBase)  func(*core.ServeEvent) error {
	return func (e *core.ServeEvent) error {
		e.Router.GET("/logout", func(c echo.Context) error {
			c.SetCookie(&http.Cookie{
				Name: AuthCookieName,
				Value: "",
				Path: "/",
				MaxAge: -1,
				Secure: true,
				HttpOnly: true,
			})
			c.Response().Header().Add("HX-Trigger", "auth-change-event")
			return c.NoContent(http.StatusOK)
		})
		return nil
	}
}

