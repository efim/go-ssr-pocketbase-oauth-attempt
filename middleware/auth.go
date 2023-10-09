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
// front end side of authentication:
// in base.gohtml template, in <nav> bar
// js code uses SDK for pocketbase to handle oauth calls to backend.
// Also custom event
// in oauth js code
//           document.body.dispatchEvent(new Event("auth-change-event"));
// and in logout route
// 			c.Response().Header().Add("HX-Trigger", "auth-change-event")
// trigger hx-get on <body>
// so that on successful auth and logout the page would refresh
// This is suboptimal in that 3 places:
// <body> with hx-get, js code with `dispatchEvent` and logout route with
// HX-Trigger share responsibility for this piece of logic.  For some reason
// returning HX-Trigger from auth routes via middleware doesn't trigger event on
// htmx side, maybe because these reqeusts are done through js and not directly
// by user in browser.  Or maybe this would be considered a bug on htmx side and
// system could be simplified to just use HX-Trigger response header.  Or some
// other way to simplify


// registeres on pocketbase middleware that
// Sets and Reads session data into a secure cookie
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
	// fires for admin authentication
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

