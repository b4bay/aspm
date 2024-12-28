package sarif

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// PropertyBag is a set of name/value pair that can store extra metadata
type PropertyBag interface{}

// Level specifies severity of the level
type Level string

const (
	None    Level = "none"
	Note    Level = "note"
	Warning Level = "warning"
	Error   Level = "error"
)

type Kind string

const (
	NotApplicable Kind = "notApplicable"
	Pass          Kind = "pass"
	Fail          Kind = "fail"
	Review        Kind = "review"
	Open          Kind = "open"
	Information   Kind = "informational"
)

// Version is the version of Sarif to use
type Version string

// Version210 represents Version210 of Sarif
const (
	Version210 Version = "2.1.0"
	// @Deprecated - use Version210 instead
	Version210RTM5 Version = "2.1.0-rtm.5"
)

var versions = map[Version]string{
	Version210: "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
	// keeping this for backwards support but marked as deprecated
	Version210RTM5: "https://docs.oasis-open.org/sarif/sarif/v2.1.0/errata01/os/schemas/sarif-schema-2.1.0.json",
}

func getVersionSchema(version Version) (string, error) {
	for ver, schema := range versions {
		if ver == version {
			return schema, nil
		}
	}
	return "", fmt.Errorf("version [%s] is not supported", version)
}

// Report is top level object that represents the log file as a whole
type Report struct {
	Version string `json:"version"`
	Schema  string `json:"$schema,omitempty"` // URI of SARIF Schema
	Runs    []Run  `json:"runs"`
}

// FromFile loads a Report from a file
func FromFile(filename string) (*Report, error) {
	if _, err := os.Stat(filename); err != nil && os.IsNotExist(err) {
		return nil, fmt.Errorf("the provided file path doesn't have a file")
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("the provided filepath could not be opened. %w", err)
	}
	return FromBytes(content)
}

// FromString loads a Report from string content
func FromString(content string) (*Report, error) {
	return FromBytes([]byte(content))
}

// FromBytes loads a Report from a byte array
func FromBytes(content []byte) (*Report, error) {
	var report Report
	if err := json.Unmarshal(content, &report); err != nil {
		return nil, err
	}
	return &report, nil
}

func FromBase64(content string) (*Report, error) {
	data, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return nil, err
	}
	return FromBytes(data)
}

// Run represents a single invocation of a single analysis tool
type Run struct {
	Tool        Tool         `json:"tool"`
	Results     []Result     `json:"results"`
	Invocations []Invocation `json:"invocations,omitempty"`
	Taxonomies  []Taxonomy   `json:"taxonomies,omitempty"`
	ruleToCWE   map[string]string
	ruleToCVE   map[string]string
}

// Added
type Taxonomy struct {
	Name string `json:"name,omitempty"`
	Taxa []Taxa `json:"taxa,omitempty"`
}

// Added
type Taxa struct {
	FullDescription  Description `json:"fullDescription,omitempty"`
	GUID             string      `json:"guid,omitempty"`
	HelpURI          string      `json:"helpUri,omitempty"`
	ID               string      `json:"id,omitempty"`
	ShortDescription Description `json:"shortDescription,omitempty"`
}

type Description struct {
	Text string `json:"text,omitempty"`
}

type Tags []string

// New Creates a new Report or returns an error
func New(version Version, includeSchema ...bool) (*Report, error) {
	schema := ""

	if len(includeSchema) == 0 || includeSchema[0] {
		var err error

		schema, err = getVersionSchema(version)
		if err != nil {
			return nil, err
		}
	}
	return &Report{
		Version: string(version),
		Schema:  schema,
		Runs:    []Run{},
	}, nil
}

func (run *Run) makeRuleToCWEMap() {
	run.ruleToCWE = map[string]string{}
	if run.Tool.Driver.Name == "gosec" {
		var cweTaxonomy Taxonomy
		for _, t := range run.Taxonomies {
			if t.Name == "CWE" {
				cweTaxonomy = t
				break
			}
		}

		for _, tx := range cweTaxonomy.Taxa {
			guid := tx.GUID
			cweId := tx.ID
			var ruleId string
			for _, rule := range run.Tool.Driver.Rules {
				for _, rel := range rule.Relationships {
					if rel.Target.GUID == guid {
						ruleId = rule.Id
						break
					}
				}
				if ruleId != "" {
					break
				}
			}

			run.ruleToCWE[ruleId] = "CWE-" + cweId
		}
	}
}

func (run *Run) makeRuleToCVEMap() {
	run.ruleToCVE = map[string]string{}
	if run.Tool.Driver.Name == "govulncheck" {
		for _, rule := range run.Tool.Driver.Rules {
			propertiesAsMap := rule.Properties.(map[string]interface{})
			tags := propertiesAsMap["tags"]
			//tagsAsTags := tags.(Tags)
			tagsAsSlice := tags.([]interface{})
			for _, tag := range tagsAsSlice {
				tagAsString := tag.(string)
				if strings.HasPrefix(tagAsString, "CVE") {
					run.ruleToCVE[rule.Id] = tagAsString
				}
			}
		}
	}
}

