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
package mpkg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type PackageDesc struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Release     string   `yaml:"release"`
	Source      Source   `yaml:"source"`
	Licence     string   `yaml:"licence"`
	HomePage    string   `yaml:"homePage"`
	Summary     string   `yaml:"summary"`
	Description string   `yaml:"description"`
	Setup       []string `yaml:"setup"`
	Build       []string `yaml:"build"`
	Install     []string `yaml:"install"`
}

func (s *PackageDesc) GetFullName() string {
	return fmt.Sprintf("%s-%s-%s", s.Name, s.Version, s.Release)
}

func (s *PackageDesc) SetupStep(shell *Shell, dir string) error {
	steps, err := GetStepsFromMacros(s.Setup)
	if err != nil {
		return err
	}
	return shell.Exec(dir, steps)
}

func (s *PackageDesc) BuildStep(shell *Shell, dir string) error {
	steps, err := GetStepsFromMacros(s.Build)
	if err != nil {
		return err
	}
	return shell.Exec(dir, steps)
}

func (s *PackageDesc) InstallStep(shell *Shell, dir string) error {
	steps, err := GetStepsFromMacros(s.Install)
	if err != nil {
		return err
	}
	return shell.Exec(dir, steps)
}

func (s *PackageDesc) Archive(installDir, dest, prefix string) error {
	curdir, err := filepath.Abs("./")
	if err != nil {
		return fmt.Errorf("unable to get current directory: %w", err)
	}
	frootPath := filepath.Join(installDir, prefix)
	container, err := UnmarshalFilesXML(frootPath)
	if err != nil {
		return err
	}

	if err := os.Chdir(installDir); err != nil {
		return fmt.Errorf("could not enter dir %s: %w", installDir, err)
	}
	filenames := map[string]string{}

	for _, file := range container.Files {
		fullpath := filepath.Join(installDir, file.Path)
		filenames[fullpath] = file.Path
	}

	prefix, _ = strings.CutPrefix(prefix, "/")
	filesXMLPath := filepath.Join(prefix, "files.xml")

	fullpath := filepath.Join(installDir, filesXMLPath)
	filenames[fullpath] = filesXMLPath

	dest = filepath.Join(curdir, fmt.Sprintf("%s.%s.%s", dest, DefaultArchival, DefaultCompression))

	return ArchiveFiles(context.Background(), installDir, dest, filenames, DefaultArchival, DefaultCompression)
}
