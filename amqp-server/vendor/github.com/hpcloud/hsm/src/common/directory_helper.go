package common

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// DirHelper interface abstracts functions for walking the catalog directory structure
type DirHelper interface {
	Walk(root string, walkFn filepath.WalkFunc) error
	DeleteEmptyDirectories(directoryDiskLocation string) error
	Stat(name string) (root os.FileInfo, err error)
}

// DirectoryHelper abstracts functions for walking the catalog directory structure
type DirectoryHelper struct {
	DirHelper
}

// Walk walks a directory
func (direc *DirectoryHelper) Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}

// DeleteEmptyDirectories removes directories that do not have services in them
func (direc *DirectoryHelper) DeleteEmptyDirectories(directoryDiskLocation string) error {
	var rootDir, files, versions []os.FileInfo
	var err error

	rootDir, err = ioutil.ReadDir(directoryDiskLocation)
	for _, directory := range rootDir {
		if directory.IsDir() {
			files, err = ioutil.ReadDir(directoryDiskLocation + "/" + directory.Name())
			if len(files) == 0 {
				os.Remove(directoryDiskLocation + "/" + directory.Name())
				direc.DeleteEmptyDirectories(directoryDiskLocation)
			}
			for _, secondLevelFile := range files {
				if secondLevelFile.IsDir() {
					versions, err = ioutil.ReadDir(directoryDiskLocation + "/" + directory.Name() + "/" + secondLevelFile.Name())
					if len(versions) == 0 {
						os.Remove(directoryDiskLocation + "/" + directory.Name() + "/" + secondLevelFile.Name())
						direc.DeleteEmptyDirectories(directoryDiskLocation)
					} else {
						direc.DeleteEmptyDirectories(directoryDiskLocation + "/" + directory.Name())
					}
				}
			}
		}
	}
	return err
}

// Stat returns a state of the file specified in name
func (direc *DirectoryHelper) Stat(name string) (root os.FileInfo, err error) {
	return os.Stat(name)
}
