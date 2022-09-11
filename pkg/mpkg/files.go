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
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// File is the description of file within the archive.
// The archive represent compiled tarball
type File struct {
	XMLName xml.Name `xml:"File"`
	Path    string   `xml:"Path"`
	Type    string   `xml:"Type"`
	UID     int      `xml:"Uid"`
	GID     int      `xml:"Gid"`
	Mode    string   `xml:"Mode"`
	Hash    string   `xml:"Hash"`
}

// Set is the list of file inside the compiled tarball (package)
type Set struct {
	XMLName xml.Name `xml:"Files"`
	Files   []File   `xml:"File"`
}

// FileTypes find the type of a file based on its path
var FileTypes = map[string]string{
	"PREFIX/lib/pkgconfig":   "data",
	"PREFIX/lib64/pkgconfig": "data",
	"PREFIX/lib32/pkgconfig": "data",
	"PREFIX/libexec":         "executable",
	"PREFIX/lib":             "library",
	"PREFIX/share/info":      "info",
	"PREFIX/share/man":       "man",
	"PREFIX/share/doc":       "doc",
	"PREFIX/share/help":      "doc",
	"PREFIX/share/gtk-doc":   "doc",
	"PREFIX/share/locale":    "localedata",
	"PREFIX/include":         "header",
	"PREFIX/bin":             "executable",
	"/bin":                   "executable",
	"PREFIX/sbin":            "executable",
	"/sbin":                  "executable",
	"/etc":                   "config",
}

// PkgConfigDirs the valid path of pkgconfig files
var PkgConfigDirs = []string{
	"/lib/pkgconfig",
	"/lib64/pkgconfig",
	"/lib32/pkgconfig",
}

// IsPkgConfig check if the file is a package config
func IsPkgConfig(fpath string) bool {
	for _, prefix := range PkgConfigDirs {
		fpath := filepath.Join("/", fpath)
		if strings.HasPrefix(fpath, prefix) {
			return true
		}
	}
	return false
}

// NewPackageFile
func NewPackageFile(path, hash, mode, ftype string) *File {
	return &File{
		Path: path,
		Type: ftype,
		UID:  os.Geteuid(),
		GID:  os.Getegid(),
		Mode: mode,
		Hash: hash,
	}
}

func (s *Set) VerifyIntegrity() error {
	for _, file := range s.Files {
		hash, err := GetHashString(file.Path)
		if err != nil {
			return err
		}
		if hash != file.Hash {
			return fmt.Errorf("hash mismatch, want %v, got %v", file.Hash, hash)
		}
	}
	return nil
}

func CreatePackageXMLFile(rootPath, prefix string) (*Set, error) {
	container := &Set{}
	xmlFiles := container.Files
	srootPath := rootPath
	sprefix := prefix
	srootPath = strings.TrimSuffix(srootPath, "/")
	sprefix = strings.TrimSuffix(sprefix, "/")

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			strippedPath := strings.ReplaceAll(path, srootPath, "")
			basicPath := strings.ReplaceAll(strippedPath, sprefix, "")
			if IsPkgConfig(basicPath) {
				if err := UpdateLineInFile(path, srootPath, ""); err != nil {
					return err
				}
			}
			hash, err := GetHashString(path)
			if err != nil {
				return fmt.Errorf("could not get the hash; %v", err)
			}
			info, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("could not get the file; %v", err)
			}
			mode := fmt.Sprintf("%04o", info.Mode().Perm())
			// Update package config
			strippedPath = strings.TrimPrefix(strippedPath, "/")
			filePackage := NewPackageFile(strippedPath, hash, mode, "data")
			xmlFiles = append(xmlFiles, *filePackage)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	container.Files = xmlFiles
	return container, nil
}

func WritePackageXMLFile(rootPath, prefix string) error {
	frootPath := filepath.Join(rootPath, prefix)
	if err := CreateDirIfNotExist(frootPath); err != nil {
		return err
	}
	container, err := CreatePackageXMLFile(rootPath, prefix)
	if err != nil {
		return err
	}
	output, err := xml.MarshalIndent(container, "", "    ")
	if err != nil {
		return fmt.Errorf("could not marchal xml: %v", err)
	}
	fpath := filepath.Join(frootPath, "/files.xml")
	return ioutil.WriteFile(fpath, output, os.ModePerm)
}

func UnmarshalFilesXML(rootPath string) (*Set, error) {
	fpath := filepath.Join(rootPath, "/files.xml")
	fileContent, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	var pkgFiles Set
	if err := xml.Unmarshal(fileContent, &pkgFiles); err != nil {
		return nil, err
	}
	return &pkgFiles, nil
}
