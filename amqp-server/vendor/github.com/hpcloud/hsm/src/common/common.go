package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
)

// AppVersion Build version of binary
var AppVersion string

// ListOfStrings list/array of strings
type ListOfStrings []string

//supporting implementation of String() for ListOfStrings
func (i *ListOfStrings) String() string {
	return fmt.Sprint(*i)
}

//Set supporting implementation of Set() for listOfStrings
func (i *ListOfStrings) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// CatalogConfiguration struct for the environment variables needed for the server
type CatalogConfiguration struct {
	Host string `env:"HOST" default:"0.0.0.0" long:"host" description:"HTTP listen server"`
	Port int    `env:"PORT" default:"8081" long:"port" description:"HTTP listen port"`

	TLSCertFile       string `env:"TLS_CERT_FILE" default:"/etc/secrets/tls-cert-file" long:"tls_cert_file" description:"Path to the TLS certificate file to be used"`
	TLSPrivateKeyFile string `env:"TLS_PRIVATE_KEY_FILE" default:"/etc/secrets/tls-private-key-file" long:"tls_private_key_file" description:"Path to the TLS private certificate key file to be used"`

	GeneratorCACertFile       string `env:"GENERATOR_CA_CERT" default:"/etc/secrets/generator-ca-cert" long:"generator_ca_cert" description:"Path to the parameter generation CA certificate file to be used"`
	GeneratorCAPrivateKeyFile string `env:"GENERATOR_CA_PRIVATE_KEY" default:"/etc/secrets/generator-ca-private-key" long:"generator_ca_private_key" description:"Path to the parameter generation private CA certificate key file to be used"`

	CatalogDir                 string        `env:"CATALOG_DIR" default:"CATALOG_CACHE" long:"catalog_dir" description:"Catalog root directory"`
	CatalogList                ListOfStrings `env:"CATALOG_LIST" default:"stackato;helion-service-manager/lkg/master/catalog-templates/HCP_v1:us-west-2;s3" long:"catalog_list" description:"CatalogList declared to hold the catalogs list"`
	DevMode                    string        `env:"DEV_MODE" long:"dev_mode" default:"true" description:"A boolean value to enable Developer Mode"`
	InstanceCreateRetryTimeout int           `env:"INSTANCE_CREATE_RETRY_TIMEOUT" default:"30" long:"instance_create_retry_timeout" description:"Timeout for instance create operation on UCP"`
	LogLevel                   string        `env:"LOG_LEVEL" default:"debug" long:"log_level" description:"Set log level (info,debug,error,fatal)"`
	SyncInterval               int           `env:"SYNC_INTERVAL" default:"300" long:"sync_interval" description:"Time interval to periodically refresh the catalog"`

	HCPEndpoint     string `env:"HCP_ENDPOINT" long:"hcp_endpoint" required:"true" description:"HCP Service Endpoint"`
	HCPCACertFile   string `env:"HCP_CA_CERT_FILE" default:"/etc/secrets/hcp-ca-cert" long:"hcp_ca_cert_file" description:"HCP CA Cert file location"`
	HCPDomainSuffix string `env:"HCP_DOMAIN_SUFFIX" long:"hcp_domain_suffix" default:"cluster.hcp" description:"The DNS domain name of the HCP"`
	HCPTimeout      int64  `env:"HCP_TIMEOUT" default:"30" long:"hcp_timeout" description:"Timeout for HCP API calls"`
	USBTimeout      int64  `env:"USB_TIMEOUT" default:"30" long:"usb_timeout" description:"Timeout for USB API calls"`

	UAADisabled                bool   `env:"UAA_DISABLED" long:"uaa_disabled" description:"Whether to disable UAA authentication"`
	HCPIdentityScheme          string `env:"HCP_IDENTITY_SCHEME" long:"hcp_identity_schema" description:"URL where UAA is running"`
	HCPIdentityHost            string `env:"HCP_IDENTITY_HOST" long:"hcp_identity_host" description:"URL where UAA is running"`
	HCPIdentityPort            int    `env:"HCP_IDENTITY_PORT" long:"hcp_identity_port" description:"URL where UAA is running"`
	HCPIdentityExternalHost    string `env:"HCP_IDENTITY_EXTERNAL_HOST" long:"hcp_identity_ext_host" description:"URL where UAA is running"`
	HCPIdentityExternalPort    int    `env:"HCP_IDENTITY_EXTERNAL_PORT" long:"hcp_identity_ext_port" description:"URL where UAA is running"`
	RefreshVerificationKeyTime int    `env:"REFRESH_VERIFICATION_KEY_TIME" long:"refresh_verification_key_time" default:"60" description:"Time interval to keep cached verification key before attempting to refetch the key"`

	FlightRecorderHost string `env:"HCP_FLIGHTRECORDER_HOST" long:"hcp_flightrecorder_host" description:"Flight Recorder Service Endpoint"`
	FlightRecorderPort int    `env:"HCP_FLIGHTRECORDER_PORT" long:"hcp_flightrecorder_port" description:"Flight Recorder Service Endpoint"`
	GitRetry           int    `env:"GIT_RETRY" long:"git_retry" default:"5" description:"Number of retries for git clone or git pull"`
	ReleaseChannel     string `env:"RELEASE_CHANNEL" long:"release_channel" default:"stable" description:"Release channel for catalog and artifacts."`
}

