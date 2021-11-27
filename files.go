package main

import (
	"io/fs"
	"path/filepath"
)

func getSongList(input string) ([]string, error) {
	result := make([]string, 0)
	addPath := func(fpath string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// skip hidden dir
		if dirEntry.IsDir() && dirEntry.Name()[:1] == "." {
			return filepath.SkipDir
		}
		if !dirEntry.IsDir() && // not a directory
			Contains(supportedFormats, filepath.Ext(fpath)) && // is supported format
			dirEntry.Name()[:1] != "." { // skip hidden files
			result = append(result, fpath)
		}
		return nil
	}
	err := filepath.WalkDir(input, addPath)

	return result, err

}

func Contains(arr []string, input string) bool {
	for _, v := range arr {
		if v == input {
			return true
		}
	}
	return false
}
