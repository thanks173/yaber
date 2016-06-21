//go:generate yaber -out assets/assets templates/

// Code generated by yaber v0.3 (https://github.com/lmas/yaber)
// DO NOT EDIT.

package assets

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var _useRawAssets bool = false

func SetRawAssets(b bool) {
	_useRawAssets = b
}

func Asset(path string) ([]byte, error) {
	if _useRawAssets {
		return GetRaw(path)
	}
	return GetEmbedded(path)
}

func AssetDir(dir string) (map[string][]byte, error) {
	if _useRawAssets {
		return GetRawDir(dir)
	}
	return GetEmbeddedDir(dir)
}

func GetRaw(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func GetRawDir(dir string) (map[string][]byte, error) {
	list := make(map[string][]byte)
	dirs := []string{dir}

	for len(dirs) > 0 {
		d := dirs[0]
		dirs = dirs[1:]
		files, e := ioutil.ReadDir(d)
		if e != nil {
			return nil, e
		}

		for _, f := range files {
			fpath := filepath.Join(d, f.Name())

			if f.IsDir() {
				dirs = append(dirs, fpath)
				continue
			}
			if !f.Mode().IsRegular() {
				continue
			}

			fbody, e := ioutil.ReadFile(fpath)
			if e != nil {
				return nil, e
			}
			list[fpath] = fbody
		}
	}
	return list, nil
}

func GetEmbedded(path string) ([]byte, error) {
	body, ok := _rawAssets[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return decompress(body)
}

func GetEmbeddedDir(dir string) (map[string][]byte, error) {
	var e error
	files := make(map[string][]byte)
	for path, body := range _rawAssets {
		if strings.HasPrefix(path, dir) {
			files[path], e = decompress(body)
			if e != nil {
				return nil, e
			}
		}
	}
	return files, nil
}

func decompress(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(data)
	gr, e := gzip.NewReader(buf)
	if e != nil {
		if e == io.EOF {
			return []byte{}, nil
		}
		return nil, e
	}
	defer gr.Close()
	return ioutil.ReadAll(gr)
}

var _rawAssets = map[string][]byte{

	"templates/aaa.txt": []byte("\x1f\x8b\b\x00\x00\tn\x88\x00\xff\n\xc9\xc8,VH\xcb\xccIU\x00\xd2Y\xa5\xc5%\n\x89\n%\xa9\xc5%z\\\x00\x00\x00\x00\xff\xff\x01\x00\x00\xff\xff\x9d\xc5\x12$\x1a\x00\x00\x00"),

	"templates/empty_file": []byte(""),

	"templates/hello": []byte("\x1f\x8b\b\x00\x00\tn\x88\x00\xff\xf2H\xcd\xc9\xc9W(\xcf/\xcaIQ\xe4\x02\x00\x00\x00\xff\xff\x01\x00\x00\xff\xffA\u4a72\r\x00\x00\x00"),
}
