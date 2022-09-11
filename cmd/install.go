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
	"os"

	"path/filepath"
	"strings"

	"github.com/iisteev/mypkg/pkg/mpkg"
	"github.com/mholt/archiver/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install compiled tarball",
	Long:  `This is used to install tarball in the system`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("Non or more than one argument provide. Accepting ONLY one argument")
		}
		tarball := args[0]
		// Check if the tarball exist
		if mpkg.IsNotExist(tarball) {
			log.Fatalf("Could not found tarball %v\n", tarball)
		}
		// Get the dbpath
		dbdir := getKeyFromConf("dbDir")
		// Get files.xml
		tarxz := mpkg.GeFileBaseName(tarball)
		fileBaseName := mpkg.GeFileBaseName(tarxz)
		all := strings.Split(fileBaseName, "-")
		if len(all) < 3 {
			log.Fatalf("name not correct %v\n", fileBaseName)
		}
		name := strings.Join(all[:len(all)-2], "")
		version := all[len(all)-2]
		release := all[len(all)-1]
		pkg := &mpkg.PackageDesc{
			Name:    name,
			Version: version,
			Release: release,
		}
		prefixDir := getKeyFromConf("prefix")
		pkgPath := filepath.Join(dbdir, pkg.GetFullName())
		// Verify each file hash against its hash
		log.Printf("Installing %v\n", pkg.GetFullName())
		if err := os.MkdirAll(pkgPath, os.ModePerm); err != nil {
			log.Fatal(err)
		}
		if err := archiver.Unarchive(tarball, "/"); err != nil {
			log.Fatal(err)
		}
		if err := os.Rename(filepath.Join(prefixDir, "/files.xml"), filepath.Join(pkgPath, "/files.xml")); err != nil {
			log.Fatal(err)
		}
		//Unmarchall files.xml
		filesXML, err := mpkg.UnmarshalFilesXML(pkgPath)
		if err != nil {
			log.Fatalf("Could not unmarchal files.xml %v\n", err)
		}
		log.Println("Verifying integrity")
		for _, file := range filesXML.Files {
			fpath := filepath.Join("/", file.Path)
			hash, err := mpkg.GetHashString(fpath)
			hash = strings.TrimSpace(hash)
			fhash := strings.TrimSpace(file.Hash)
			if err != nil {
				log.Fatalf("Error getting hash %v\n", err)
			}
			if strings.Compare(hash, fhash) != 0 {
				log.Printf("Hash not correct, deleting file %v\n", file.Path)
				if err := os.Remove(fpath); err != nil {
					log.Fatalf("Could not remove file %v\n", err)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().StringVar(&tarball, "file", "", "the path to the tarball (required)")
	// installCmd.MarkFlagRequired("file")
}
