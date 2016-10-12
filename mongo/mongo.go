package mongo

import (
	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/cp16net/hod-test-app/common"
	"github.com/cp16net/hod-test-app/rabbitmq"
	"gopkg.in/mgo.v2"
)

// GetLogs returns the list of logs in the db
func GetLogs() []rabbitmq.Log {
	uri, ok := mongosvc.CredentialString("uri")
	if !ok {
		panic("failed to get the credential uri for mongo")
	}
	user, ok := mongosvc.CredentialString("user")
	if !ok {
		panic("failed to get the credential user for mongo")
	}
	password, ok := mongosvc.CredentialString("password")
	if !ok {
		panic("failed to get the credential password for mongo")
	}
	host, ok := mongosvc.CredentialString("host")
	if !ok {
		panic("failed to get the credential host for mongo")
	}
	port, ok := mongosvc.CredentialString("port")
	if !ok {
		panic("failed to get the credential port for mongo")
	}
	dbname, ok := mongosvc.CredentialString("db")
	if !ok {
		panic("failed to get the credential name of db for mongo")
	}
	// mongo connection uri should be in the form of:
	// [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
	common.Logger.Debug(uri)
	common.Logger.Debug(user)
	common.Logger.Debug(password)
	common.Logger.Debug(host)
	common.Logger.Debug(port)
	common.Logger.Debug(dbname)
	generatedURI := "mongodb://" + user + ":" + password + "@" + host + ":" + port
	common.Logger.Debug(generatedURI)
	session, err := mgo.Dial(generatedURI)
	if err != nil {
		common.Logger.Error("failed to connect to mongo: ", err)
	}
	defer session.Close()
	// session.SetMode(mgo.Monotonic, true)
	c := session.DB("logs").C("gologger")
	result := []rabbitmq.Log{}
	err = c.Find(nil).All(&result)
	if err != nil {
		common.Logger.Fatal(err)
	}
	return result
}

var (
	// mongouri    string
	// mongodbname string
	mongosvc *cfenv.Service
)

func init() {
	appEnv, _ := cfenv.Current()
	svc, err := appEnv.Services.WithName("cp16net-mongo")
	if err != nil {
		panic("failed to get the cp16net-mongo service details")
	}
	mongosvc = svc
	// uri, ok := svc.CredentialString("uri")
	// if !ok {
	// 	panic("failed to get the credential uri for mongo")
	// }
	// mongouri = uri
	// dbname, ok := svc.CredentialString("name")
	// if !ok {
	// 	panic("failed to get the credential uri for mongo")
	// }
	// mongodbname = dbname
}
