package env

import "os"

// Defaults for environment variables
var defaultPort string = "8080"

// Exported environment variables
var Port string = ""

// init initializes the environment variables to either passed values or to defaults.
func init() {
	if Port = os.Getenv("PORT"); Port == "" {
		Port = defaultPort
	}
}
