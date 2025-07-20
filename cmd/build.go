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
			log.Fatal("Non or more than one argument provided. Accepting ONLY one argument")
		}

		// Read application file definition
		readFileDefinition(&args[0])
		buildDir := getKeyFromConf("buildDir")
		// Check if dbdir exists, if not, create it
		if err := mpkg.CreateDirIfNotExist(buildDir); err != nil {
			log.Fatalln(err)
		}
		installDBDir := getKeyFromConf("installDBDir")
		// Check if installDBDir exists, if not, create it
		if err := mpkg.CreateDirIfNotExist(installDBDir); err != nil {
			log.Fatalln(err)
		}
		baseFileName := path.Base(packageDesc.Source.URI)

		var destinationFileName string
		// Prepare tarball
		if packageDesc.Source.Decompressed {
			destinationFileName = filepath.Join(buildDir, baseFileName)
		} else {
			destinationFileName = filepath.Join("/tmp", baseFileName)
		}
		// Download the tarball
		if err := packageDesc.Source.DownloadIfNoCache(destinationFileName); err != nil {
			log.Fatalf("Could not download file; %v\n", err)
		}
		log.Infof("File downloaded in %s\n", destinationFileName)
		// Verify the tarball
		if err := packageDesc.Source.Verify(destinationFileName); err != nil {
			log.Fatalf("Could not verify file; %v\n", err)
		}
		log.Info("Integrity OK")
		packageBuildDir := buildDir
		if !packageDesc.Source.Decompressed {
			// unpack the tarball
			if err := packageDesc.Source.Unpack(destinationFileName, buildDir); err != nil {
				log.Fatalf("Could not unpack tarball; %v\n", err)
			}
			log.Infof("Tarball unpacked in %v\n", buildDir)
			// look for the unpacked tarball folder in destination directory
			folders, err := ioutil.ReadDir(buildDir)
			if err != nil {
				log.Fatal(err)
			}
			if len(folders) != 1 {
				log.Warnf("We should find only one directory in %v! Anyway using it as a package source\n", buildDir)
			} else {
				candidateDir := folders[0]
				if !candidateDir.IsDir() {
					log.Fatalf("%v is not a directory\n", candidateDir.Name())
				}
				packageBuildDir = filepath.Join(buildDir, candidateDir.Name())
			}
		}

		// get prefix and installation directory
		prefixDir := getKeyFromConf("prefix")
		installDir := getKeyFromConf("installDir")
		fullInstallDir := filepath.Join(installDir, prefixDir)
		command := mpkg.NewShell()
		command.AddArgs("PREFIX=" + prefixDir)
		command.AddArgs("BUILD_DIR=" + buildDir)
		command.AddArgs("INSTALL_DIR=" + installDir)
		command.AddArgs("FULL_INSTALL_DIR=" + fullInstallDir)
		command.AddArgs("PKG_NAME=" + packageDesc.GetFullName())
		command.AddArgs("PKG_BUILD_DIR=" + packageBuildDir)
		environment := vcfg.GetStringSlice("environment")
		for _, value := range environment {
			command.AddArgs(value)
		}

		if err := packageDesc.SetupStep(command, packageBuildDir); err != nil {
			log.Fatal(err)
		}
		// build
		if err := packageDesc.BuildStep(command, packageBuildDir); err != nil {
			log.Fatal(err)
		}
		// Install
		if err := packageDesc.InstallStep(command, packageBuildDir); err != nil {
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
		if err := packageDesc.Archive(installDir, pkgFullName, prefixDir); err != nil {
			log.Fatal(err)
		}
		log.Println("Cleanup...")
		defer os.RemoveAll(installDir)
		defer os.RemoveAll(buildDir)
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
