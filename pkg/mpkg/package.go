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
package mpkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v3"
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

func addToArchive(file File, installDir string, wr *archiver.TarXz) error {
	info, err := os.Stat(file.Path)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}
	absPath := fmt.Sprintf("%s/%s", installDir, file.Path)
	// open the file
	ofile, err := os.Open(absPath)
	if err != nil {
		return err
	}
	// write it to the archive
	err = wr.Write(archiver.File{
		FileInfo: archiver.FileInfo{
			FileInfo:   info,
			CustomName: file.Path,
		},
		ReadCloser: ofile,
	})
	ofile.Close()
	return err
}

func (s *PackageDesc) Archive(installDir, dest, prefix string) error {
	frootPath := filepath.Join(installDir, prefix)
	container, err := UnmarshalFilesXML(frootPath)
	if err != nil {
		return err
	}
	tarxz := archiver.NewTarXz()

	binary, err := os.Create(dest)
	if err != nil {
		return err
	}
	if err := tarxz.Create(binary); err != nil {
		return err
	}

	if err := os.Chdir(installDir); err != nil {
		return fmt.Errorf("could not enter dir %v %v", installDir, err)
	}
	for _, file := range container.Files {
		if err := addToArchive(file, installDir, tarxz); err != nil {
			return err
		}
	}

	if strings.HasPrefix(prefix, "/") {
		prefix = strings.TrimPrefix(prefix, "/")
	}
	filesXML := &File{
		Path: prefix + "/files.xml",
		Mode: "0644",
	}

	if err := addToArchive(*filesXML, installDir, tarxz); err != nil {
		return err
	}
	// Close the in-memory archive so that it writes trailing data
	if err := tarxz.Close(); err != nil {
		return err
	}

	if err = binary.Close(); err != nil {
		return err
	}
	return nil
}
