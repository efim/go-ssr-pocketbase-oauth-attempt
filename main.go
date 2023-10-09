package main

import (
	"log"
	"github.com/pocketbase/pocketbase"
	"sunshine.industries/auth-pocketbase-attempt/middleware"
	"sunshine.industries/auth-pocketbase-attempt/pages"
)

func main() {
	app := pocketbase.New()

	middleware.AddCookieSessionMiddleware(app)
	middleware.AddErrorsMiddleware(app)
	pages.AddPageRoutes(app)

	// starts the pocketbase backend
	// parses cli arguments for hostname and data dir
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
