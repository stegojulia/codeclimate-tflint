package codeclimate

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/terraform-linters/tflint/formatter"
	"github.com/terraform-linters/tflint/tflint"
)

// CodeClimateIssue is a temporary structure for converting TFLint issues to CodeClimate report format.
// See specs here: https://github.com/codeclimate/platform/blob/master/spec/analyzers/SPEC.md#data-types
// We're only mapping the types for which we have data and the required ones
type CodeClimateIssue struct {
	Type           string                `json:"type"`
	CheckName      string                `json:"check_name"`
	Description    string                `json:"description"`
	Content        string                `json:"content,omitempty"`
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

// Downloads the provided link and returns it as a string
func downloadLinkContent(link string) string {
	resp, err := http.Get(link)

	// We don't care about the error as there's no way to recover from it.
	// In case of errors we just return an empty string
	if err != nil || resp.StatusCode != http.StatusOK {
		return ""
	}

	// No errors mean that we can go on and extract the text data
	defer resp.Body.Close()
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)

	// Again, we don't deal with errors
	if err != nil {
		return ""
	}

	// If everything went fine we can return the buffered string's contents
	return buf.String()
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
			Type:        "issue",
			CheckName:   issue.Rule.Name,
			Description: issue.Message,
			Content:     downloadLinkContent(issue.Rule.Link),
			Categories:  []string{"Style"},
			Location: CodeClimateLocation{
				Path: issue.Range.Filename,
				Positions: CodeClimatePositions{
					Begin: CodeClimatePosition{Line: issue.Range.Start.Line, Column: issue.Range.Start.Column},
					End:   CodeClimatePosition{Line: issue.Range.End.Line, Column: issue.Range.End.Column},
				},
			},
			OtherLocations: make([]CodeClimateLocation, len(issue.Callers)),
			Severity:       toCodeClimateSeverity(issue.Rule.Severity),
			Fingerprint:    getMD5Hash(issue.Range.Filename + issue.Rule.Name + issue.Message),
		}
		for i, caller := range issue.Callers {
			ccIssue.OtherLocations[i] = CodeClimateLocation{
				Path: caller.Filename,
				Positions: CodeClimatePositions{
					Begin: CodeClimatePosition{Line: caller.Start.Line, Column: caller.Start.Column},
					End:   CodeClimatePosition{Line: caller.End.Line, Column: caller.End.Column},
				},
			}
		}

		// Since CodeClimate prefers issues to be streamed we just print it out once we find it
		printIssueJson(ccIssue)
	}

	for _, issue := range issues.Errors {
		var ccError CodeClimateIssue
		if issue.Severity == string(tflint.ERROR) {
			ccError = CodeClimateIssue{
				Type:        "issue",
				CheckName:   "tflint_application_error",
				Description: issue.Message,
				Categories:  []string{"Bug Risk"},
				Severity:    toCodeClimateSeverity(issue.Severity),
				Fingerprint: getMD5Hash(issue.Message),
				Location:    CodeClimateLocation{},
			}
		} else {
			ccError = CodeClimateIssue{
				Type:        "issue",
				CheckName:   "tflint_application_error",
				Description: fmt.Sprintf("[%v] %v", issue.Summary, issue.Message),
				Categories:  []string{"Bug Risk"},
				Severity:    toCodeClimateSeverity(issue.Severity),
				Fingerprint: getMD5Hash(issue.Range.Filename + issue.Message),
				Location: CodeClimateLocation{
					Path: issue.Range.Filename,
					Positions: CodeClimatePositions{
						Begin: CodeClimatePosition{Line: issue.Range.Start.Line, Column: issue.Range.Start.Column},
						End:   CodeClimatePosition{Line: issue.Range.End.Line, Column: issue.Range.End.Column},
					},
				},
			}
		}

		// Since CodeClimate prefers issues to be streamed we just print it out once we find it
		printIssueJson(ccError)
	}
}
