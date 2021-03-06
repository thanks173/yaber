package yaber

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// ErrNoPaths is returned from Generator.Generate() whenever the user has failed to
// provide at least a single file path to assets to embedd.
var ErrNoPaths = errors.New("no file paths to assets")

// AssetFile is the final, generated product from a AssetGenerator.
type AssetFile struct {
	Path string
	Body []byte
}

// AssetGenerator is the main object used for generating new files with embedded
// assets and tests.
type AssetGenerator struct {
	// Package sets the package name for the newly generated files.
	Package string

	// OutputPrefix is the path prefix to append to the generated files.
	OutputPrefix string

	// StripPath will strip this prefix from the embedded asset file paths.
	StripPath string

	// If true, PublicFuncs let's you publicly export some functions for
	// accessing your embedded assets, from outside the package.
	PublicFuncs bool
}

// NewGenerator is a shortcut function that will try to guess and use good
// default values for a new AssetGenerator.
func NewGenerator(pkg, output, strip string, publicFuncs bool) (*AssetGenerator, error) {
	if len(pkg) < 1 {
		var e error
		// Default to use the output (or the current) dir as the pkg name
		pkg, e = getPackageName(filepath.Dir(output))
		if e != nil {
			return nil, e
		}
	}

	if len(output) < 1 {
		output = "assets"
	}

	g := &AssetGenerator{
		Package:      pkg,
		OutputPrefix: output,
		StripPath:    strip,
		PublicFuncs:  publicFuncs,
	}
	return g, nil
}

// Generate attempts to read the provided asset files, compress them and
// then embedd them in a new Go file, along with a basic test file.
func (g *AssetGenerator) Generate(paths []string) ([]*AssetFile, error) {
	if len(paths) < 1 {
		return nil, ErrNoPaths
	}

	files := make(map[string][]byte)
	for _, p := range paths {
		f, e := embedAsset(p, g.StripPath)
		if e != nil {
			return nil, e
		}
		for k, v := range f {
			files[k] = v
		}
	}

	data := map[string]interface{}{
		"version": VERSION,
		"package": g.Package,
		"command": executedCommand(),
		"files":   files,
	}
	if g.PublicFuncs {
		data["assetFunc"] = "Asset"
		data["setRawFunc"] = "SetRawAssets"
	} else {
		data["assetFunc"] = "asset"
		data["setRawFunc"] = "setRawAssets"
	}

	// Generate the main file with embedded files.
	mainBody, e := runTemplate(tmplMain, data)
	if e != nil {
		return nil, e
	}
	main := &AssetFile{
		Path: g.OutputPrefix + ".go",
		Body: mainBody,
	}

	// Generate the test file.
	var first string
	for k := range files {
		first = k
		break
	}
	data["firstPath"] = first
	data["firstBody"] = files[first]
	data["dirs"] = paths

	testBody, e := runTemplate(tmplTest, data)
	if e != nil {
		return nil, e
	}
	test := &AssetFile{
		Path: g.OutputPrefix + "_test.go",
		Body: testBody,
	}

	return []*AssetFile{main, test}, nil
}

// Recursively reads all regular files in path, into memory as gzipped data.
// Returns a map where the keys are file paths and the values are the gzip byte data.
func embedAsset(path string, stripPath string) (map[string][]byte, error) {
	list := make(map[string][]byte)
	dirs := []string{path}

	for len(dirs) > 0 {
		d := dirs[0]
		dirs = dirs[1:]
		files, e := ioutil.ReadDir(d)
		if e != nil {
			return nil, e
		}

		for _, f := range files {
			fpath := filepath.Join(d, f.Name())
			tmpPath := strings.TrimPrefix(fpath, stripPath)

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
			if len(fbody) < 1 {
				list[tmpPath] = []byte{}
				continue
			}

			buf := new(bytes.Buffer)
			gw := gzip.NewWriter(buf)
			defer gw.Close()

			if _, e = gw.Write(fbody); e != nil {
				return nil, e
			}
			gw.Flush()
			if gw.Close() != nil {
				return nil, e
			}

			list[tmpPath] = buf.Bytes()
		}
	}
	return list, nil
}
