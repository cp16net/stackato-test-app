package common

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/blang/semver"
)

// NameSeparator string constant separate catalog, vendor and service
const NameSeparator = "."

// GetCatalogURL builds appropriate catalog URL from http.Request object
func GetCatalogURL(r *http.Request) string {

	if r.URL.IsAbs() {
		return (r.URL.Scheme + "://" + r.URL.Host + r.URL.Path)
	}

	return ("https://" + r.Host + r.RequestURI)
}

// GetID returns the normalized ID
func GetID(args ...string) string {
	var normalizedID string
	for _, arg := range args {
		arg = strings.Trim(arg, " ")
		normalizedID = normalizedID + NameSeparator + strings.ToLower(strings.Replace(arg, " ", "-", -1))
	}
	return strings.Trim(normalizedID, NameSeparator)
}

// RetryOperation re-runs provided callback function with multiple attempts until either
// the Operation is successful or the timeout specified with timeoutSeconds has reached.
// It also sleeps for specified sleepSeconds, in between the retry/re-run.
func RetryOperation(timeoutSeconds int, sleepSeconds int, taskName string, callback func() error) (err error) {
	Logger.WithFields(logrus.Fields{"Task": taskName, "Timeout": timeoutSeconds}).Debug("RetryOperation")
	var (
		usedSeconds = 0
	)
	for usedSeconds < timeoutSeconds {
		// calling callback function
		err = callback()
		if err == nil {
			// if err is nil, that means function execution was successful,
			// we are done retrying
			return nil
		}

		if usedSeconds+sleepSeconds > timeoutSeconds {
			// with next sleep since we are exceeding provided timeoutSeconds
			// lets adjust the sleepSeconds to
			// seconds we have left between what we have used already and provided timeoutSeconds
			sleepSeconds = timeoutSeconds - usedSeconds
		}

		d := time.Duration(sleepSeconds * 1000 * 1000 * 1000)
		time.Sleep(d)
		usedSeconds = usedSeconds + sleepSeconds

		Logger.WithFields(logrus.Fields{"Task": taskName}).Debug("Retrying")
	}
	return fmt.Errorf("after %d seconds, last error: %s", timeoutSeconds, err)
}

// RetryExBackoffOperation re-runs provided callback function with multiple attempts until either
// the Operation is successful or the timeout specified with timeoutSeconds has reached.
// The sleep between two successive retry increases by 2 times
func RetryExBackoffOperation(timeoutSeconds int, taskName string, callback func() error) (err error) {
	Logger.WithFields(logrus.Fields{"Task": taskName, "Timeout": timeoutSeconds}).Debug("RetryExBackoffOperation")
	var (
		usedSeconds  = 0
		sleepSeconds = 1
	)
	for usedSeconds < timeoutSeconds {
		// calling callback function
		err = callback()
		if err == nil {
			// if err is nil, that means function execution was successful,
			// we are done retrying
			return nil
		}

		// sleep duration in Seconds
		sleepSeconds = sleepSeconds * 2
		if usedSeconds+sleepSeconds > timeoutSeconds {
			// with next sleep since we are exceeding provided
			// timeoutSeconds lets adjust the sleepSeconds to
			// seconds we have left between what we have used already and provided timeoutSeconds
			sleepSeconds = timeoutSeconds - usedSeconds
		}

		d := time.Duration(sleepSeconds * 1000 * 1000 * 1000)
		time.Sleep(d)
		usedSeconds = usedSeconds + sleepSeconds

		Logger.WithFields(logrus.Fields{"Task": taskName}).Debug("Retrying")
	}
	return fmt.Errorf("after %d seconds, last error: %s", timeoutSeconds, err)
}

//GetLatestSdlVersion returns the latest sdl version for a given product version
func GetLatestSdlVersion(semverVersions []semver.Version) string {

	if len(semverVersions) == 0 {
		return ""
	}

	semver.Sort(semverVersions)
	return semverVersions[len(semverVersions)-1].String()
}