// FlightRecorderEndpoint concatenates the host and port for flight recorder
func (c *CatalogConfiguration) FlightRecorderEndpoint() string {
	return c.FlightRecorderHost + ":" + strconv.Itoa(c.FlightRecorderPort)
}

// UAAInternalEndpoint returns URL for UAA Internal Endpoint
// While running outside of HCP it will be  https://<HCP_NODE_IP>:<UAA_PORT>
// While running on HCP this will be https://ident-api.hcp.svc.cluster.hcp:443
func (c *CatalogConfiguration) UAAInternalEndpoint() string {
	return c.HCPIdentityScheme + "://" + c.HCPIdentityHost + ":" + strconv.Itoa(c.HCPIdentityPort)
}

// UAAExternalEndpoint returns URL for UAA External Endpoint
// While running outside of HCP it will be  https://<HCP_NODE_IP>:<UAA_PORT>
// While running on HCP this will be https://<ELB Load Balancer>:443
func (c *CatalogConfiguration) UAAExternalEndpoint() string {
	return c.HCPIdentityScheme + "://" + c.HCPIdentityExternalHost + ":" + strconv.Itoa(c.HCPIdentityExternalPort)
}

// PublicKey returns the publicKey if available otherwise pulls from disk
func (c *CatalogConfiguration) PublicKey() string {
	if publicKey == "" {
		filePath := ""
		switch c.ReleaseChannel {
		case "stable":
			filePath = "gpg/stable.key"
		case "development":
			filePath = "gpg/development.key"
		}
		Logger.Info("Using sign verification key : " + filePath)
		file, err := os.Open(filePath)
		defer file.Close()
		if err != nil {
			return ""
		}
		pubKey, err := ioutil.ReadAll(file)
		if err != nil {
			return ""
		}
		publicKey = string(pubKey)
	}

	return publicKey
}

// DNSMaxLength is the max length of a DNS name (or instance ID) to accept or generate
const DNSMaxLength = 63

// DNSRegexPattern is the regex to use when validating a DNS entry is ok
const DNSRegexPattern = "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"

var (

	// CatalogConfig variable of type CatalogConfiguration
	CatalogConfig CatalogConfiguration

	// Parser parses the value either from the env variable,
	// values passed in or the default values.
	parser = flags.NewParser(&CatalogConfig, flags.Default)

	// Logger would be used for logging
	Logger *logrus.Logger

	// Location to cache the public key after it is loaded from file
	publicKey string

	//config parameters which we dont want to log
	secretConfigs = []string{}
)

// ParseAndVerifyCli parses and verifies the command line parameters supplied
func ParseAndVerifyCli() {

	_, err := parser.Parse()
	if e, ok := err.(*flags.Error); ok {
		if e.Type == flags.ErrHelp {
			os.Exit(0) //exit without error in case of help
		} else {
			os.Exit(1) //exit with error for other cases
		}
	}
	//initialize logger as requested using flags
	Logger = NewLogger(strings.ToLower(CatalogConfig.LogLevel))

	Logger.Info("Environment has been set, as per parameters supplied")
	v := reflect.ValueOf(CatalogConfig)
	values := make([]interface{}, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		values[i] = v.Field(i).Interface()
		if IsElementInList(v.Type().Field(i).Name, secretConfigs) {
			continue
		}
		if v.Field(i).Interface() != "" {
			Logger.Infof("%s initialized to: %+v", v.Type().Field(i).Name, v.Field(i).Interface())
		} else {
			Logger.Infof("%s is not initialized.", v.Type().Field(i).Name)
		}
	}
}

// IsElementInList checks if the input element in present in the list
func IsElementInList(input interface{}, inputList interface{}) bool {
	list := reflect.ValueOf(inputList)
	for i := 0; i < list.Len(); i++ {
		if list.Index(i).Interface() == input {
			return true
		}
	}
	return false
}