func (run *Run) CWE(result *Result) string {
	if run.ruleToCWE == nil {
		run.makeRuleToCWEMap()
	}
	return run.ruleToCWE[result.RuleId]
}

func (run *Run) CVE(result *Result) string {
	if run.ruleToCVE == nil {
		run.makeRuleToCVEMap()
	}
	return run.ruleToCVE[result.RuleId]
}

// The runtime environment of the analysis tool run
type Invocation struct {
	CommandLine          string             `json:"commandLine,omitempty"`
	Arguments            []string           `json:"arguments,omitempty"`
	ResponseFiles        []ArtifactLocation `json:"responseFiles,omitempty"`
	StartTimeUtc         string             `json:"startTimeUtc,omitempty"` // UTC Time when run started
	EndTimeUtc           string             `json:"endTimeUtc,omitempty"`   // UTC Time when run stopped
	ExecutionSuccessful  bool               `json:"executionSuccessful"`
	ExecutableLocation   ArtifactLocation   `json:"executableLocation,omitempty"`
	WorkingDirectory     ArtifactLocation   `json:"workingDirectory,omitempty"`
	EnvironmentVariables map[string]string  `json:"environmentVariables,omitempty"`
	Stdin                ArtifactLocation   `json:"stdin,omitempty"`
	Stdout               ArtifactLocation   `json:"stdout,omitempty"`
	Stderr               ArtifactLocation   `json:"stderr,omitempty"`
	Properties           PropertyBag        `json:"properties,omitempty"`
}

// Tool describes the analysis tool that produced the run
type Tool struct {
	Driver     ToolComponent   `json:"driver"`               // Tool Driver i.e Actual tool
	Extensions []ToolComponent `json:"extensions,omitempty"` // Tool Extensions/Plugins
	Properties PropertyBag     `json:"properties,omitempty"`
}

// A component of a tool ex: plugin ,template, driver etc
type ToolComponent struct {
	GUID             string                    `json:"guid,omitempty"` //"^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$"
	Name             string                    `json:"name,omitempty"`
	Organization     string                    `json:"organization,omitempty"`
	Product          string                    `json:"product,omitempty"` // A product suite to which the tool component belongs
	ShortDescription *MultiformatMessageString `json:"shortDescription,omitempty"`
	FullDescription  *MultiformatMessageString `json:"fullDescription,omitempty"`
	FullName         string                    `json:"fullName,omitempty"` // Name Along with version
	SemanticVersion  string                    `json:"semanticVersion,omitempty"`
	ReleaseDateUTC   string                    `json:"releaseDateUtc,omitempty"`
	DownloadUri      string                    `json:"downloadUri,omitempty"`
	InformationUri   string                    `json:"informationUri,omitempty"`
	Notifications    []ReportingDescriptor     `json:"notifications,omitempty"`
	Rules            []ReportingDescriptor     `json:"rules,omitempty"`
	Locations        []ArtifactLocation        `json:"locations,omitempty"`
	Properties       PropertyBag               `json:"properties,omitempty"`
}

// Metadata that describes a specific report produced by the tool
type ReportingDescriptor struct {
	Id               string                    `json:"id,omitempty"`
	Name             string                    `json:"name,omitempty"`
	HelpUri          string                    `json:"helpUri,omitempty"`
	ShortDescription *MultiformatMessageString `json:"shortDescription,omitempty"`
	FullDescription  *MultiformatMessageString `json:"fullDescription,omitempty"`
	MessageStrings   *MultiformatMessageString `json:"messageStrings,omitempty"`
	Help             *MultiformatMessageString `json:"help,omitempty"`
	Properties       PropertyBag               `json:"properties,omitempty"`
	// Added
	Relationships []Relationship `json:"relationships,omitempty"`
}

// Added
type Relationship struct {
	Kinds  []string           `json:"kinds,omitempty"`
	Target RelationshipTarget `json:"target,omitempty"`
}

type RelationshipTarget struct {
	GUID string `json:"guid,omitempty"`
	ID   string `json:"id,omitempty"`
}

// Result contains result produced by analysis tool
type Result struct {
	RuleId         string                       `json:"ruleId,omitempty"`
	RuleIndex      int                          `json:"ruleIndex,omitempty"` //The index within the tool component rules array
	Rank           int                          `json:"rank,omitempty"`      // Specifies the relative priority of the report
	Rule           ReportingDescriptorReference `json:"rule,omitempty"`
	Level          Level                        `json:"level,omitempty"`
	Kind           Kind                         `json:"kind,omitempty"`
	Message        *Message                     `json:"message,omitempty"`
	AnalysisTarget ArtifactLocation             `json:"analysisTarget,omitempty"`
	WebRequest     WebRequest                   `json:"webRequest,omitempty"`
	WebResponse    WebResponse                  `json:"webResponse,omitempty"`
	Properties     PropertyBag                  `json:"properties,omitempty"`
	Locations      []Location                   `json:"locations,omitempty"` // location where result was detected
	// Attachments    interface{}                  `json:"attachments,omitempty"`
}

