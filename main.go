package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/ST-Apps/codeclimate-tflint/codeclimate"
	"github.com/ST-Apps/codeclimate-tflint/tflint"
	"github.com/terraform-linters/tflint/cmd"
	"github.com/terraform-linters/tflint/formatter"
)

func getConfigurationArgs() []string {
	// Read configuration file
	log.Println("[main.go/getConfigurationArgs] Reading config.json file...")
	file, err := os.ReadFile("/config.json")
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal into our TFLint configuration struct
	var tflintConfigurationRoot *tflint.TFLintRoot
	if err := json.Unmarshal(file, &tflintConfigurationRoot); err != nil {
		log.Fatal(err)
	}
	log.Printf("[main.go/getConfigurationArgs] Unmarshaled config.json file: %v\n", tflintConfigurationRoot)

	return tflint.ToCLIArguments(tflintConfigurationRoot.Config)
}

func main() {
	// Extract TFLint configuration from config.json file
	log.Println("[main.go/main] Extracting TFLint configuration...")
	args := []string{"tflint", "--force", "--format=json"}
	args = append(args, getConfigurationArgs()...)
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
