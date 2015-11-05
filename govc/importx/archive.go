/*
Copyright (c) 2014 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package importx

import (
	"archive/tar"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/vmware/govmomi/ovf"
)

// ArchiveFlag doesn't register any flags;
// only encapsulates some common archive related functionality.
type ArchiveFlag struct {
	Archive
}

func (f *ArchiveFlag) Register(fs *flag.FlagSet) { return }

func (f *ArchiveFlag) Process() error { return nil }

func (f *ArchiveFlag) ReadOvf(fpath string) ([]byte, error) {
	r, _, err := f.Archive.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return ioutil.ReadAll(r)
}

func (f *ArchiveFlag) ReadEnvelope(fpath string) (*ovf.Envelope, error) {
	if fpath == "" {
		return nil, nil
	}

	r, _, err := f.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	e, err := ovf.Unmarshal(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ovf: %s", err.Error())
	}

	return e, nil
}

type Archive interface {
	Open(string) (io.ReadCloser, int64, error)
}

type TapeArchive struct {
	path string
}

type TapeArchiveEntry struct {
	io.Reader
	f *os.File
}

func (t *TapeArchiveEntry) Close() error {
	return t.f.Close()
}

func (t *TapeArchive) Open(name string) (io.ReadCloser, int64, error) {
	f, err := os.Open(t.path)
	if err != nil {
		return nil, 0, err
	}

	r := tar.NewReader(f)

	for {
		h, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, 0, err
		}

		matched, err := filepath.Match(name, path.Base(h.Name))
		if err != nil {
			return nil, 0, err
		}

		if matched {
			return &TapeArchiveEntry{r, f}, h.Size, nil
		}
	}

	_ = f.Close()

	return nil, 0, os.ErrNotExist
}

type FileArchive struct {
	path string
}

func (t *FileArchive) Open(name string) (io.ReadCloser, int64, error) {
	fpath := name
	if name != t.path {
		fpath = filepath.Join(filepath.Dir(t.path), name)
	}

	s, err := os.Stat(fpath)
	if err != nil {
		return nil, 0, err
	}

	f, err := os.Open(fpath)
	if err != nil {
		return nil, 0, err
	}

	return f, s.Size(), nil
}

type HTTPFileArchive struct {
	path   string
	once   sync.Once
	Client clientDoer
}

type clientDoer interface {
	Do(*http.Request) (*http.Response, error)
}

func (s *HTTPFileArchive) Open(name string) (contents io.ReadCloser, contentLength int64, err error) {
	var (
		req *http.Request
		res *http.Response
	)
	dir := os.TempDir()
	localFilename := fmt.Sprintf("%s/%s", dir, path.Base(name))
	s.once.Do(func() {
		if s.Client == nil {
			s.Client = new(http.Client)
		}

		if req, err = http.NewRequest("GET", name, nil); err == nil {

			if res, err = s.Client.Do(req); err == nil {
				out, _ := os.Create(localFilename)
				io.Copy(out, res.Body)
			}
		}
	})
	fileArchive := &FileArchive{path: localFilename}
	return fileArchive.Open(localFilename)
}
