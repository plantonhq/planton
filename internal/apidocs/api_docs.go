package apidocs

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/version"
	"github.com/pseudomuto/protoc-gen-doc"
	"io"
	"net/http"
)

const downloadUrlFormatString = "https://github.com/plantonhq/planton/releases/download/%s/docs.json"
const latestReleaseDownloadUrl = "https://github.com/plantonhq/planton/releases/latest/download/docs.json"

// GetApiDocsJson downloads the docs.json from GitHub, parses it into a gendoc.Template, and returns it.
// If docsJsonPath is not empty, you could alternatively load from a local file instead,
// but here we focus on downloading from GitHub.
func GetApiDocsJson() (*gendoc.Template, error) {
	var data []byte
	var err error

	data, err = downloadDocsJson()
	if err != nil {
		return nil, errors.Wrap(err, "failed to download docs JSON")
	}

	var tpl gendoc.Template
	if err := json.Unmarshal(data, &tpl); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal docs JSON into template")
	}

	return &tpl, nil
}

func downloadDocsJson() ([]byte, error) {
	downloadUrl := latestReleaseDownloadUrl
	if version.Version != version.DefaultVersion {
		downloadUrl = fmt.Sprintf(downloadUrlFormatString, version.Version)
	}

	resp, err := http.Get(downloadUrl)
	if err != nil {
		return nil, errors.Wrap(err, "http GET request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrapf(err, "unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	return data, nil
}
