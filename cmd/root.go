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
	"io"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var vcfg *viper.Viper
var vsd *viper.Viper

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mypkg [COMMAND] [COMMAND|ARGUMENTS]",
	Short: "A very simple application manager",
	Long:  `Compile tarballs and install it based on provided configuration. Simple build description based on yaml file`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mypkg.yaml)")

	vcfg = viper.New()
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.AddCommand(completionCmd)
}

func getKeyFromConf(key string) string {
	skey := vcfg.GetString(key)
	if skey == "" {
		log.Fatalf("could not find %v in config\n", key)
	}
	return skey
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		vcfg.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".mypkg" (without extension).
		vcfg.AddConfigPath(home)
		vcfg.SetConfigName(".mypkg")
	}

	vcfg.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := vcfg.ReadInConfig(); err == nil {
		log.Debugf("Using config file: %v\n", vcfg.ConfigFileUsed())
	}
}

// Completion; thanks to rancher/k3d
var completionFunctions = map[string]func(io.Writer) error{
	"bash": rootCmd.GenBashCompletion,
	"zsh": func(writer io.Writer) error {
		if err := rootCmd.GenZshCompletion(writer); err != nil {
			return err
		}

		fmt.Fprintf(writer, "\n# source completion file\ncompdef _mypkg mypkg\n")

		return nil
	},
}

// NewCmdCompletion creates a new completion command
// create new cobra command
var completionCmd = &cobra.Command{
	Use:   "completion SHELL",
	Short: "Generate completion scripts for [bash, zsh]",
	Long:  `Generate completion scripts for [bash, zsh]`,
	Args:  cobra.ExactArgs(1), // TODO: NewCmdCompletion: add support for 0 args = auto detection
	Run: func(cmd *cobra.Command, args []string) {
		if completionFunc, ok := completionFunctions[args[0]]; ok {
			if err := completionFunc(os.Stdout); err != nil {
				log.Fatalf("Failed to generate completion script for shell '%s'", args[0])
			}
			return
		}
		log.Fatalf("Shell '%s' not supported for completion", args[0])
	},
}
