package ncdsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/samber/lo"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type OriginalOutClone struct {
	FileName  string  `json:"FileName"`
	StartLine int     `json:"StartLine"`
	EndLine   int     `json:"EndLine"`
	Distance  float64 `json:"Distance"`
	StartChar int     `json:"StartChar"`
	EndChar   int     `json:"EndChar"`
	Tokens    string  `json:"Tokens"`
}

type OriginalOutResult struct {
	Result []*OriginalOutClone `json:"Result"`
}

// SearchOriginal searches clones using original dockerized NCDSearch implementation.
func SearchOriginal(
	basePath string,
	queryFile string,
	startLine int,
	endLine int,
) *OriginalOutResult {
	var outBuf bytes.Buffer
	args := []string{
		"docker", "run", "--rm",
		"--workdir", "/work",
		"-v", fmt.Sprintf("%v:/work", basePath),
		"-e", "BPL_JAVA_NMT_ENABLED=false",
		"registry.toki317.dev/pub/ncdsearch",
		".",
		"-lang", "txt",
		"-th", "0.3",
	}
	if extension := filepath.Ext(queryFile); extension != "" {
		args = append(args, "-i", extension)
	}
	args = append(args,
		"-json",
		"-pos",
		"-q", queryFile,
		"-sline", strconv.Itoa(startLine),
		"-eline", strconv.Itoa(endLine),
	)
	slog.Info("cmd args", "args", args)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = &outBuf
	cmd.Stderr = os.Stderr
	lo.Must0(cmd.Run())

	var res OriginalOutResult
	lo.Must0(json.Unmarshal(outBuf.Bytes(), &res))
	slog.Info("Detection result", "count", len(res.Result))
	return &res
}
