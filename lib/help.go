package kipam

import (
	"fmt"
)

// Displays the detailed help information of the kube-install.
func ShowHelp() {
	fmt.Println(`Usage of ./kube-ipam:
    -help
          Display usage help information of kube-ipam.
    -outputconf string
          Generate the configuration files required by different CNI plug-ins.(Use with "macvlan | ipvlan | kube-router | bridge | calico")
    -version
          Display software version information of kube-ipam.
    `)
}