func (r *Result) LocationHash() string {
	var hash string
	loc := r.Locations[0].PhysicalLocation

	if loc.ArtifactLocation.Uri != "" {
		hash = loc.ArtifactLocation.Uri
		if loc.Region.StartLine != 0 {
			hash += fmt.Sprintf("(%d", loc.Region.StartLine)
			if loc.Region.StartColumn != 0 {
				hash += fmt.Sprintf(":%d", loc.Region.StartColumn)
			}
			if loc.Region.EndLine != 0 {
				hash += fmt.Sprintf("-%d", loc.Region.EndLine)
			}
			if loc.Region.EndColumn != 0 {
				hash += fmt.Sprintf(":%d", loc.Region.EndColumn)
			}
			hash += ")"
		}
	}

	return hash
}

// Location
type Location struct {
	Id               int              `json:"id,omitempty"`
	Message          *Message         `json:"message,omitempty"`
	PhysicalLocation PhysicalLocation `json:"physicalLocation,omitempty"`
}

// PhysicalLocation
type PhysicalLocation struct {
	Address          Address          `json:"address,omitempty"`
	ArtifactLocation ArtifactLocation `json:"artifactLocation,omitempty"`
	Properties       PropertyBag      `json:"properties,omitempty"`
	// Added
	Region Region `json:"region,omitempty"`
}

// Added
type Region struct {
	StartLine      int     `json:"startLine,omitempty"`
	StartColumn    int     `json:"startColumn,omitempty"`
	EndLine        int     `json:"endLine,omitempty"`
	EndColumn      int     `json:"endColumn,omitempty"`
	SourceLanguage string  `json:"sourceLanguage,omitempty"`
	Snippet        Snippet `json:"snippet,omitempty"`
}

// Added
type Snippet struct {
	Text string `json:"text,omitempty"`
}

// WebRequest describes http request
type WebRequest struct {
	Protocol   string            `json:"protocol,omitempty"`
	Version    string            `json:"version,omitempty"`
	Target     string            `json:"target,omitempty"`
	Method     string            `json:"method,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`
	Body       ArtifactContent   `json:"body,omitempty"`
	Properties PropertyBag       `json:"properties,omitempty"`
}

// WebResponse describes http Response
type WebResponse struct {
	Protocol   string            `json:"protocol,omitempty"`
	Version    string            `json:"version,omitempty"`
	StatusCode int               `json:"statusCode,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       ArtifactContent   `json:"body,omitempty"`
	Properties PropertyBag       `json:"properties,omitempty"`
}

// ReportingDescriptorReference contains Information about how to locate a relevant reporting descriptor
type ReportingDescriptorReference struct {
	Id            string        `json:"id,omitempty"`
	Index         int           `json:"index,omitempty"`
	GUID          string        `json:"guid,omitempty"`
	ToolComponent ToolComponent `json:"toolComponent,omitempty"`
	Properties    PropertyBag   `json:"properties,omitempty"`
}

// Encapsulates a message intended to be read by the end user
type Message struct {
	Text       string      `json:"text,omitempty"`
	Markdown   string      `json:"markdown,omitempty"`
	Id         string      `json:"id,omitempty"`
	Arguments  []string    `json:"arguments,omitempty"`
	Properties PropertyBag `json:"properties,omitempty"`
}

// A message string or message format string rendered in multiple formats
type MultiformatMessageString struct {
	Text       string      `json:"text,omitempty"`
	Markdown   string      `json:"markdown,omitempty"`
	Properties PropertyBag `json:"properties,omitempty"`
}

// Specifies Location of the Artifact
type ArtifactLocation struct {
	Uri         string      `json:"uri,omitempty"`
	Description *Message    `json:"description,omitempty"`
	Properties  PropertyBag `json:"properties,omitempty"`
}

// Represents Content of Artifact
type ArtifactContent struct {
	Text       string      `json:"text,omitempty"`   // UTF-8 encoded content
	Binary     string      `json:"binary,omitempty"` // Base64 Encoded String
	Properties PropertyBag `json:"properties,omitempty"`
}

// A physical or virtual address, or a range of addresses, in an 'addressable region' (memory or a binary file)
type Address struct {
	Length             int         `json:"length,omitempty"` // number of bytes of address
	Kind               string      `json:"kind,omitempty"`
	Name               string      `json:"name,omitempty"`
	FullyQualifiedName string      `json:"fullyQualifiedName,omitempty"`
	Properties         PropertyBag `json:"properties,omitempty"`
}
