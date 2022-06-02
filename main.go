package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	"github.com/terraform-linters/tflint/formatter"
)

func main() {
	out, err := exec.Command("tflint", "--force", "--format", "json", "/home/stefano/code-climate").Output()

	// Even with --force, tflint returns an exit code != 0 so we check stdout which should be empty in case of fatal errors
	if len(out) == 0 && err != nil {
		log.Fatal(err)
	}

	var result *formatter.JSONOutput
	if err := json.Unmarshal(out, &result); err != nil {
		log.Fatal(err)
	}

	output := codeClimatePrint(*result)
	fmt.Printf(output)
}
