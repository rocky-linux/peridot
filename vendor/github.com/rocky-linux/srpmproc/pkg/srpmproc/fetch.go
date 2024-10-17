package srpmproc

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/rocky-linux/srpmproc/pkg/blob"
	"github.com/rocky-linux/srpmproc/pkg/data"
)

func Fetch(logger io.Writer, cdnUrl string, dir string, fs billy.Filesystem, storage blob.Storage) error {
	pd := &data.ProcessData{
		Log: log.New(logger, "", log.LstdFlags),
	}

	metadataPath := ""
	ls, err := fs.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range ls {
		if strings.HasSuffix(f.Name(), ".metadata") {
			if metadataPath != "" {
				return errors.New("multiple metadata files found")
			}
			metadataPath = filepath.Join(dir, f.Name())
		}
	}
	if metadataPath == "" {
		return errors.New("no metadata file found")
	}

	metadataFile, err := fs.Open(metadataPath)
	if err != nil {
		return fmt.Errorf("could not open metadata file: %v", err)
	}

	fileBytes, err := io.ReadAll(metadataFile)
	if err != nil {
		return fmt.Errorf("could not read metadata file: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression: false,
		},
	}
	fileContent := strings.Split(string(fileBytes), "\n")
	for _, line := range fileContent {
		if strings.TrimSpace(line) == "" {
			continue
		}

		lineInfo := strings.SplitN(line, " ", 2)
		hash := strings.TrimSpace(lineInfo[0])
		path := strings.TrimSpace(lineInfo[1])

		url := fmt.Sprintf("%s/%s", cdnUrl, hash)
		if storage != nil {
			url = hash
		}
		pd.Log.Printf("downloading %s", url)

		var body []byte

		if storage != nil {
			body, err = storage.Read(hash)
			if err != nil {
				return fmt.Errorf("could not read blob: %v", err)
			}
		} else {
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				return fmt.Errorf("could not create new http request: %v", err)
			}
			req.Header.Set("Accept-Encoding", "*")

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("could not download dist-git file: %v", err)
			}

			body, err = io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("could not read the whole dist-git file: %v", err)
			}
			err = resp.Body.Close()
			if err != nil {
				return fmt.Errorf("could not close body handle: %v", err)
			}
		}

		hasher := pd.CompareHash(body, hash)
		if hasher == nil {
			return fmt.Errorf("checksum in metadata does not match dist-git file")
		}

		err = fs.MkdirAll(filepath.Join(dir, filepath.Dir(path)), 0o755)
		if err != nil {
			return fmt.Errorf("could not create all directories")
		}

		f, err := fs.Create(filepath.Join(dir, path))
		if err != nil {
			return fmt.Errorf("could not open file pointer: %v", err)
		}

		_, err = f.Write(body)
		if err != nil {
			return fmt.Errorf("could not copy dist-git file to in-tree: %v", err)
		}
		_ = f.Close()
	}

	return nil
}
