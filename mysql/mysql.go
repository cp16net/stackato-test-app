package mysql

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"

	"github.com/cp16net/hod-test-app/common"
	"github.com/cp16net/hod-test-app/mysql/models"
	"github.com/jinzhu/gorm"
	"github.com/julienschmidt/httprouter"
)

// Service Definition
type Service struct {
	Service []Mysql `json:"mysql"`
}

// Mysql Service Info
type Mysql struct {
	Creds Credentials `json:"credentials"`
}

// Credentials for HavenOnDemand
type Credentials struct {
	User     string `json:"user"`
	Username string `json:"username"`
	Database string `json:"database"`
	Host     string `json:"host"`
	Hostname string `json:"hostname"`
	Port     string `json:"port"`
	Password string `json:"password"`
}

var envVcapServices = `
{
	"mysql": [
		{
			"credentials": {
				"database": "d78289ac53d224a06beb99fe67775c876",
                "host": "mysql-int.mysql.svc",
                "hostname": "mysql-int.mysql.svc",
                "password": "ckm7DURqexO_hdvFCSqGLqmlojua-uBcY6JAXS0ogzY",
                "port": "3306",
                "user": "58780af232f16243",
                "username": "58780af232f16243"
			},
			"syslog_drain_url": null,
			"volume_mounts": [],
			"label": "mysql",
			"provider": null,
			"plan": "default",
			"name": "mysql",
			"tags": [
				"mysql"
			]
		}
	]
}`

var mysql Credentials

func setMysqlVcapServices() {
	vcap := os.Getenv("VCAP_SERVICES")
	if vcap == "" {
		// t := template.New("hello template")
		// t, _ = t.Parse(envVcapServices)
		// v := Vcap{HodKey: os.Getenv("HODKEY")}
		// var doc bytes.Buffer
		// t.Execute(&doc, v)
		// vcap = doc.String()
		vcap = envVcapServices
	}
	if err := json.Unmarshal([]byte(vcap), &mysql); err != nil {
		common.Logger.Error(err)
	}
	return
}

func dbConnection() *gorm.DB {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to connect database")
	}
	return db
}

func closeConnection(db *gorm.DB) {
	db.Close()
}

func init() {
	db := dbConnection()
	defer closeConnection(db)

	// Migrate the schema
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Email{})
}

func generateString(length int, characters string) (string, error) {
	b := make([]byte, length)

	max := big.NewInt(int64(len(characters)))

	for i := range b {
		var c byte
		rint, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", errors.New("Unable to generate a string. Error : " + err.Error())
		}
		c = characters[rint.Int64()]
		b[i] = c
	}
	return string(b), nil
}

const usercharacters = "abcdefghijklmnopqrstuvwxyz"
const passwordcharacters = `abcdefghijklmnopqrstuvwxyz1234567890-=[]\;',./~!@#$%&^*()_+{}|:\"<>?`

// CreateData generates a random user
func CreateData(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	db := dbConnection()
	defer closeConnection(db)
	username, err := generateString(10, usercharacters)
	if err != nil {
		common.Logger.Error("Error generating a username", err)
	}
	password, err := generateString(10, passwordcharacters)
	if err != nil {
		common.Logger.Error("Error generating a password", err)
	}
	db.Create(&models.User{
		Username: username,
		Password: password,
		Emails: []models.Email{
			{Email: username + "@gmail.com"},
		},
	})
}

// Test handler to get coordinate details from havenondemand
func Test(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	db := dbConnection()
	defer closeConnection(db)

	var users []models.User
	db.Find(&users)

	fmt.Fprintln(w, "this is going to be testing mysql")
}
