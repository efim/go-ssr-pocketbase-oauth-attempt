package middleware

import (
	"bytes"
	"html/template"
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func AddErrorsMiddleware(app *pocketbase.PocketBase) {
	app.OnBeforeApiError().Add(func(e *core.ApiErrorEvent) error {
		log.Printf("in before api error with %+v with response %v and error %+v", e, e.HttpContext.Response(), e.Error)
		// oh, i guess i could do redirect?
		return renderErrorPage(e)
	})
}

var redirectTemplate = `
<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="refresh" content="0; url='/error/{{ . }}'" />
  </head>
  <body>
    <p>Redirecting to error page</p>
  </body>
</html>
`
var tmpl = template.Must( template.New("redirect-to-error").Parse(redirectTemplate) )

func renderErrorPage(e *core.ApiErrorEvent) error {
	errorMessage := e.Error.Error()
	log.Printf("in error to html middleware for %s with status %+v", errorMessage, e)

	errorCode := 500
	switch errorMessage {
	case "Not Found.":
		// not authorized
		errorCode = 404
	case "The request requires admin or record authorization token to be set.":
		// not found
		errorCode = 401
	}

	var instantiatedTemplate bytes.Buffer
	if err := tmpl.Execute(&instantiatedTemplate, errorCode); err != nil {
		// couldn't execute the template
		return e.HttpContext.HTML(200, "Error 500")
	}

	return e.HttpContext.HTML(200, instantiatedTemplate.String())
}
