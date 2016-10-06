package main

//go:generate go-bindata-assetfs static/... templates/...

import (
	"html/template"
	"log"
	"net/http"

	"github.com/jmcfarlane/golang-templates-example/hod"
	"github.com/julienschmidt/httprouter"
)

// Templates with functions available to them
var templates = template.New("").Funcs(templateMap)

// Parse all of the bindata templates
func init() {
	for _, path := range AssetNames() {
		bytes, err := Asset(path)
		if err != nil {
			log.Panicf("Unable to parse: path=%s, err=%s", path, err)
		}
		templates.New(path).Parse(string(bytes))
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
	// mux handler
	router := httprouter.New()

	// Routes for hod page and api
	router.GET("/hod", hodIndex)
	router.GET("/hodinfo/:lat/:lng", hod.Info)

	// Serve static assets via the "static" directory
	router.ServeFiles("/static/*filepath", assetFS())

	// Serve this program forever
	log.Fatal(http.ListenAndServe(":8888", router))
}
