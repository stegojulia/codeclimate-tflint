package codeclimate

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/terraform-linters/tflint/formatter"
)

// CodeClimateIssue is a temporary structure for converting TFLint issues to CodeClimate report format.
// See specs here: https://github.com/codeclimate/platform/blob/master/spec/analyzers/SPEC.md#data-types
// We're only mapping the types for which we have data and the required ones
type CodeClimateIssue struct {
	Type           string                `json:"type"`
	CheckName      string                `json:"check_name"`
	Description    string                `json:"description"`
	Content        CodeClimateContent    `json:"content,omitempty"`
	Categories     []string              `json:"categories"`
	Location       CodeClimateLocation   `json:"location"`
	OtherLocations []CodeClimateLocation `json:"other_locations,omitempty"`
	Fingerprint    string                `json:"fingerprint"`
	Severity       string                `json:"severity,omitempty"`
}

type CodeClimateLocation struct {
	Path      string               `json:"path"`
	Positions CodeClimatePositions `json:"positions"`
}

type CodeClimatePositions struct {
	Begin CodeClimatePosition `json:"begin"`
	End   CodeClimatePosition `json:"end,omitempty"`
}

type CodeClimatePosition struct {
	Line   int `json:"line"`
	Column int `json:"column,omitempty"`
}

type CodeClimateContent struct {
	Body string `json:"body"`
}

// Extracts the rule content from the pre-downloaded rules description on our filesystem (Code Climate doesn't allow network operations)
func generateIssueContent(link string) string {
	// Rules are links with this format: https://github.com/terraform-linters/tflint/blob/<TFLINT VERSION>/docs/rules/terraform_typed_variables.md
	// Our application expects a /tflint-rules folder containing all the files
	// To convert the path we need to strip everything but the file name
	localFile := fmt.Sprintf("/tflint-rules/%v", path.Base(link))

	log.Printf("[formatter.go/generateIssueContent] Generating content for issue: %v (%v)\n", link, localFile)

	file, err := os.ReadFile(localFile)
	if err != nil {
		log.Fatal(err)
	}

	return string(file)
}

func getMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Map TFLint severities with the ones expected by Code Climate
func toCodeClimateSeverity(tfSeverity string) string {
	switch tfSeverity {
	case "error":
		return "critical"
	case "warning":
		return "minor"
	case "info":
		return "info"
	default:
		panic(fmt.Errorf("Unexpected severity type: %s", tfSeverity))
	}
}

func toCodeClimatePosition(jsonRange *formatter.JSONRange) CodeClimateLocation {
	return CodeClimateLocation{
		Path: jsonRange.Filename,
		Positions: CodeClimatePositions{
			Begin: CodeClimatePosition{Line: jsonRange.Start.Line, Column: jsonRange.Start.Column},
			End:   CodeClimatePosition{Line: jsonRange.End.Line, Column: jsonRange.End.Column},
		},
	}
}

func printIssueJson(issue CodeClimateIssue) {
	out, err := json.Marshal(issue)
	if err != nil {
		log.Fatal(err)
		fmt.Print(err)
	}

	// CodeClimate expects issues to be separated by a NULL character (\0)
	fmt.Printf(string(out) + "\x00")
}

func CodeClimatePrint(issues formatter.JSONOutput) {
	for _, issue := range issues.Issues {
		ccIssue := CodeClimateIssue{
			Type:           "issue",
			CheckName:      issue.Rule.Name,
			Description:    issue.Message,
			Content:        CodeClimateContent{Body: generateIssueContent(issue.Rule.Link)},
			Categories:     []string{"Style"},
			Location:       toCodeClimatePosition(&issue.Range),
			OtherLocations: make([]CodeClimateLocation, len(issue.Callers)),
			Severity:       toCodeClimateSeverity(issue.Rule.Severity),
			Fingerprint:    getMD5Hash(issue.Range.Filename + issue.Rule.Name + issue.Message),
		}
		for i, caller := range issue.Callers {
			ccIssue.OtherLocations[i] = toCodeClimatePosition(&caller)
		}

		log.Printf("[formatter.go/CodeClimatePrint] Converting tflint issue\nTF:%+v\nCC:%+v\n", issue, ccIssue)

		// Since CodeClimate prefers issues to be streamed we just print it out once we find it
		printIssueJson(ccIssue)
	}

	for _, issue := range issues.Errors {
		ccError := CodeClimateIssue{
			Type:       "issue",
			CheckName:  "tflint_application_error",
			Categories: []string{"Bug Risk"},
			Severity:   toCodeClimateSeverity(issue.Severity),
		}

		// Originally this was compared with string(tflint.ERROR), but my knowledge of go is not good enough
		// to convert an "enum" to a string without filling this file with spaghetti-code.
		if issue.Severity == "error" {
			ccError.Description = issue.Message
			ccError.Fingerprint = getMD5Hash(issue.Message)
		} else {
			ccError.Description = fmt.Sprintf("[%v] %v", issue.Summary, issue.Message)
			ccError.Fingerprint = getMD5Hash(issue.Range.Filename + issue.Message)
			ccError.Location = toCodeClimatePosition(issue.Range)
		}

		log.Printf("[formatter.go/CodeClimatePrint] Converting tflint application error\nTF:%+v\nCC:%+v\n", issue, ccError)

		// Since CodeClimate prefers issues to be streamed we just print it out once we find it
		printIssueJson(ccError)
	}
}
