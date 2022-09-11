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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"mypkg/pkg/mpkg"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove an installed tarball.",
	Long: `Removes the files of the tarball based on files.xml present in the defined dbDir.
It will look in the db directory for a folder starting with provided name which reperesent
the tarball/package installed in dbDir.

Provide only one argument to this command. The name of the tarball is without version and release.

for example:
    mypkg remove htop`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("Non or more than one argument provide. Accepting ONLY one argument")
		}
		// Get the dbpath
		dbdir := getKeyFromConf("dbDir")
		prefix := getKeyFromConf("prefix")
		var packageName string
		// We look for in package name
		files, err := ioutil.ReadDir(dbdir)
		if err != nil {
			log.Fatal(err)
		}
		for _, folder := range files {
			if strings.HasPrefix(folder.Name(), args[0]) {
				packageName = folder.Name()
			}
		}
		if packageName == "" {
			log.Fatalf("Could not find an installed package with name %v\n", args[0])
		}
		all := strings.Split(packageName, "-")
		if len(all) < 3 {
			log.Fatalf("name not correct %v\n", packageName)
		}
		name := strings.Join(all[:len(all)-2], "")
		version := all[len(all)-2]
		release := all[len(all)-1]
		pkg := &mpkg.PackageDesc{
			Name:    name,
			Version: version,
			Release: release,
		}
		pkgPath := filepath.Join(dbdir, pkg.GetFullName())
		if mpkg.IsNotExist(pkgPath) {
			log.Fatalf("Could not find any %s\n", pkgPath)
		}
		// Unmarchall files.xml
		filesXML, err := mpkg.UnmarshalFilesXML(pkgPath)
		log.Infof("Deleting files of %v\n", pkg.GetFullName())
		if err != nil {
			log.Fatalf("Could not unmarchal files.xml %v\n", err)
		}
		for _, file := range filesXML.Files {
			fpath := filepath.Join("/", file.Path)
			if err := os.Remove(fpath); err != nil {
				log.Printf("Error deleting %v %v\n", fpath, err)
			}
		}
		// delete it in db
		if err := os.RemoveAll(pkgPath); err != nil {
			log.Fatalf("Could not delete package file in dbDir %v", err)
		}
		// Check for empty directory and delete it
		if err := mpkg.DeleteEmptyFolder(prefix); err != nil {
			log.Fatalf("Could not delete empty directories %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().StringVar(&packageName, "name", "", "the name of the package (required)")
	// removeCmd.MarkFlagRequired("name")
}
