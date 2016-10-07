package main

//go:generate go-bindata-assetfs static/... templates/...

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/cp16net/hod-test-app/common"
	"github.com/cp16net/hod-test-app/hod"
	"github.com/cp16net/hod-test-app/mysql"
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

	// parse the flags
	_, err := parser.Parse()
	if e, ok := err.(*flags.Error); ok {
		if e.Type == flags.ErrHelp {
			os.Exit(0) //exit without error in case of help
		} else {
			os.Exit(1) //exit with error for other cases
		}
	}

	setGoogleVcapServices()
}

// Render a template given a model
func renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Service Definition
type Service struct {
	Service []Google `json:"cp16net-googleapi"`
}

// Google Service Info
type Google struct {
	Creds Credentials `json:"credentials"`
}

// Credentials for HavenOnDemand
type Credentials struct {
	PassthroughData string `json:"PASSTHROUGH_DATA"`
}

var envVcapServices = `
{
	"cp16net-googleapi": [
		{
			"credentials": {
				"PASSTHROUGH_DATA": "{\"google_api_key\":\"AIzaSyCdRGSQcby_ya2BY-5aViKW0o4pMcHJS-g\"}"
			}
		}
	]
}`

var creds Credentials

func setGoogleVcapServices() {
	vcap := os.Getenv("VCAP_SERVICES")
	if vcap == "" {
		vcap = envVcapServices
	}
	var svc Service
	if err := json.Unmarshal([]byte(vcap), &svc); err != nil {
		common.Logger.Error(err)
		panic("could not read vcap")
	}
	common.Logger.Debug(vcap)
	common.Logger.Debug(svc)
	common.Logger.Debug(svc.Service[0])
	common.Logger.Debug(svc.Service[0].Creds)
	creds = svc.Service[0].Creds
	return
}
func hodIndex(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var data map[string]string
	if err := json.Unmarshal([]byte(creds.PassthroughData), &data); err != nil {
		common.Logger.Error(err)
		panic("could not read vcap: " + creds.PassthroughData)
	}
	key := data["google_api_key"]
	renderTemplate(w, "templates/hod.html", key)
}

func envHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintln(w, strings.Join(os.Environ(), "\n"))
}

func mainHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	renderTemplate(w, "templates/main.html", nil)
}

func mysqlHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	data := mysql.Users()
	renderTemplate(w, "templates/mysql.html", data)
}

func mysqlCreateUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	mysql.GenerateUser()
	http.Redirect(w, r, "/mysql", 302)
}

// The server itself
func main() {
	common.Logger.Info("Starting up web application")
	// mux handler
	router := httprouter.New()
	router.GET("/", mainHandler)

	router.GET("/env", envHandler)

	// Routes for hod page and api
	router.GET("/hod", hodIndex)
	router.GET("/hodinfo/:lat/:lng", hod.Info)

	router.GET("/mysql", mysqlHandler)
	router.GET("/mysql/generate", mysqlCreateUserHandler)

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

	go func() {
		<-httpServer.StopChan()
	}()
}

// shutdown closes down the api server
func shutdown(err error) {
	common.Logger.Fatalln(err)
}
