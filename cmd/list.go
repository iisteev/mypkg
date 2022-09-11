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
	"fmt"
	"io/ioutil"

	"os"
	"strings"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed packages",
	Long: `Simply show the installed packages.
Reads simply the dbDir for present folders.  Each folder is an installed package.`,

	Run: func(cmd *cobra.Command, args []string) {
		// Get the build directory
		dbDir := getKeyFromConf("dbDir")
		files, err := ioutil.ReadDir(dbDir)
		if err != nil {
			log.Fatal(err)
		}
		const padding = 3
		w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.TabIndent)
		fmt.Fprintf(w, "Name\tVersion\tRelease\t\n")
		fmt.Fprintf(w, "----\t-------\t-------\t\n")
		for _, f := range files {
			all := strings.Split(f.Name(), "-")
			if len(all) < 3 {
				log.Fatalf("name not correct %v\n", f.Name())
			}
			name := strings.Join(all[:len(all)-2], "")
			version := all[len(all)-2]
			release := all[len(all)-1]
			fmt.Fprintf(w, "%s\t%s\t%s\t\n", name, version, release)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
