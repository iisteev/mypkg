/*
Copyright © 2022 Isteevan Shetoo <isteevan.shetoo@is-info.fr>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package cmd

import (
	"os"
	"path"
	"path/filepath"
	"text/template"

	"github.com/iisteev/mypkg/pkg/mpkg"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var fetchTemplate = `---
name     : {{.Name}}
version  : {{.Version}}
release  : {{.Release}}
source   :
  uri: {{.Uri}}
  sha256: {{.SHA256}}
setup    :
  - $configure
build    :
  - $make
install  :
  - $make_install
`

var uri string
var sha256 string
var version string
var release string

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "create a description file with provided url",
	Long: `Creates a pre defined yaml description file.

if the flag --sha256 is not provided then the cli will try to fetch from the given url and calculate the its 256 sum.
Other parametres should be completed before build.

example:
    mypkg fetch htop
	mypkg fetch htop \
		--uri https://bintray.com/htop/source/download_file?file_path=htop-3.0.0.tar.gz \
		--sha256 4c2629bd50895bd24082ba2f81f8c972348aa2298cc6edc6a21a7fa18b73990c \
		--version 3.0.0 \
		--release 1`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("Non or more than one argument provide. Accepting ONLY one argument")
		}
		if sha256 == "" && uri != "" {
			pkg := &mpkg.PackageDesc{}
			pkg.Source.URI = uri
			base := path.Base(uri)
			temptarball := filepath.Join("/tmp", base)
			if err := pkg.Source.Download(temptarball); err != nil {
				log.Fatalf("Could not download from %v %v\n", uri, err)
			}
			hash, err := mpkg.GetHashString(temptarball)
			if err != nil {
				log.Fatalf("Could not get the hash of %v\n", temptarball)
			}
			sha256 = hash
		}
		tmplate, err := template.New("").Parse(fetchTemplate)
		if err != nil {
			log.Fatal(err)
		}
		v := struct {
			Name    string
			Version string
			Release string
			URI     string
			SHA256  string
		}{
			Name:    args[0],
			Version: version,
			Release: release,
			URI:     uri,
			SHA256:  sha256,
		}

		if err := tmplate.Execute(os.Stdout, v); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fetchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	fetchCmd.Flags().StringVar(&uri, "uri", "", "The uri of the tarball to fetch")
	fetchCmd.Flags().StringVar(&sha256, "sha256", "", "The sha256 of the tarball")
	fetchCmd.Flags().StringVar(&version, "version", "", "The version of the tarball")
	fetchCmd.Flags().StringVar(&release, "release", "", "The release of the tarball")
}
