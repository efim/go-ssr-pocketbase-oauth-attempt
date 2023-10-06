package main

import (
    "log"
    // "os"

    "github.com/pocketbase/pocketbase"
    // "github.com/pocketbase/pocketbase/apis"
    // "github.com/pocketbase/pocketbase/core"
)

func main() {
	app := pocketbase.New()

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
