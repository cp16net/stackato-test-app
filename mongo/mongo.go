package mongo

import (
	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/cp16net/hod-test-app/common"
	"github.com/cp16net/hod-test-app/rabbitmq"
	"gopkg.in/mgo.v2"
)

// GetLogs returns the list of logs in the db
func GetLogs() []rabbitmq.Log {
	appEnv, _ := cfenv.Current()
	mongosvc, err := appEnv.Services.WithName("cp16net-mongo")
	if err != nil {
		panic("failed to get the cp16net-mongo service details")
	}
	uri, ok := mongosvc.CredentialString("uri")
	if !ok {
		panic("failed to get the credential uri for mongo")
	}
	username, ok := mongosvc.CredentialString("username")
	if !ok {
		panic("failed to get the credential username for mongo")
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
	// common.Logger.Debug(username)
	// common.Logger.Debug(password)
	// common.Logger.Debug(host)
	// common.Logger.Debug(port)
	// common.Logger.Debug(dbname)
	generatedURI := "mongodb://" + username + ":" + password + "@" + host + ":" + port + "/" + dbname
	common.Logger.Debug(generatedURI)
	session, err := mgo.Dial(generatedURI)
	if err != nil {
		common.Logger.Error("failed to connect to mongo: ", err)
	}
	defer session.Close()
	// session.SetMode(mgo.Monotonic, true)
	c := session.DB(dbname).C("gologger")
	iter := c.Find(nil).Sort("-$natural").Limit(100).Iter()
	result := []rabbitmq.Log{}
	err = iter.All(&result)
	if err != nil {
		common.Logger.Fatal(err)
	}
	return result
}
