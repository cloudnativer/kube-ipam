package kipam

import (
	"fmt"
)

// Displays the detailed version information of the kube-install.
func ShowVersion(Version string, ReleaseDate string) {
	fmt.Println("  Version: " + Version + "\n  Release Date: " + ReleaseDate)
}
