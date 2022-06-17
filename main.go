package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/ST-Apps/codeclimate-tflint/codeclimate"
	"github.com/ST-Apps/codeclimate-tflint/tflint"
	"github.com/gobwas/glob"
	"github.com/terraform-linters/tflint/cmd"
	"github.com/terraform-linters/tflint/formatter"
)

func getConfiguration() (codeClimateConfiguration *codeclimate.CodeClimateConfiguration, tflintConfiguration *tflint.TFLintRoot) {
	// Read configuration file
	log.Println("[main.go/getConfigurationArgs] Reading config.json file...")
	file, err := os.ReadFile("config.json")
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
	//return tflint.ToCLIArguments(tflintConfigurationRoot.Config)
}

func main() {
	// Extract configuration from config.json file
	log.Println("[main.go/main] Extracting configuration...")
	codeClimateConfiguration, tflintConfiguration := getConfiguration()

	// Load TFLint specific configuration
	args := []string{"tflint", "--force", "--format=json"}
	args = append(args, tflint.ToCLIArguments(tflintConfiguration.Config)...)

	// Handle exclude patterns by creating a list of files/dirs to be analyzed
	// The final list will contain all the include_paths that do not match any of the expressions in exclude_patterns
	// We first must convert the ExcludePattern into a bash-like glob pattern (e.g. {*.tf, *.something})
	globPattern := fmt.Sprintf("{%v}", strings.Join(codeClimateConfiguration.ExcludePatterns[:], ","))
	log.Printf("[main.go/main] Generated glob pattern from ExcludePatterns: %v\n", globPattern)

	g := glob.MustCompile(globPattern)
	for _, file := range codeClimateConfiguration.IncludePaths {
		if g.Match(file) {
			log.Printf("[main.go/main] File is matching, excluding it from scanning: %v\n", file)
		} else {
			log.Printf("[main.go/main] File is NOT matching, adding it to args: %v\n", file)
			args = append(args, file)
		}
	}
	log.Printf("[main.go/main] Extracted TFLint configuration: %v\n", args)

	// Redirect TFLint's output to an internal stream
	r, w, _ := os.Pipe()

	// Run TFLint providing some basic options
	log.Printf("[main.go/main] Running TFLint with args: %v", args)
	cli := cmd.NewCLI(w, os.Stderr)
	cli.Run(args)

	// Read back from the internal stram
	// See: https://stackoverflow.com/a/10476304
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

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
