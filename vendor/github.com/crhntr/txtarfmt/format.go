package txtarfmt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"path/filepath"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/txtar"
)

type Configuration struct {
	SkipGo, SkipJSON, SkipGoMod bool
}

func Archive(archive *txtar.Archive, config Configuration) error {
	seen := make(map[string]struct{})
	for i, file := range archive.Files {
		if _, ok := seen[file.Name]; ok {
			return fmt.Errorf("duplicate archive file %s", file.Name)
		}
		fmtFile, err := File(file, config)
		if err != nil {
			return err
		}
		archive.Files[i] = fmtFile
		seen[fmtFile.Name] = struct{}{}
	}
	return nil
}

func File(file txtar.File, config Configuration) (txtar.File, error) {
	if !config.SkipGo && filepath.Ext(file.Name) == ".go" {
		out, err := format.Source(file.Data)
		if err != nil {
			return file, err
		}
		file.Data = out
	} else if !config.SkipJSON && filepath.Ext(file.Name) == ".json" {
		var buf bytes.Buffer
		if err := json.Indent(&buf, file.Data, "", "  "); err != nil {
			return file, err
		}
		file.Data = buf.Bytes()
	} else if !config.SkipGoMod && filepath.Base(file.Name) == "go.mod" {
		modFile, err := modfile.Parse(file.Name, file.Data, nil)
		if err != nil {
			return file, err
		}
		buf, err := modFile.Format()
		if err != nil {
			return file, err
		}
		file.Data = buf
	}
	return file, nil
}
