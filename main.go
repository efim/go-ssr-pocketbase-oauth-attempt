package main

import (
	"log"
	"strings"

	"github.com/pocketbase/pocketbase"
	"sunshine.industries/auth-pocketbase-attempt/middleware"
	"sunshine.industries/auth-pocketbase-attempt/pages"
)

func main() {
	app := pocketbase.New()

	servedName := app.Settings().Meta.AppUrl
	isTlsEnabled := strings.HasPrefix(servedName, "https://")

	middleware.AddCookieSessionMiddleware(app, isTlsEnabled)
	pages.AddPageRoutes(app)

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
