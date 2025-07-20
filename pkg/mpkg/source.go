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
	"io"
	"net/http"
	"os"
)

// Source contains uri and sha256 of the tarball
type Source struct {
	URI          string `yaml:"uri"`
	Sha256       string `yaml:"sha256"`
	Decompressed bool   `yaml:"decompressed"`
}

// Download Download the tarball
func (s *Source) Download(filepath string) error {

	// Get the data
	resp, err := http.Get(s.URI)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func (s *Source) DownloadIfNoCache(filepath string) error {
	if IsExit(filepath) {
		err := s.Verify(filepath)
		if err == nil {
			return nil
		}
	}
	return s.Download(filepath)
}

// Verify verify the sha256 of a file, return error if mismatch
func (s *Source) Verify(filepath string) error {
	sum, err := GetHashString(filepath)
	if err != nil {
		return err
	}
	if sum != s.Sha256 {
		return fmt.Errorf("wrong sha256 of file %s; want %v, got %v", filepath, s.Sha256, sum)
	}
	return nil
}

// Unpack unpacks the given compressed file to destination
func (s *Source) Unpack(filepath string, dest string) error {
	return Unarchive(context.Background(), filepath, dest)
}
