package common

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// RepositoryHelper interface
type RepositoryHelper interface {
	Clone(URL, downloadDir string) error
	Sync(URL, downloadDir string) (string, error)
}

// GitRepo interface
type GitRepo struct {
	URL    string
	Branch string
}

// S3Bucket interface
type S3Bucket struct {
	Bucket string
	Prefix string
	Region string
}

func (repo *GitRepo) parse(URL string) error {
	urlPieces := strings.Split(URL, "#")
	switch l := len(urlPieces); {
	case l > 2:
		return errors.New("Please provide URL information in the format git-url#(optional)branch-name")
	case l > 1:
		repo.URL = urlPieces[0]
		repo.Branch = urlPieces[1]
		return nil
	default:
		repo.URL = urlPieces[0]
		repo.Branch = "master"
		return nil
	}
}

// Clone clones a git repo using the git utility
func (repo *GitRepo) Clone(URL, downloadDir string) error {
	if err := repo.parse(URL); err != nil {
		return err
	}
	command := exec.Command("git", "clone", "--branch", repo.Branch, repo.URL, downloadDir)
	err := command.Run()
	if err != nil {
		err = errors.New("Cannot download " + URL + " at " + downloadDir)
	}
	return err
}

// Sync syncs a git repo using git -c
func (repo *GitRepo) Sync(URL, downloadDir string) (string, error) {
	if err := repo.parse(URL); err != nil {
		return "", err
	}
	command := exec.Command("git", "-C", downloadDir, "pull", "-r", "origin", repo.Branch)

	err := command.Run()
	if err != nil {
		err = errors.New("Cannot pull contents for repository " + downloadDir)
	}
	return "", err
}

func (repo *S3Bucket) parse(URL string) error {
	urlPieces := strings.Split(URL, ":")
	switch l := len(urlPieces); {
	case l != 2:
		return errors.New("Please provide S3 bucket information in the format bucket-name:region-name")
	default:
		bucketPieces := strings.SplitN(urlPieces[0], "/", 2)
		switch len(bucketPieces) {
		case 1:
			repo.Bucket = urlPieces[0]
		default:
			repo.Bucket = bucketPieces[0]
			repo.Prefix = strings.TrimSuffix(bucketPieces[1], "/")
		}
		repo.Region = urlPieces[1]
		return nil
	}
}

// Clone downloads an entire S3 bucket
func (repo *S3Bucket) Clone(URL, downloadDir string) error {
	if err := repo.parse(URL); err != nil {
		return err
	}
	session := session.New(&aws.Config{
		Region:      &repo.Region,
		Credentials: credentials.AnonymousCredentials})
	params := &s3.ListObjectsInput{
		Bucket: &repo.Bucket,
		Prefix: &repo.Prefix,
	}

	//Grab list of objects on S3 bucket
	resp, err := s3.New(session).ListObjects(params)

	//Download the objects in s3 bucket
	for _, object := range resp.Contents {
		objName := *object.Key
		if objName[len(objName)-1:] == "/" {
			continue
		}
		objectRemotePath := *object.Key
		if repo.Prefix != "" {
			objectRemotePath = strings.TrimPrefix(*object.Key, repo.Prefix+"/")
		}

		if _, err := os.Stat(filepath.Dir(downloadDir + "/" + objectRemotePath)); os.IsNotExist(err) {
			os.MkdirAll(downloadDir+"/"+filepath.Dir(objectRemotePath), 0700)
		}

		bucketFile, err := os.Create(downloadDir + "/" + objectRemotePath)
		if err != nil {
			return err
		}
		defer bucketFile.Close()

		downloader := s3manager.NewDownloader(session)
		_, err = downloader.Download(bucketFile,
			&s3.GetObjectInput{
				Bucket: params.Bucket,
				Key:    object.Key,
			})
		if err != nil {
			return err
		}

	}
	return err
}

// Sync syncs an entire S3 bucket looking for changes
func (repo *S3Bucket) Sync(URL, downloadDir string) (string, error) {
	var dirHelper DirectoryHelper
	var warn string

	if err := repo.parse(URL); err != nil {
		return "", err
	}

	_, err := dirHelper.Stat(downloadDir)
	if os.IsNotExist(err) {
		return warn, errors.New(downloadDir + " not found.")
	}

	//Create file-map for files existing at bucketDiskLocation
	fileList := []string{}
	err = dirHelper.Walk(downloadDir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			fileList = append(fileList, path)
		}
		return err
	})

	session := session.New(&aws.Config{
		Region:      &repo.Region,
		Credentials: credentials.AnonymousCredentials})
	params := &s3.ListObjectsInput{
		Bucket: &repo.Bucket,
		Prefix: &repo.Prefix,
	}

	// Grab list of objects in bucket
	resp, err := s3.New(session).ListObjects(params)
	if err != nil {
		return warn, err
	}

	//Track that files in bucketDiskLocation are still part of original bucket
	for _, file := range fileList {
		fileNotFound := true
		for _, bucketObject := range resp.Contents {
			objectRemotePath := *bucketObject.Key
			if repo.Prefix != "" {
				objectRemotePath = strings.TrimPrefix(*bucketObject.Key, repo.Prefix+"/")
			}

			if file == downloadDir+"/"+objectRemotePath {
				fileNotFound = false
			}
		}
		if fileNotFound {
			os.Remove(file)
		}
	}

	//Clean empty directories
	dirHelper.DeleteEmptyDirectories(downloadDir)

	// Download existing content of bucket, if newer than existing available
	for _, object := range resp.Contents {
		objName := *object.Key
		if objName[len(objName)-1:] == "/" {
			continue
		}
		objectRemotePath := *object.Key
		if repo.Prefix != "" {
			objectRemotePath = strings.TrimPrefix(*object.Key, repo.Prefix+"/")
		}

		if _, err := dirHelper.Stat(filepath.Dir(downloadDir + "/" + objectRemotePath)); os.IsNotExist(err) {
			err = os.MkdirAll(downloadDir+"/"+filepath.Dir(objectRemotePath), 0700)
			if err != nil {
				warn = warn + ": " + err.Error()
				continue
			}
		}

		file, err := dirHelper.Stat(downloadDir + "/" + objectRemotePath)
		if err != nil && !os.IsNotExist(err) {
			warn = warn + ": " + err.Error()
			continue
		}

		if err == nil && (*object.LastModified).Sub(file.ModTime()) <= 0 {
			continue
		}

		bucketFile, err := os.Create(downloadDir + "/" + objectRemotePath)
		if err != nil && !os.IsNotExist(err) {
			warn = warn + ": " + err.Error()
			continue
		}
		defer bucketFile.Close()

		downloader := s3manager.NewDownloader(session)
		_, err = downloader.Download(bucketFile,
			&s3.GetObjectInput{
				Bucket: params.Bucket,
				Key:    object.Key,
			})
		if err != nil {
			warn = warn + ": " + err.Error()
			continue
		}
	}
	return warn, nil
}
