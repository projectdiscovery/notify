package runner

import (
	"github.com/projectdiscovery/gologger"
	updateutils "github.com/projectdiscovery/utils/update"
)

const banner = `
             __  _ ___    
  ___  ___  / /_(_) _/_ __
 / _ \/ _ \/ __/ / _/ // /
/_//_/\___/\__/_/_/ \_, /
                   /___/  
`

// version is the current version
const version = `1.0.6`

// showBanner is used to show the banner to the user
func showBanner() {
	gologger.Print().Msgf("%s\n", banner)
	gologger.Print().Msgf("\t\tprojectdiscovery.io\n\n")
}

// GetUpdateCallback returns a callback function that updates notify
func GetUpdateCallback() func() {
	return func() {
		showBanner()
		updateutils.GetUpdateToolCallback("notify", version)()
	}
}
