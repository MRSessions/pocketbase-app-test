package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	_ "pocketbase-server/migrations"
)

func main() {
	app := pocketbase.New()

	migratecmd.MustRegister(app, app.RootCmd, &migratecmd.Options{
		Automigrate: true, // auto creates migration files when making collection changes
	})

	var publicDir string
	app.RootCmd.PersistentFlags().StringVar(
		&publicDir,
		"publicDir",
		defaultPublicDir(),
		"the directory to serve static files",
	)

	var indexFallback bool
	app.RootCmd.PersistentFlags().BoolVar(
		&indexFallback,
		"indexFallback",
		true,
		"fallback the request to index.html on missing static path (eg. when pretty urls are used with SPA)",
	)

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		var filePath = os.DirFS(publicDir)
		var blah = apis.StaticDirectoryHandler(filePath, indexFallback)
		e.Router.GET("/*", blah)
		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/api/init-check", func(c echo.Context) error {
			var total int

			err := app.DB().
				Select("count(*)").
				From("_admins").
				Row(&total)
			if err != nil {
				return err
			}
			if total > 0 {
				return c.JSON(http.StatusOK, map[string]interface{}{
					"isSetup": true,
					"message": "Initial setup is complete",
				})
			}
			return c.JSON(http.StatusOK, map[string]interface{}{
				"isSetup": false,
				"message": "Initial setup needed",
			})
		})
		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

// the default pb_public dir location is relative to the executable
func defaultPublicDir() string {
	if strings.HasPrefix(os.Args[0], os.TempDir()) {
		return "./dist"
	}
	return filepath.Join(os.Args[0], "../dist")
}
