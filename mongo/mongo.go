package mongo

import (
	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/cp16net/hod-test-app/common"
	"github.com/cp16net/hod-test-app/rabbitmq"
	"gopkg.in/mgo.v2"
)

// LogData struct for log data
type LogData struct {
	Logs  []rabbitmq.Log
	Count int
}

// GetLogs returns the list of logs in the db
func GetLogs() LogData {
	appEnv, _ := cfenv.Current()
	mongosvc, err := appEnv.Services.WithName("cp16net-mongo")
	if err != nil {
		panic("failed to get the cp16net-mongo service details")
	}
	uri, ok := mongosvc.CredentialString("uri")
	if !ok {
		panic("failed to get the credential uri for mongo")
	}

	// TODO: this is a hack for the mongodb uri
	// made PR to fix this in csm
	uri = uri[:len(uri)-1]

	dbname, ok := mongosvc.CredentialString("db")
	if !ok {
		panic("failed to get the credential name of db for mongo")
	}
	session, err := mgo.Dial(uri)
	if err != nil {
		common.Logger.Error("failed to connect to mongo: ", err)
	}
	defer session.Close()
	// session.SetMode(mgo.Monotonic, true)
	result := LogData{}
	c := session.DB(dbname).C("gologger")
	query := c.Find(nil)
	size, err := query.Count()
	if err != nil {
		common.Logger.Error("failed to get count of query from mongo: ", err)
	}
	result.Count = size
	iter := query.Sort("-$natural").Limit(100).Iter()
	err = iter.All(&result.Logs)
	if err != nil {
		common.Logger.Fatal(err)
	}
	return result
}
