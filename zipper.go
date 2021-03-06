package zipper

import (
	"io"
	"archive/zip"
	"path"
	"bytes"
	"os"
	"github.com/go-errors/errors"
	"strings"
)

type Zipper struct {
	files map[string]interface{}
}

// create empty zipper
func NewZipper() *Zipper {
	return &Zipper{
		files: make(map[string]interface{}),
	}
}

// create zipper from zip file
func FromZip(f zip.ReadCloser) *Zipper {
	zipper := NewZipper()
	for _, f := range f.File {
		zipper.AddZip(f)
	}

	return zipper
}

// get data reader from zip by name
func (self *Zipper) Reader(name string) (io.Reader, error) {
	value, ok := self.files[name]
	if !ok {
		return nil, errors.New("zipper: name " + name + " not found")
	}

	switch v := value.(type) {
	case []byte:
		return bytes.NewBuffer(v), nil
	case *zip.File:
		return v.Open()
	case io.Reader:
		return v, nil
	case string:
		return bytes.NewBufferString(v), nil
	}

	return nil, nil
}

func (self *Zipper) AddBytes(name string, b []byte) *Zipper {
	self.files[name] = b
	return self
}

func (self *Zipper) AddReader(name string, r io.Reader) *Zipper {
	self.files[name] = r
	return self
}

func (self *Zipper) AddFile(name string, dist string) *Zipper {
	self.files[name] = dist
	return self
}

func (self *Zipper) AddString(name string, s string) *Zipper {
	self.files[name] = []byte(s)
	return self
}

func (self *Zipper) AddZip(f *zip.File) *Zipper {
	self.files[f.Name] = f
	return self
}

func (self *Zipper) Remove(name string) *Zipper {
	delete(self.files, name)
	return self
}

func (self *Zipper) RemoveByPath(path string) *Zipper {
	for name, _ := range self.files {
		if strings.Index(name, path) == 0 {
			delete(self.files, name)
		}
	}
	return nil
}

// remove data from zip by shell mask
func (self *Zipper) RemoveByMask(mask string) error {
	for name, _ := range self.files {
		ok, err := path.Match(mask, name)
		if err != nil {
			return err
		}

		if ok {
			delete(self.files, name)
		}
	}
	return nil
}

func AddFile(zw *zip.Writer, name string, file string) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	if err != nil {
		return err
	}

	return nil
}

func AddReader(zw *zip.Writer, name string, r io.Reader) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}

	return nil
}

func AddZip(zw *zip.Writer, name string, f *zip.File) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	w, err := zw.CreateHeader(&f.FileHeader)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, rc)
	if err != nil {
		return err
	}

	return nil
}

func AddBytes(zw *zip.Writer, name string, b []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}

	_, err = w.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func AddString(zw *zip.Writer, name string, s string) error {
	return AddBytes(zw, name, []byte(s))
}

// packs data to zip
func WriteTo(zp *Zipper, w io.Writer) error {
	zw := zip.NewWriter(w)
	defer zw.Close()
	return Concat(zw, zp)
}

func Concat(zw *zip.Writer, zp *Zipper) error {
	for name, value := range zp.files {
		var err error
		switch v := value.(type) {
		case []byte:
			err = AddBytes(zw, name, v)
		case *zip.File:
			err = AddZip(zw, name, v)
		case io.Reader:
			err = AddReader(zw, name, v)
		case string:
			err = AddFile(zw, name, v)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// write data to writer (for example to zip file)
func (self *Zipper) WriteTo(w io.Writer) error {
	return WriteTo(self, w)
}
