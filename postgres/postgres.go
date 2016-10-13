package postgres

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"math/big"
	"os"

	"github.com/cp16net/hod-test-app/common"
	"github.com/cp16net/hod-test-app/mysql/models"
	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/postgres" // needed for gorm
)

// Service Definition
type Service struct {
	Service []Postgres `json:"cp16net-postgres"`
}

// Postgres Service Info
type Postgres struct {
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
	"cp16net-mysql": [
		{
			"credentials": {
				"database": "appdb",
                "host": "localhost",
                "hostname": "localhost",
                "password": "password",
                "port": "3306",
                "user": "cp16net",
                "username": "cp16net"
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

var postgres Credentials

func setPostgresVcapServices() {
	vcap := os.Getenv("VCAP_SERVICES")
	if vcap == "" {
		vcap = envVcapServices
	}
	var svc Service
	if err := json.Unmarshal([]byte(vcap), &svc); err != nil {
		common.Logger.Error(err)
		panic("could not read vcap")
	}
	postgres = svc.Service[0].Creds
	return
}

func dbConnection() *gorm.DB {
	common.Logger.Debug("RUNNING IN CF MODE WITH POSTGRES")
	setPostgresVcapServices()
	// connectionString := postgres.User + ":" + postgres.Password
	// connectionString += "@tcp(" + postgres.Host + ":" + postgres.Port + ")"
	// connectionString += "/" + postgres.Database + "?charset=utf8&parseTime=True"
	connectionString := "host=" + postgres.Host + " user=" + postgres.User + " dbname=" + postgres.Database + " sslmode=disable password=" + postgres.Password + ""
	db, err := gorm.Open("postgres", connectionString)
	if err != nil {
		panic("failed to connect database")
	}
	db.LogMode(true)
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
}

func generateString(length int, characters string) (string, error) {
	b := make([]byte, length)
	max := big.NewInt(int64(len(characters)))
	for i := range b {
		var c byte
		rint, err := rand.Int(rand.Reader, max)
		if err != nil {
			common.Logger.Error(err)
			return "", errors.New("Unable to generate a string. Error : " + err.Error())
		}
		c = characters[rint.Int64()]
		b[i] = c
	}
	common.Logger.Debug("generated string: ", string(b))
	return string(b), nil
}

const usercharacters = "abcdefghijklmnopqrstuvwxyz"
const passwordcharacters = `abcdefghijklmnopqrstuvwxyz1234567890`

// GenerateUser generates a random user
func GenerateUser() models.User {
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
	var user = models.User{
		Username: username,
		Password: password,
		Email:    username + "@gmail.com",
	}
	common.Logger.Debug("Creating User: ", user)
	common.Logger.Debug("has user been created?: ", db.NewRecord(user))
	db.Create(&user)
	common.Logger.Debug("has user been created?: ", db.NewRecord(user))
	return user
}

// Users handler to get coordinate details from havenondemand
func Users() []models.User {

	db := dbConnection()
	defer closeConnection(db)

	var users []models.User
	db.Find(&users)
	common.Logger.Debug("found users: ", users)

	return users
}
