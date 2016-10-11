package redis

import (
	"encoding/json"
	"os"

	"github.com/cp16net/hod-test-app/common"
	"gopkg.in/redis.v4"
)

// Service Definition
type Service struct {
	Service []Redis `json:"cp16net-redis"`
}

// Redis Service Info
type Redis struct {
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
	"cp16net-redis": [
		{
			"credentials": {
				"host": "localhost",
                "hostname": "localhost",
                "password": "",
                "port": "6379",
                "uri": "redis://:@localhost:6379/"
			}
		}
	]
}`

var creds Credentials

func setMysqlVcapServices() {
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

func dbConnection() *redis.Client {
	common.Logger.Debug("Building connection to redis")
	setMysqlVcapServices()
	client := redis.NewClient(&redis.Options{
		Addr:     creds.Host + ":" + creds.Port,
		Password: creds.Password, // no password set
		DB:       0,              // use default DB
	})
	pong, err := client.Ping().Result()
	if err != nil || pong != "PONG" {
		common.Logger.Error(pong, err)
		panic(err)
	}
	common.Logger.Debug("Connected to redis")
	return client
}

func closeConnection(db *redis.Client) {
	db.Close()
}

// Increment does just that
func Increment() {
	client := dbConnection()
	defer closeConnection(client)
	if err := client.Incr("counter").Err(); err != nil {
		common.Logger.Error(err)
		panic(err)
	}
}

// GetCount gets the current counter
func GetCount() int64 {
	client := dbConnection()
	defer closeConnection(client)
	n, _ := client.Get("counter").Int64()
	return n
}

// ListKeys gets the list of all keys in db
func ListKeys() []string {
	client := dbConnection()
	defer closeConnection(client)
	n := client.Keys("*")
	return n.Val()
}

// GetVal gets the value of the key out of the db
func GetVal(key string) string {
	client := dbConnection()
	defer closeConnection(client)
	n := client.Get(key)
	return n.Val()
}
