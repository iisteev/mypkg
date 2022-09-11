/*
Copyright © 2021 Isteevan Shetoo <isteevan.shetoo@is-info.fr>

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
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/iisteev/mypkg/pkg/mpkg"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var packageDesc *mpkg.PackageDesc

// readFileDefinition reads in config file and ENV variables if set.
func readFileDefinition(fileDefinition *string) {
	if fileDefinition == nil {
		log.Fatal("No file provided description provided")
	}

	vsd = viper.New()

	// Use config file from the flag.
	vsd.SetConfigFile(*fileDefinition)
	vsd.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := vsd.ReadInConfig(); err != nil {
		log.Fatalf("Could not read file %v, %v\n", fileDefinition, err)
	}
	log.Infof("Using package description file: %v\n", vsd.ConfigFileUsed())
	// Use config file from the flag.
	if err := vsd.Unmarshal(&packageDesc); err != nil {
		log.Fatalf("unable to decode into struct, %v\n", err)
	}
}

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a tarball",
	Long: `This is used to build tarball and create another compiled tarball.
An example of the yaml file:
---
name     : htop
version  : 3.0.5
release  : 1
homePage : https://htop.dev/
licence  : GPL-2.0-or-later
source   :
  uri: https://github.com/htop-dev/htop/archive/3.0.5.tar.gz
  sha256: 4c2629bd50895bd24082ba2f81f8c972348aa2298cc6edc6a21a7fa18b73990c
setup    :
  - sed -i .bak.sh s/glibtoolize/llibtoolize/g ./autogen.sh
  - ./autogen.sh
  - $configure
build    :
  - $make
install  :
  - $make_install
---
A macro starts with $, defined macros are:
$configure     : ./configure ${CONF_OPTS}
$make          : make -j${NBJOBS-1} ${MAKE_OPTS}
$make_install  : make install DESTDIR=${INSTALL_DIR-${prefix}} ${MAKE_INSTALL_OPTS}
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("Non or more than one argument provide. Accepting ONLY one argument")
		}
		// We read the config
		readFileDefinition(&args[0])
		tarballDir := getKeyFromConf("buildDir")
		// Check if dbdir exists, if not, create it
		if err := mpkg.CreateDirIfNotExist(tarballDir); err != nil {
			log.Fatalln(err)
		}
		installDBDir := getKeyFromConf("installDBDir")
		// Check if installDBDir exists, if not, create it
		if err := mpkg.CreateDirIfNotExist(installDBDir); err != nil {
			log.Fatalln(err)
		}
		// Prepare tarball
		// Download the tarball
		base := path.Base(packageDesc.Source.URI)
		temptarball := filepath.Join("/tmp", base)
		if err := packageDesc.Source.Download(temptarball); err != nil {
			log.Fatalf("Could not download tarball; %v\n", err)
		}
		log.Infof("File downloaded in %s\n", temptarball)
		// Verify the tarball
		if err := packageDesc.Source.Verify(temptarball); err != nil {
			log.Fatalf("Could not verify tarball; %v\n", err)
		}
		log.Info("Integrity OK")
		// unpack the tarball
		if err := packageDesc.Source.Unpack(temptarball, tarballDir); err != nil {
			log.Fatalf("Could not unpack tarball; %v\n", err)
		}
		log.Infof("Tarball unpacked in %v\n", tarballDir)
		// look for the unpacked tarball folder in destination directory
		folders, err := ioutil.ReadDir(tarballDir)
		if err != nil {
			log.Fatal(err)
		}
		if len(folders) != 1 {
			log.Fatalf("We chould find only one directory in %v\n", tarballDir)
		}
		candidateFile := folders[0]
		if !candidateFile.IsDir() {
			log.Fatalf("%v is not a directory\n", candidateFile.Name())
		}
		unpackedDir := filepath.Join(tarballDir, candidateFile.Name())
		// get prefix and installation directory
		prefixDir := getKeyFromConf("prefix")
		installDir := getKeyFromConf("installDir")
		command := mpkg.NewShell()
		command.AddArgs("PREFIX=" + prefixDir)
		command.AddArgs("BUILD_DIR=" + tarballDir)
		command.AddArgs("INSTALL_DIR=" + installDir)
		command.AddArgs("PKG_NAME=" + packageDesc.GetFullName())
		command.AddArgs("PKG_BUILD_DIR=" + unpackedDir)
		environment := vcfg.GetStringSlice("environment")
		for _, value := range environment {
			command.AddArgs(value)
		}

		if err := packageDesc.SetupStep(command, unpackedDir); err != nil {
			log.Fatal(err)
		}
		// build
		if err := packageDesc.BuildStep(command, unpackedDir); err != nil {
			log.Fatal(err)
		}
		// Install
		if err := packageDesc.InstallStep(command, unpackedDir); err != nil {
			log.Fatal(err)
		}
		// Prepare package file
		pkgFullName := packageDesc.GetFullName()
		pkgPath := filepath.Join(installDBDir, pkgFullName)
		// Create the xml files metadata
		if err := mpkg.CreateDirIfNotExist(pkgPath); err != nil {
			log.Fatalln(err)
		}
		// write files.xml
		if err := mpkg.WritePackageXMLFile(installDir, prefixDir); err != nil {
			log.Fatal(err)
		}
		// Archive it
		log.Println("Packaging...")
		pkgFullName = fmt.Sprintf("%s.tar.xz", pkgFullName)
		if err := packageDesc.Archive(installDir, pkgFullName, prefixDir); err != nil {
			log.Fatal(err)
		}
		log.Println("Cleanup...")
		defer os.RemoveAll(installDir)
		defer os.RemoveAll(tarballDir)
		defer os.RemoveAll(installDBDir)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// buildCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	//buildCmd.Flags().StringVar(&fileDefinition, "file", "", "The file of the tarball to build")
}
