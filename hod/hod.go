package hod

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/cp16net/hod-test-app/common"
	"github.com/julienschmidt/httprouter"
)

// Coordinate model
type Coordinate struct {
	Latitude string `json:"lat"`
	Longitue string `json:"lng"`
}

// Service Definition
type Service struct {
	HOD []HavenOnDemand `json:"hod10"`
}

// HavenOnDemand Service Info
type HavenOnDemand struct {
	Creds Credentials `json:"credentials"`
}

// Credentials for HavenOnDemand
type Credentials struct {
	APIKey string `json:"HOD_API_KEY"`
}

var envVcapServices = `
{
	"hod10": [
		{
			"credentials": {
				"HOD_API_KEY": "{{.HodKey}}"
			},
			"syslog_drain_url": null,
			"volume_mounts": [],
			"label": "hod10",
			"provider": null,
			"plan": "default",
			"name": "hod10",
			"tags": [
				"hod10"
			]
		}
	]
}`

// Vcap type to store VCAP_SERVICES
type Vcap struct {
	HodKey string
}

var hodServiceInfo Service

func setHodVcapServices() {
	vcap := os.Getenv("VCAP_SERVICES")
	if vcap == "" {
		t := template.New("hello template")
		t, _ = t.Parse(envVcapServices)
		v := Vcap{HodKey: os.Getenv("HODKEY")}
		var doc bytes.Buffer
		t.Execute(&doc, v)
		vcap = doc.String()
	}
	if err := json.Unmarshal([]byte(vcap), &hodServiceInfo); err != nil {
		common.Logger.Error(err)
	}
	return
}

// Info handler to get coordinate details from havenondemand
func Info(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	coord := Coordinate{
		Latitude: ps.ByName("lat"),
		Longitue: ps.ByName("lng"),
	}
	setHodVcapServices()
	hodAPIKey := hodServiceInfo.HOD[0].Creds.APIKey
	location := "&lat=" + coord.Latitude + "&lon=" + coord.Longitue + "&"
	hodurl := "https://api.havenondemand.com/1/api/sync/mapcoordinates/v1?apikey=" + hodAPIKey + location + "targets=country&targets=timezone&targets=zipcode_us"
	resp, err := http.Get(hodurl)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	var out bytes.Buffer
	err = json.Indent(&out, body, "", "\t")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, out.String())
}
