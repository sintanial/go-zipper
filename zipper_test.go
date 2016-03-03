package zipper

import (
	"testing"
	"archive/zip"
	"io/ioutil"
	"bytes"
	"os"
)

func TestFromZip(t *testing.T) {
}

var simpleName = "zipper/file/test.name"
var simpleValue = []byte("Hello world")

func TestAddBytes(t *testing.T) {
	zp := NewZipper()
	zp.AddBytes(simpleName, simpleValue)
	checkZipper(t, zp, simpleName, simpleValue)
}

func TestAddReader(t *testing.T) {
	zp := NewZipper()
	zp.AddReader(simpleName, bytes.NewBuffer(simpleValue))
	checkZipper(t, zp, simpleName, simpleValue)
}
func TestAddFile(t *testing.T) {
	fname := "testdata/simple.data"
	zp := NewZipper()
	err := ioutil.WriteFile(fname, simpleValue, os.ModePerm)
	if err != nil {
		t.Fatal("pack: " + err.Error())
	}
	defer os.Remove(fname)

	zp.AddFile(simpleName, fname)
	checkZipper(t, zp, simpleName, simpleValue)
}
func TestAddString(t *testing.T) {
	zp := NewZipper()
	zp.AddString(simpleName, string(simpleValue))
	checkZipper(t, zp, simpleName, simpleValue)
}

func TestAddZip(t *testing.T) {
	buffer := &bytes.Buffer{}

	zw := zip.NewWriter(buffer)
	f, err := zw.Create(simpleName)
	if err != nil {
		t.Fatal("pack: " + err.Error())
	}

	_, err = f.Write(simpleValue)
	if err != nil {
		t.Fatal("pack: " + err.Error())
	}
	zw.Close()

	zr, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(buffer.Len()))
	if err != nil {
		t.Fatal("pack: " + err.Error())
	}

	zp := NewZipper()
	zp.AddZip(zr.File[0])
	checkZipper(t, zp, simpleName, simpleValue)
}

func TestRemove(t *testing.T) {
	zp := NewZipper()
	zp.AddBytes(simpleName, simpleValue)
	zp.Remove(simpleName)
	rd := packZipper(t, zp, simpleName, simpleValue)
	if isZipContainsKey(rd.File, simpleName) {
		t.Fail()
	}

}
func TestRemoveByMask(t *testing.T) {
	zp := NewZipper()
	zp.AddBytes(simpleName, simpleValue)
	zp.RemoveByMask("zipper/file/*")
	rd := packZipper(t, zp, simpleName, simpleValue)
	if isZipContainsKey(rd.File, simpleName) {
		t.Fail()
	}
}

func TestPack(t *testing.T) {
	TestAddBytes(t)
}
func TestWriteTo(t *testing.T) {
	zp := NewZipper()
	zp.AddBytes(simpleName, simpleValue)

	buffer := &bytes.Buffer{}
	err := zp.WriteTo(buffer)
	if err != nil {
		t.Fatal("pack: " + err.Error())
	}

	rd, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(buffer.Len()))
	if err != nil {
		t.Fatal("pack: " + err.Error())
	}

	contains, err := isZipContainsValue(rd.File, simpleName, simpleValue)
	if err != nil {
		t.Fatal("pack: " + err.Error())
	}
	if !contains {
		t.Fail()
	}
}

func checkZipper(t *testing.T, zp *Zipper, name string, value []byte) {
	rd := packZipper(t, zp, name, value)
	contains, err := isZipContainsValue(rd.File, simpleName, simpleValue)
	if err != nil {
		t.Fatal("pack: " + err.Error())
	}
	if !contains {
		t.Fail()
	}
}

func packZipper(t *testing.T, zp *Zipper, name string, value []byte) *zip.Reader {
	buffer, err := zp.Pack()
	if err != nil {
		t.Fatal("pack: " + err.Error())
	}

	rd, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(buffer.Len()))
	if err != nil {
		t.Fatal("pack: " + err.Error())
	}

	return rd
}

func isZipContainsKey(files []*zip.File, name string) bool {
	for _, file := range files {
		if file.Name == name {
			return true
		}
	}

	return false
}

func isZipContainsValue(files []*zip.File, name string, value []byte) (bool, error) {
	for _, file := range files {
		if file.Name == name {
			rd, err := file.Open()
			if err != nil {
				return false, err
			}
			data, err := ioutil.ReadAll(rd)
			if err != nil {
				return false, err
			}

			return bytes.Equal(data, value), nil
		}
	}

	return false, nil
}