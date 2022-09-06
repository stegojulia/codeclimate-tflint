package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ST-Apps/codeclimate-tflint/codeclimate"
	"github.com/ST-Apps/codeclimate-tflint/tflint"
	"github.com/gobwas/glob"
	"github.com/terraform-linters/tflint/cmd"
	"github.com/terraform-linters/tflint/formatter"
	"golang.org/x/exp/slices"
)

func getConfiguration() (codeClimateConfiguration *codeclimate.CodeClimateConfiguration, tflintConfiguration *tflint.TFLintRoot) {
	// Read configuration file
	log.Println("[main.go/getConfigurationArgs] Reading config.json file...")
	file, err := os.ReadFile("/config.json")
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal CodeClimate specific configuration
	var codeClimateConfigurationRoot *codeclimate.CodeClimateConfiguration
	if err := json.Unmarshal(file, &codeClimateConfigurationRoot); err != nil {
		log.Fatal(err)
	}
	log.Printf("[main.go/getConfigurationArgs] Unmarshaled CodeClimate configuration: %v\n", codeClimateConfigurationRoot)

	// Unmarshal into our TFLint configuration struct
	var tflintConfigurationRoot *tflint.TFLintRoot
	if err := json.Unmarshal(file, &tflintConfigurationRoot); err != nil {
		log.Fatal(err)
	}
	log.Printf("[main.go/getConfigurationArgs] Unmarshaled TFLint configuration: %v\n", tflintConfigurationRoot)

	return codeClimateConfigurationRoot, tflintConfigurationRoot
}

func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), err
}

func run(args []string, path string) {
	// Check if the provided path has an allowed extensions, if not we must return
	// This check is valid for files only, directories are always allowed
	allowedExtensions := []string{".tf", ".tfvars"}
	fileExt := filepath.Ext(path)
	IsDirectory, _ := IsDirectory(path)
	if IsDirectory {
		log.Printf("[main.go/main] Skipping extensions check because %v is a directory", path)
	} else if !slices.Contains(allowedExtensions, fileExt) {
		log.Printf("[main.go/main] Skipping TFLint for file %v because extension %v is not allowed", path, fileExt)
		return
	}

	// Redirect TFLint's output to an internal stream
	r, w, _ := os.Pipe()

	// Read back from the internal stram
	// See: https://stackoverflow.com/a/10476304
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// Run TFLint providing some basic options
	tmpArgs := append(args, path)
	log.Printf("[main.go/main] Running TFLint with args: %v", tmpArgs)
	cli := cmd.NewCLI(w, os.Stderr)
	cli.Run(tmpArgs)

	// Close the stream and get the output
	w.Close()
	out := <-outC
	log.Printf("[main.go/main] Completed TFLint run with output: %v", out)

	// We can now get the output using TFLint's JSONOutput format
	log.Println("[main.go/main] Converting TFLint output to CodeClimate format...")
	var result *formatter.JSONOutput
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		log.Fatal(err)
	}

	// Finally, we use TFLint's JSONOutput format to provide results in CodeClimate's one
	codeclimate.CodeClimatePrint(*result)
	log.Println("[main.go/main] Completed converting TFLint output to CodeClimate format")
}

func main() {
	// Extract configuration from config.json file
	log.Println("[main.go/main] Extracting configuration...")
	codeClimateConfiguration, tflintConfiguration := getConfiguration()

	// Load TFLint specific configuration
	args := []string{"tflint", "--force", "--format=json"}
	args = append(args, tflint.ToCLIArguments(tflintConfiguration.Config)...)
	log.Printf("[main.go/main] Extracted TFLint configuration: %v\n", args)

	// Handle exclude patterns by creating a list of files/dirs to be analyzed
	// The final list will contain all the include_paths that do not match any of the expressions in exclude_patterns
	// We first must convert the ExcludePattern into a bash-like glob pattern (e.g. {*.tf, *.something})
	globPattern := fmt.Sprintf("{%v}", strings.Join(codeClimateConfiguration.ExcludePatterns[:], ","))
	log.Printf("[main.go/main] Generated glob pattern from ExcludePatterns: %v\n", globPattern)

	g := glob.MustCompile(globPattern)
	for _, file := range codeClimateConfiguration.IncludePaths {
		if g.Match(file) {
			log.Printf("[main.go/main] File is matching exclude_patterns, excluding it from scanning: %v\n", file)
		} else {
			log.Printf("[main.go/main] File is NOT matching exclude_patterns, adding it to args: %v\n", file)

			// Running in this loop is needed as TFLint only supports one folder per execution.
			// This means that including a folder and other folders/files in our agrs will prevent TFLint from running.
			// For this reason we actually process only one included path per single execution.
			run(args, file)
		}
	}
}
