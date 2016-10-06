package main

//go:generate go-bindata-assetfs static/... templates/...

import (
	"html/template"
	"net/http"
	"os"
	"strconv"

	"github.com/cp16net/hod-test-app/common"
	"github.com/cp16net/hod-test-app/hod"
	"github.com/jessevdk/go-flags"
	"github.com/julienschmidt/httprouter"
	"github.com/tylerb/graceful"
)

// Config for webappliation
type Config struct {
	Host string `env:"HOST" default:"0.0.0.0" long:"host" description:"HTTP listen server"`
	Port int    `env:"PORT" default:"8080" long:"port" description:"HTTP listen port"`
}

var (
	// AppConfig configuration for web application
	AppConfig Config
	parser    = flags.NewParser(&AppConfig, flags.Default)

	// Templates with functions available to them
	templates = template.New("").Funcs(templateMap)
)

// Parse all of the bindata templates
func init() {
	for _, path := range AssetNames() {
		bytes, err := Asset(path)
		if err != nil {
			common.Logger.Panicf("Unable to parse: path=%s, err=%s", path, err)
		}
		templates.New(path).Parse(string(bytes))
	}
	_, err := parser.Parse()
	if e, ok := err.(*flags.Error); ok {
		if e.Type == flags.ErrHelp {
			os.Exit(0) //exit without error in case of help
		} else {
			os.Exit(1) //exit with error for other cases
		}
	}
}

// Render a template given a model
func renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func hodIndex(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	renderTemplate(w, "templates/hod.html", nil)

}

// The server itself
func main() {
	common.Logger.Info("Starting up web application")
	// mux handler
	router := httprouter.New()

	// Routes for hod page and api
	router.GET("/hod", hodIndex)
	router.GET("/hodinfo/:lat/:lng", hod.Info)

	// Serve static assets via the "static" directory
	router.ServeFiles("/static/*filepath", assetFS())

	common.Logger.Info("Setup routes")
	// Serve this program forever
	port := strconv.Itoa(AppConfig.Port)
	host := AppConfig.Host
	httpServer := &graceful.Server{Server: new(http.Server)}
	httpServer.Addr = host + ":" + port
	httpServer.Handler = router
	common.Logger.Infof("listening at http://%s:%s", host, port)
	if err := httpServer.ListenAndServe(); err != nil {
		shutdown(err)
	}
}

// shutdown closes down the api server
func shutdown(err error) {
	common.Logger.Fatalln(err)
}
