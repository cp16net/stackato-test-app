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
	"github.com/cp16net/hod-test-app/mongo"
	"github.com/cp16net/hod-test-app/mysql"
	"github.com/cp16net/hod-test-app/rabbitmq"
	"github.com/cp16net/hod-test-app/redis"
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
		common.Logger.Error(err)
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

type redisData struct {
	Counter int64
	Data    map[string]string
}

func redisHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	rd := redisData{Data: make(map[string]string)}
	rd.Counter = redis.GetCount()
	keys := redis.ListKeys()
	for _, key := range keys {
		rd.Data[key] = redis.GetVal(key)
	}
	renderTemplate(w, "templates/redis.html", rd)
}

func redisIncrementHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	redis.Increment()
	http.Redirect(w, r, "/redis", 302)
}

func redisSetHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := r.PostFormValue("key")
	val := r.PostFormValue("value")
	redis.Set(key, val)
	http.Redirect(w, r, "/redis", 302)
}

// FibData data for output
type FibData struct {
	Input  int
	Output int
}

func rabbitmqFibHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fib := r.PostFormValue("fib")
	val, err := strconv.Atoi(fib)
	if err != nil {
		common.Logger.Error("Posted value is not an integer: ", fib)
	}
	out, err := rabbitmq.FibonacciRPC(val)
	if err != nil {
		common.Logger.Error("error calling fib on: ", err)
	}
	rd := FibData{Input: val, Output: out}
	renderTemplate(w, "templates/rabbitmq.html", rd)
}

func rabbitmqHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	renderTemplate(w, "templates/rabbitmq.html", FibData{Input: 1, Output: 0})
}

func rabbitmqLogHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logs := r.PostFormValue("logs")
	if logs == "" {
		renderTemplate(w, "templates/logs.html", []rabbitmq.Log{})
	}

	go func() {
		val, err := strconv.Atoi(logs)
		if err != nil {
			common.Logger.Error("Posted value is not an integer: ", logs)
		}
		rabbitmq.WriteLogs(val)
	}()

	result := mongo.GetLogs()
	renderTemplate(w, "templates/logs.html", result)
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

	router.GET("/redis", redisHandler)
	router.GET("/redis/increment", redisIncrementHandler)
	router.POST("/redis/set", redisSetHandler)

	// rabbitmq test route
	router.GET("/rabbitmq", rabbitmqHandler)
	router.POST("/rabbitmq/fib", rabbitmqFibHandler)
	router.GET("/logs", rabbitmqLogHandler)
	router.POST("/logs", rabbitmqLogHandler)

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
