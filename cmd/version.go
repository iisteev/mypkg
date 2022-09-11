/*
Copyright © 2022 Isteevan Shetoo <isteevan.shetoo@is-info.fr>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"os"
	"runtime"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// BuildDate the build time
	BuildDate = "Do not know yet"

	// Version the version of the binary
	Version = "dev"
)

var versionTemplate = `Version:      {{.Version}}
Go version:   {{.GoVersion}}
Built:        {{.BuildTime}}
OS/Arch:      {{.Os}}/{{.Arch}}`

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Disply mypkg version",
	Long:  `Disply mypkg version and exit`,
	Run: func(cmd *cobra.Command, args []string) {
		tmplate, err := template.New("").Parse(versionTemplate)
		if err != nil {
			log.Fatal(err)
		}
		v := struct {
			Version   string
			GoVersion string
			BuildTime string
			Os        string
			Arch      string
		}{
			Version:   Version,
			GoVersion: runtime.Version(),
			BuildTime: BuildDate,
			Os:        runtime.GOOS,
			Arch:      runtime.GOARCH,
		}

		if err := tmplate.Execute(os.Stdout, v); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
