// Copyright 2015 CNI authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"os"

	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"
	log "github.com/sirupsen/logrus"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/version"

	kipam "kube-ipam/lib"
)

// Set the version number and release date of Kube-ipam.
const (
	Version     string = "v0.2.0"
	ReleaseDate string = "11/3/2021"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
	file, err := os.OpenFile("/var/log/kube-ipam.log", os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		log.SetOutput(file)
	} else {
		log.SetOutput(os.Stdout)
	}
	log.SetLevel(log.DebugLevel)
}

func main() {

	var outputconf string

	flag.StringVar(&outputconf, "outputconf", "", "Generate the configuration files required by different CNI plug-ins.(Use with \"macvlan | ipvlan | kube-router | bridge | flannel\")")
	versionFlag := flag.Bool("version", false, "Display software version information of kube-ipam.")
	helpFlag := flag.Bool("help", false, "Display usage help information of kube-ipam.")
	flag.Parse()

	switch {
	// Help information of Kube-ipam
	case *helpFlag:
		kipam.ShowHelp()
		// View software version details.
	case *versionFlag:
		kipam.ShowVersion(Version, ReleaseDate)
	case outputconf != "":
		switch {
		case outputconf == "macvlan":
			kipam.OutputCniConfig("macvlan")
		case outputconf == "ipvlan":
			kipam.OutputCniConfig("ipvlan")
		case outputconf == "kube-router":
			kipam.OutputCniConfig("kube-router")
		case outputconf == "bridge":
			kipam.OutputCniConfig("bridge")
		case outputconf == "flannel":
			kipam.OutputCniConfig("flannel")
		}

	default:
		skel.PluginMain(kipam.CmdAdd, kipam.CmdCheck, kipam.CmdDel, version.All, bv.BuildString("kube-ipam"))
	}

}
