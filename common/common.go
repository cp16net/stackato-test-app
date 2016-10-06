package common

import "strings"

// Logger for logging stuff
var Logger = NewLogger(strings.ToLower("debug"))
