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
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var Macros = map[string]string{
	"$configure":    "./configure ${CONF_OPTS}",
	"$make":         "make -j${NBJOBS-1} ${MAKE_OPTS}",
	"$make_install": "make install DESTDIR=${INSTALL_DIR-${prefix}} ${MAKE_INSTALL_OPTS}",
}

func GetCommand(command string) string {
	for macro, cmd := range Macros {
		if strings.Compare(command, macro) == 0 {
			return cmd
		}
	}
	return ""
}

func GetHash(filepath string) (hash.Hash, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return nil, err
	}
	return hash, nil
}

func GetHashString(filepath string) (string, error) {
	hash, err := GetHash(filepath)
	if err != nil {
		return "", nil
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func GetFileType(path string) string {
	for prefix, typ := range FileTypes {
		if strings.HasPrefix(path, prefix) {
			return typ
		}
	}
	return "data"
}

func IsNotExist(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}

func CreateDirIfNotExist(dir string) error {
	if IsNotExist(dir) {
		// create the path
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("could not create directory %v; error: %v", dir, err)
		}
	}
	return nil
}

func GeFileBaseName(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

// todo check file why is no readying
func UpdateLineInFile(fpath, oldLine, newLine string) error {
	ofile, err := os.Open(fpath)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(ofile)

	scanner.Split(bufio.ScanLines)
	var text []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.ContainsAny(scanner.Text(), oldLine) {
			line = strings.ReplaceAll(line, oldLine, newLine)
		}
		text = append(text, line)
	}
	if err := ofile.Close(); err != nil {
		return err
	}
	output := strings.Join(text, "\n")
	return ioutil.WriteFile(fpath, []byte(output), os.ModePerm)
}

func GetStepsFromMacros(steps []string) ([]string, error) {
	var formatedSteps []string
	for _, step := range steps {
		if strings.HasPrefix(step, "$") {
			step = GetCommand(step)
			if step == "" {
				return nil, fmt.Errorf("could not get command if macro %v", step)
			}
		}
		formatedSteps = append(formatedSteps, step+";")
	}
	return formatedSteps, nil
}

func DeleteEmptyFolder(folder string) error {
	info, err := os.Stat(folder)
	if !info.IsDir() {
		return err
	}
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return err
	}
	if len(files) > 0 {
		for _, file := range files {
			fpath := filepath.Join(folder, file.Name())
			DeleteEmptyFolder(fpath)
		}
		// re-evaluate files; after deleting subfolder
		// we may have parent folder empty now
		files, err = ioutil.ReadDir(folder)
		if err != nil {
			return err
		}
	}
	if len(files) == 0 {
		if err := os.Remove(folder); err != nil {
			return err
		}
	}
	return nil
}
