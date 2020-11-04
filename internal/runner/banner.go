package runner

import (
	"github.com/projectdiscovery/gologger"
)

const banner = `
             __  _ ___    
  ___  ___  / /_(_) _/_ __
 / _ \/ _ \/ __/ / _/ // /
/_//_/\___/\__/_/_/ \_, / 0.0.1 
                   /___/  
`

// Version is the current version
const Version = `0.0.1`

// showBanner is used to show the banner to the user
func showBanner() {
	gologger.Printf("%s\n", banner)
	gologger.Printf("\t\tprojectdiscovery.io\n\n")

	gologger.Labelf("Use with caution. You are responsible for your actions\n")
	gologger.Labelf("Developers assume no liability and are not responsible for any misuse or damage.\n")
}
