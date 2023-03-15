package runner

import (
	"github.com/projectdiscovery/gologger"
	updateutils "github.com/projectdiscovery/utils/update"
)

const banner = `
             __  _ ___    
  ___  ___  / /_(_) _/_ __
 / _ \/ _ \/ __/ / _/ // /
/_//_/\___/\__/_/_/ \_, / v1.0.4
                   /___/  
`

// version is the current version
const version = `1.0.4`

// showBanner is used to show the banner to the user
func showBanner() {
	gologger.Print().Msgf("%s\n", banner)
	gologger.Print().Msgf("\t\tprojectdiscovery.io\n\n")

	gologger.Print().Msgf("Use with caution. You are responsible for your actions\n")
	gologger.Print().Msgf("Developers assume no liability and are not responsible for any misuse or damage.\n")
}

// GetUpdateCallback returns a callback function that updates notify
func GetUpdateCallback() func() {
	return func() {
		showBanner()
		updateutils.GetUpdateToolCallback("notify", version)()
	}
}
