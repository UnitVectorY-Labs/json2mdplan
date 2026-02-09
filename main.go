package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/UnitVectorY-Labs/gcpvalidate/location"
	"github.com/UnitVectorY-Labs/gcpvalidate/project"
	"github.com/UnitVectorY-Labs/gcpvalidate/vertexai"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"golang.org/x/oauth2/google"
)

var Version = "dev" // This will be set by the build system to the release version

// Embedded files
//go:embed plan-schema.json
var planSchemaJSON string

//go:embed plan-system-instructions.md
var systemInstructions string

// Schema validation constants
const planSchemaValidationURL = "plan-schema.json"
const inputSchemaValidationURL = "input-schema.json"

// Exit codes (aligned with prompt2json)
const (
	exitCLIUsageError   = 2
	exitInputError      = 3
	exitValidationError = 4
	exitAPIError        = 5
)

// CLI flags
var (
	// Mode flags
	planMode    bool
	convertMode bool

	// Input flags for --plan mode
	schemaFlag     string
	schemaFileFlag string

	// Input flags for --convert mode
	jsonFlag     string
	jsonFileFlag string
	planFlag     string
	planFileFlag string

	// Output flags
	outFile     string
	prettyPrint bool

	// API flags (for --plan mode)
	projectFlag  string
	locationFlag string
	modelFlag    string
	timeout      int

	// Misc flags
	verbose     bool
	showVersion bool
	showHelp    bool
)

func main() {
	// Set the build version from the build info if not set by the build system
	if Version == "dev" || Version == "" {
		if bi, ok := debug.ReadBuildInfo(); ok {
			if bi.Main.Version != "" && bi.Main.Version != "(devel)" {
				Version = bi.Main.Version
			}
		}
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(getExitCode(err))
	}
}

func run() error {
	defineFlags()
	flag.Parse()

	if showVersion {
		fmt.Fprintf(os.Stderr, "json2mdplan version %s\n", Version)
		return nil
	}

	if showHelp {
		printHelp()
		return nil
	}

	// Validate mode selection
	if planMode && convertMode {
		return &cliError{"cannot specify both --plan and --convert"}
	}
	if !planMode && !convertMode {
		return &cliError{"must specify either --plan or --convert mode"}
	}

	if planMode {
		return runPlanMode()
	}
	return runConvertMode()
}

func defineFlags() {
	// Mode flags
	flag.BoolVar(&planMode, "plan", false, "Generate a plan from JSON Schema using Gemini")
	flag.BoolVar(&convertMode, "convert", false, "Convert JSON to Markdown using schema and plan")

	// Schema input (used in both modes)
	flag.StringVar(&schemaFlag, "schema", "", "JSON Schema (inline JSON)")
	flag.StringVar(&schemaFileFlag, "schema-file", "", "JSON Schema from file")

	// JSON input (--convert mode)
	flag.StringVar(&jsonFlag, "json", "", "JSON instance (inline JSON)")
	flag.StringVar(&jsonFileFlag, "json-file", "", "JSON instance from file")

	// Plan input (--convert mode, also used as output name for --plan mode context)
	flag.StringVar(&planFlag, "plan-json", "", "Plan JSON (inline)")
	flag.StringVar(&planFileFlag, "plan-file", "", "Plan from file")

	// Output
	flag.StringVar(&outFile, "out", "", "Output file path (default: STDOUT)")
	flag.BoolVar(&prettyPrint, "pretty-print", false, "Pretty-print JSON output (--plan mode only)")

	// API flags
	flag.StringVar(&projectFlag, "project", "", "GCP project ID")
	flag.StringVar(&locationFlag, "location", "", "GCP location/region")
	flag.StringVar(&modelFlag, "model", "", "Gemini model identifier")
	flag.IntVar(&timeout, "timeout", 60, "HTTP request timeout in seconds (default: 60)")

	// Misc
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging to STDERR")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.BoolVar(&showHelp, "help", false, "Show help")
}

func printHelp() {
	fmt.Fprintf(os.Stderr, `json2mdplan - Convert JSON to Markdown using schema-validated plans

Usage:
  json2mdplan --plan [OPTIONS]     Generate a plan from JSON Schema
  json2mdplan --convert [OPTIONS]  Convert JSON to Markdown

--plan mode (requires Gemini API):
  Input:
    --schema JSON           JSON Schema (inline)
    --schema-file PATH      JSON Schema from file
    (If neither provided, reads from STDIN)

  API:
    --project ID            GCP project ID (or GOOGLE_CLOUD_PROJECT env)
    --location REGION       GCP location (or GOOGLE_CLOUD_LOCATION env)
    --model NAME            Gemini model identifier

  Output:
    --out PATH              Write plan JSON to file (default: STDOUT)
    --pretty-print          Pretty-print JSON output

--convert mode (no API required):
  Input:
    --json JSON             JSON instance (inline)
    --json-file PATH        JSON instance from file
    (If neither provided, reads from STDIN)

    --schema JSON           JSON Schema (inline)
    --schema-file PATH      JSON Schema from file

    --plan-json JSON        Plan JSON (inline)
    --plan-file PATH        Plan from file

  Output:
    --out PATH              Write Markdown to file (default: STDOUT)

Common options:
  --timeout SECONDS         HTTP request timeout (default: 60, --plan mode only)
  --verbose                 Log diagnostics to STDERR
  --version                 Print version and exit
  --help                    Print help and exit

Environment variables (used if flags not set):
  --project   GOOGLE_CLOUD_PROJECT, CLOUDSDK_CORE_PROJECT
  --location  GOOGLE_CLOUD_LOCATION, GOOGLE_CLOUD_REGION, CLOUDSDK_COMPUTE_REGION

Exit codes:
  0  Success
  2  CLI usage error
  3  Input read/parse error
  4  Validation or response parse error
  5  API/auth error

Example (plan mode):
  json2mdplan --plan \
    --schema-file schema.json \
    --project my-project \
    --location us-central1 \
    --model gemini-2.5-flash \
    --out plan.json

Example (convert mode):
  json2mdplan --convert \
    --json-file data.json \
    --schema-file schema.json \
    --plan-file plan.json \
    --out output.md
`)
}

// ============================================================================
// Plan Mode
// ============================================================================

type PlanConfig struct {
	Schema          map[string]interface{}
	SchemaBytes     []byte
	SchemaSrc       string
	CompiledSchema  *jsonschema.Schema
	Project         string
	Location        string
	Model           string
	Timeout         int
	OutFile         string
	Verbose         bool
	PrettyPrint     bool
	SchemaSourceHint string
}

func runPlanMode() error {
	config, err := loadPlanConfiguration()
	if err != nil {
		return err
	}

	// Generate schema digest for LLM
	digest := generateSchemaDigest(config.Schema)

	// Calculate schema fingerprint
	fingerprint := calculateFingerprint(config.SchemaBytes)

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "Schema fingerprint: %s\n", fingerprint)
		fmt.Fprintf(os.Stderr, "Schema digest generated with %d paths\n", len(digest.PathIndex))
	}

	// Build request body
	requestBody, err := buildPlanRequest(config, digest, fingerprint)
	if err != nil {
		return err
	}

	// Call Gemini API
	responseJSON, err := callGeminiAPI(config, requestBody)
	if err != nil {
		return err
	}

	// Parse and validate the plan
	var plan map[string]interface{}
	if err := json.Unmarshal([]byte(responseJSON), &plan); err != nil {
		return &validationError{fmt.Sprintf("LLM response is not valid JSON: %v", err)}
	}

	// Update fingerprint in the plan (LLM may not have it correct)
	plan["schema_fingerprint"] = map[string]interface{}{
		"sha256":          fingerprint,
		"canonicalization": "json-canonical-v1",
		"source_hint":     config.SchemaSourceHint,
	}

	// Ensure version is set
	plan["version"] = 1

	// Validate plan against plan schema
	if err := validatePlan(plan); err != nil {
		return err
	}

	// Format output
	var outputBytes []byte
	if config.PrettyPrint {
		outputBytes, err = json.MarshalIndent(plan, "", "  ")
	} else {
		outputBytes, err = json.Marshal(plan)
	}
	if err != nil {
		return &validationError{fmt.Sprintf("failed to marshal plan: %v", err)}
	}

	// Write output
	if err := writeOutput(config.OutFile, string(outputBytes)); err != nil {
		return err
	}

	if config.Verbose {
		if config.OutFile != "" {
			fmt.Fprintf(os.Stderr, "Plan written to: %s\n", config.OutFile)
		} else {
			fmt.Fprintf(os.Stderr, "Plan written to: stdout\n")
		}
	}

	return nil
}

func loadPlanConfiguration() (*PlanConfig, error) {
	config := &PlanConfig{
		Verbose:     verbose,
		OutFile:     outFile,
		PrettyPrint: prettyPrint,
	}

	// Load schema
	if schemaFlag != "" && schemaFileFlag != "" {
		return nil, &cliError{"cannot specify both --schema and --schema-file"}
	}

	var schemaBytes []byte
	if schemaFlag != "" {
		schemaBytes = []byte(schemaFlag)
		config.SchemaSrc = "flag"
		config.SchemaSourceHint = "inline"
	} else if schemaFileFlag != "" {
		content, err := os.ReadFile(schemaFileFlag)
		if err != nil {
			return nil, &inputError{fmt.Sprintf("failed to read schema file: %v", err)}
		}
		schemaBytes = content
		config.SchemaSrc = schemaFileFlag
		config.SchemaSourceHint = schemaFileFlag
	} else {
		// Read from STDIN
		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, &inputError{fmt.Sprintf("failed to read schema from STDIN: %v", err)}
		}
		schemaBytes = content
		config.SchemaSrc = "stdin"
		config.SchemaSourceHint = "stdin"
	}

	config.SchemaBytes = schemaBytes

	// Parse schema
	if err := json.Unmarshal(schemaBytes, &config.Schema); err != nil {
		return nil, &inputError{fmt.Sprintf("invalid JSON in schema: %v", err)}
	}

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "Schema: %d bytes (from %s) - valid JSON\n", len(schemaBytes), config.SchemaSrc)
	}

	// Load API configuration
	config.Project = getConfigValue(projectFlag, "GOOGLE_CLOUD_PROJECT", "CLOUDSDK_CORE_PROJECT")
	if config.Project == "" {
		return nil, &cliError{"--project is required (or set GOOGLE_CLOUD_PROJECT)"}
	}

	if !project.IsValidProjectID(config.Project) {
		return nil, &inputError{fmt.Sprintf("invalid GCP project ID: %s", config.Project)}
	}

	config.Location = getConfigValue(locationFlag, "GOOGLE_CLOUD_LOCATION", "GOOGLE_CLOUD_REGION", "CLOUDSDK_COMPUTE_REGION")
	if config.Location == "" {
		return nil, &cliError{"--location is required (or set GOOGLE_CLOUD_LOCATION)"}
	}

	if config.Location != "global" && !location.IsValidRegion(config.Location) {
		return nil, &inputError{fmt.Sprintf("invalid GCP region: %s", config.Location)}
	}

	config.Model = getConfigValue(modelFlag)
	if config.Model == "" {
		return nil, &cliError{"--model is required"}
	}

	if !vertexai.IsValidVertexModelName(config.Model) {
		return nil, &inputError{fmt.Sprintf("invalid Vertex AI model name: %s", config.Model)}
	}

	if timeout < 0 {
		return nil, &cliError{"--timeout must be non-negative"}
	}
	config.Timeout = timeout

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "API configuration: project=%s location=%s model=%s\n", config.Project, config.Location, config.Model)
	}

	return config, nil
}

// ============================================================================
// Schema Digest Generation
// ============================================================================

type SchemaDigest struct {
	SchemaTitle       string           `json:"schema_title,omitempty"`
	SchemaDescription string           `json:"schema_description,omitempty"`
	RootType          string           `json:"root_type"`
	PathIndex         []PathIndexEntry `json:"path_index"`
}

type PathIndexEntry struct {
	Path        string   `json:"path"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	ItemsType   string   `json:"items_type,omitempty"`
	Properties  []string `json:"properties,omitempty"`
}

func generateSchemaDigest(schema map[string]interface{}) *SchemaDigest {
	digest := &SchemaDigest{
		PathIndex: []PathIndexEntry{},
	}

	if title, ok := schema["title"].(string); ok {
		digest.SchemaTitle = title
	}
	if desc, ok := schema["description"].(string); ok {
		digest.SchemaDescription = desc
	}
	if rootType, ok := schema["type"].(string); ok {
		digest.RootType = rootType
	}

	// Walk the schema and build path index
	walkSchema(schema, "", nil, digest)

	return digest
}

func walkSchema(schema map[string]interface{}, path string, requiredFields []string, digest *SchemaDigest) {
	entry := PathIndexEntry{
		Path: path,
	}

	if schemaType, ok := schema["type"].(string); ok {
		entry.Type = schemaType
	}
	if title, ok := schema["title"].(string); ok {
		entry.Title = title
	}
	if desc, ok := schema["description"].(string); ok {
		entry.Description = desc
	}

	// Check if this field is required
	if path != "" {
		fieldName := getFieldNameFromPath(path)
		for _, req := range requiredFields {
			if req == fieldName {
				entry.Required = true
				break
			}
		}
	}

	// Handle object properties
	if props, ok := schema["properties"].(map[string]interface{}); ok {
		var propNames []string
		for name := range props {
			propNames = append(propNames, name)
		}
		sort.Strings(propNames)
		entry.Properties = propNames

		// Get required fields for children
		var childRequired []string
		if req, ok := schema["required"].([]interface{}); ok {
			for _, r := range req {
				if s, ok := r.(string); ok {
					childRequired = append(childRequired, s)
				}
			}
		}

		// Add current entry before recursing
		digest.PathIndex = append(digest.PathIndex, entry)

		// Recurse into properties
		for _, name := range propNames {
			if propSchema, ok := props[name].(map[string]interface{}); ok {
				childPath := path + "/" + name
				walkSchema(propSchema, childPath, childRequired, digest)
			}
		}
		return
	}

	// Handle array items
	if items, ok := schema["items"].(map[string]interface{}); ok {
		if itemType, ok := items["type"].(string); ok {
			entry.ItemsType = itemType
		}
		digest.PathIndex = append(digest.PathIndex, entry)

		// Recurse into items schema if it's an object
		if itemType, _ := items["type"].(string); itemType == "object" {
			var itemRequired []string
			if req, ok := items["required"].([]interface{}); ok {
				for _, r := range req {
					if s, ok := r.(string); ok {
						itemRequired = append(itemRequired, s)
					}
				}
			}
			// Use [*] notation for array items
			walkSchema(items, path+"[*]", itemRequired, digest)
		}
		return
	}

	// Add leaf entry
	digest.PathIndex = append(digest.PathIndex, entry)
}

func getFieldNameFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// ============================================================================
// Schema Fingerprinting
// ============================================================================

func calculateFingerprint(schemaBytes []byte) string {
	// Canonicalize: parse and re-serialize with sorted keys
	var schema interface{}
	if err := json.Unmarshal(schemaBytes, &schema); err != nil {
		// If can't parse, hash raw bytes
		hash := sha256.Sum256(schemaBytes)
		return hex.EncodeToString(hash[:])
	}

	canonicalBytes := canonicalizeJSON(schema)
	hash := sha256.Sum256(canonicalBytes)
	return hex.EncodeToString(hash[:])
}

func canonicalizeJSON(v interface{}) []byte {
	switch val := v.(type) {
	case map[string]interface{}:
		// Sort keys and recursively canonicalize values
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var buf bytes.Buffer
		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			keyBytes, _ := json.Marshal(k)
			buf.Write(keyBytes)
			buf.WriteByte(':')
			buf.Write(canonicalizeJSON(val[k]))
		}
		buf.WriteByte('}')
		return buf.Bytes()

	case []interface{}:
		var buf bytes.Buffer
		buf.WriteByte('[')
		for i, item := range val {
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.Write(canonicalizeJSON(item))
		}
		buf.WriteByte(']')
		return buf.Bytes()

	default:
		// For primitives, use standard JSON encoding
		b, _ := json.Marshal(val)
		return b
	}
}

// ============================================================================
// Gemini API Integration
// ============================================================================

func buildPlanRequest(config *PlanConfig, digest *SchemaDigest, fingerprint string) ([]byte, error) {
	// Build user prompt with schema digest
	digestBytes, err := json.MarshalIndent(digest, "", "  ")
	if err != nil {
		return nil, &inputError{fmt.Sprintf("failed to marshal schema digest: %v", err)}
	}

	userPrompt := fmt.Sprintf(`Here is the Schema Digest for which you need to generate a plan:

%s

The schema fingerprint is: %s

Generate a Plan JSON that will guide deterministic JSON-to-Markdown conversion for documents conforming to this schema.`, string(digestBytes), fingerprint)

	// Parse plan schema for response constraint
	var planSchema map[string]interface{}
	if err := json.Unmarshal([]byte(planSchemaJSON), &planSchema); err != nil {
		return nil, &inputError{fmt.Sprintf("failed to parse plan schema: %v", err)}
	}

	request := map[string]interface{}{
		"systemInstruction": map[string]interface{}{
			"parts": []interface{}{
				map[string]interface{}{
					"text": systemInstructions,
				},
			},
		},
		"contents": []interface{}{
			map[string]interface{}{
				"role": "user",
				"parts": []interface{}{
					map[string]interface{}{
						"text": userPrompt,
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"responseMimeType":   "application/json",
			"responseJsonSchema": planSchema,
		},
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, &inputError{fmt.Sprintf("failed to marshal request: %v", err)}
	}

	return requestBytes, nil
}

func callGeminiAPI(config *PlanConfig, requestBody []byte) (string, error) {
	ctx := context.Background()

	// Get credentials and token
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return "", &apiError{fmt.Sprintf("failed to get credentials: %v", err)}
	}

	token, err := creds.TokenSource.Token()
	if err != nil {
		return "", &apiError{fmt.Sprintf("failed to get access token: %v", err)}
	}

	// Build URL
	var url string
	if config.Location == "global" {
		url = fmt.Sprintf("https://aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:generateContent",
			config.Project, config.Location, config.Model)
	} else {
		url = fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:generateContent",
			config.Location, config.Project, config.Location, config.Model)
	}

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "Request: POST %s\n", url)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return "", &apiError{fmt.Sprintf("failed to create request: %v", err)}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	// Send request
	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", &apiError{fmt.Sprintf("failed to call API: %v", err)}
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &apiError{fmt.Sprintf("failed to read response: %v", err)}
	}

	if resp.StatusCode != http.StatusOK {
		return "", &apiError{fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(respBody))}
	}

	// Parse response
	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason  string `json:"finishReason"`
			FinishMessage string `json:"finishMessage"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return "", &validationError{fmt.Sprintf("failed to parse response: %v", err)}
	}

	if len(geminiResp.Candidates) == 0 {
		return "", &validationError{"no candidates in response"}
	}

	candidate := geminiResp.Candidates[0]

	// Check finish reason
	if candidate.FinishReason != "STOP" {
		errorMsg := fmt.Sprintf("unexpected finish reason: %s", candidate.FinishReason)
		if candidate.FinishMessage != "" {
			errorMsg = fmt.Sprintf("%s (finishMessage: %s)", errorMsg, candidate.FinishMessage)
			fmt.Fprintf(os.Stderr, "Generation stopped: finishReason=%s, finishMessage=%s\n", candidate.FinishReason, candidate.FinishMessage)
		} else {
			fmt.Fprintf(os.Stderr, "Generation stopped: finishReason=%s\n", candidate.FinishReason)
		}
		return "", &validationError{errorMsg}
	}

	if len(candidate.Content.Parts) == 0 {
		return "", &validationError{"no content parts in response"}
	}

	// Concatenate all parts
	var jsonTextBuilder strings.Builder
	for _, part := range candidate.Content.Parts {
		jsonTextBuilder.WriteString(part.Text)
	}
	jsonText := jsonTextBuilder.String()

	if jsonText == "" {
		return "", &validationError{"empty response text"}
	}

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "API response: finish_reason=%s\n", candidate.FinishReason)
		if geminiResp.UsageMetadata.TotalTokenCount > 0 {
			fmt.Fprintf(os.Stderr, "Token usage: prompt=%d, candidates=%d, total=%d\n",
				geminiResp.UsageMetadata.PromptTokenCount,
				geminiResp.UsageMetadata.CandidatesTokenCount,
				geminiResp.UsageMetadata.TotalTokenCount)
		}
	}

	return jsonText, nil
}

func validatePlan(plan map[string]interface{}) error {
	planBytes, err := json.Marshal(plan)
	if err != nil {
		return &validationError{fmt.Sprintf("failed to marshal plan for validation: %v", err)}
	}

	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020
	if err := compiler.AddResource(planSchemaValidationURL, bytes.NewReader([]byte(planSchemaJSON))); err != nil {
		return &validationError{fmt.Sprintf("failed to load plan schema: %v", err)}
	}

	compiledSchema, err := compiler.Compile(planSchemaValidationURL)
	if err != nil {
		return &validationError{fmt.Sprintf("failed to compile plan schema: %v", err)}
	}

	var planObj interface{}
	if err := json.Unmarshal(planBytes, &planObj); err != nil {
		return &validationError{fmt.Sprintf("failed to parse plan: %v", err)}
	}

	if err := compiledSchema.Validate(planObj); err != nil {
		return &validationError{fmt.Sprintf("plan validation failed: %v", err)}
	}

	return nil
}

// ============================================================================
// Convert Mode
// ============================================================================

type ConvertConfig struct {
	JSONInstance    interface{}
	JSONBytes       []byte
	JSONSrc         string
	Schema          map[string]interface{}
	SchemaBytes     []byte
	SchemaSrc       string
	Plan            *Plan
	PlanBytes       []byte
	PlanSrc         string
	OutFile         string
	Verbose         bool
}

type Plan struct {
	Version           int                    `json:"version"`
	SchemaFingerprint SchemaFingerprint      `json:"schema_fingerprint"`
	Settings          PlanSettings           `json:"settings"`
	Overrides         []Override             `json:"overrides"`
}

type SchemaFingerprint struct {
	SHA256          string `json:"sha256"`
	Canonicalization string `json:"canonicalization"`
	SourceHint      string `json:"source_hint,omitempty"`
}

type PlanSettings struct {
	BaseHeadingLevel    int    `json:"base_heading_level"`
	IncludeDescriptions bool   `json:"include_descriptions"`
	DefaultArrayMode    string `json:"default_array_mode"`
	FallbackMode        string `json:"fallback_mode"`
}

type Override struct {
	Path              string   `json:"path"`
	Role              string   `json:"role"`
	Order             []string `json:"order,omitempty"`
	ItemTitleFrom     string   `json:"item_title_from,omitempty"`
	ItemTitleFallback string   `json:"item_title_fallback,omitempty"`
}

func runConvertMode() error {
	config, err := loadConvertConfiguration()
	if err != nil {
		return err
	}

	// Validate JSON instance against schema
	if err := validateJSONAgainstSchema(config); err != nil {
		return err
	}

	// Verify plan-schema compatibility
	schemaFingerprint := calculateFingerprint(config.SchemaBytes)
	if config.Plan.SchemaFingerprint.SHA256 != schemaFingerprint {
		if config.Verbose {
			fmt.Fprintf(os.Stderr, "Warning: schema fingerprint mismatch\n")
			fmt.Fprintf(os.Stderr, "  Plan fingerprint: %s\n", config.Plan.SchemaFingerprint.SHA256)
			fmt.Fprintf(os.Stderr, "  Schema fingerprint: %s\n", schemaFingerprint)
		}
		return &validationError{"plan-schema fingerprint mismatch: the plan was generated for a different schema"}
	}

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "Schema fingerprint verified: %s\n", schemaFingerprint)
	}

	// Convert to Markdown
	markdown := convertToMarkdown(config)

	// Write output
	if err := writeOutput(config.OutFile, markdown); err != nil {
		return err
	}

	if config.Verbose {
		if config.OutFile != "" {
			fmt.Fprintf(os.Stderr, "Markdown written to: %s\n", config.OutFile)
		} else {
			fmt.Fprintf(os.Stderr, "Markdown written to: stdout\n")
		}
	}

	return nil
}

func loadConvertConfiguration() (*ConvertConfig, error) {
	config := &ConvertConfig{
		Verbose: verbose,
		OutFile: outFile,
	}

	// Load JSON instance
	if jsonFlag != "" && jsonFileFlag != "" {
		return nil, &cliError{"cannot specify both --json and --json-file"}
	}

	var jsonBytes []byte
	if jsonFlag != "" {
		jsonBytes = []byte(jsonFlag)
		config.JSONSrc = "flag"
	} else if jsonFileFlag != "" {
		content, err := os.ReadFile(jsonFileFlag)
		if err != nil {
			return nil, &inputError{fmt.Sprintf("failed to read JSON file: %v", err)}
		}
		jsonBytes = content
		config.JSONSrc = jsonFileFlag
	} else {
		// Read from STDIN
		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, &inputError{fmt.Sprintf("failed to read JSON from STDIN: %v", err)}
		}
		jsonBytes = content
		config.JSONSrc = "stdin"
	}

	config.JSONBytes = jsonBytes

	if err := json.Unmarshal(jsonBytes, &config.JSONInstance); err != nil {
		return nil, &inputError{fmt.Sprintf("invalid JSON instance: %v", err)}
	}

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "JSON instance: %d bytes (from %s)\n", len(jsonBytes), config.JSONSrc)
	}

	// Load schema
	if schemaFlag == "" && schemaFileFlag == "" {
		return nil, &cliError{"--schema or --schema-file is required in convert mode"}
	}
	if schemaFlag != "" && schemaFileFlag != "" {
		return nil, &cliError{"cannot specify both --schema and --schema-file"}
	}

	var schemaBytes []byte
	if schemaFlag != "" {
		schemaBytes = []byte(schemaFlag)
		config.SchemaSrc = "flag"
	} else {
		content, err := os.ReadFile(schemaFileFlag)
		if err != nil {
			return nil, &inputError{fmt.Sprintf("failed to read schema file: %v", err)}
		}
		schemaBytes = content
		config.SchemaSrc = schemaFileFlag
	}

	config.SchemaBytes = schemaBytes

	if err := json.Unmarshal(schemaBytes, &config.Schema); err != nil {
		return nil, &inputError{fmt.Sprintf("invalid JSON schema: %v", err)}
	}

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "Schema: %d bytes (from %s)\n", len(schemaBytes), config.SchemaSrc)
	}

	// Load plan
	if planFlag == "" && planFileFlag == "" {
		return nil, &cliError{"--plan-json or --plan-file is required in convert mode"}
	}
	if planFlag != "" && planFileFlag != "" {
		return nil, &cliError{"cannot specify both --plan-json and --plan-file"}
	}

	var planBytes []byte
	if planFlag != "" {
		planBytes = []byte(planFlag)
		config.PlanSrc = "flag"
	} else {
		content, err := os.ReadFile(planFileFlag)
		if err != nil {
			return nil, &inputError{fmt.Sprintf("failed to read plan file: %v", err)}
		}
		planBytes = content
		config.PlanSrc = planFileFlag
	}

	config.PlanBytes = planBytes

	var plan Plan
	if err := json.Unmarshal(planBytes, &plan); err != nil {
		return nil, &inputError{fmt.Sprintf("invalid plan JSON: %v", err)}
	}

	// Validate plan against plan schema
	var planObj interface{}
	if err := json.Unmarshal(planBytes, &planObj); err != nil {
		return nil, &inputError{fmt.Sprintf("invalid plan JSON: %v", err)}
	}

	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020
	if err := compiler.AddResource(planSchemaValidationURL, bytes.NewReader([]byte(planSchemaJSON))); err != nil {
		return nil, &validationError{fmt.Sprintf("failed to load plan schema: %v", err)}
	}

	compiledPlanSchema, err := compiler.Compile(planSchemaValidationURL)
	if err != nil {
		return nil, &validationError{fmt.Sprintf("failed to compile plan schema: %v", err)}
	}

	if err := compiledPlanSchema.Validate(planObj); err != nil {
		return nil, &validationError{fmt.Sprintf("plan validation failed: %v", err)}
	}

	config.Plan = &plan

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "Plan: %d bytes (from %s), version=%d\n", len(planBytes), config.PlanSrc, plan.Version)
	}

	return config, nil
}

func validateJSONAgainstSchema(config *ConvertConfig) error {
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020
	if err := compiler.AddResource(inputSchemaValidationURL, bytes.NewReader(config.SchemaBytes)); err != nil {
		return &validationError{fmt.Sprintf("failed to load input schema: %v", err)}
	}

	compiledSchema, err := compiler.Compile(inputSchemaValidationURL)
	if err != nil {
		return &validationError{fmt.Sprintf("failed to compile input schema: %v", err)}
	}

	if err := compiledSchema.Validate(config.JSONInstance); err != nil {
		return &validationError{fmt.Sprintf("JSON instance validation failed: %v", err)}
	}

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "JSON instance validated against schema\n")
	}

	return nil
}

// ============================================================================
// Markdown Conversion
// ============================================================================

type MarkdownRenderer struct {
	config    *ConvertConfig
	output    strings.Builder
	overrides map[string]Override
}

func convertToMarkdown(config *ConvertConfig) string {
	renderer := &MarkdownRenderer{
		config:    config,
		overrides: make(map[string]Override),
	}

	// Index overrides by path
	for _, override := range config.Plan.Overrides {
		renderer.overrides[override.Path] = override
	}

	// Find document_title and prominent_paragraph
	var documentTitlePath string
	var prominentParagraphPath string
	for path, override := range renderer.overrides {
		if override.Role == "document_title" {
			documentTitlePath = path
		} else if override.Role == "prominent_paragraph" {
			prominentParagraphPath = path
		}
	}

	baseLevel := config.Plan.Settings.BaseHeadingLevel
	if baseLevel < 1 {
		baseLevel = 1
	}

	// Render document title if specified
	if documentTitlePath != "" {
		titleValue := renderer.getValueAtPath(config.JSONInstance, documentTitlePath)
		if titleValue != nil {
			if titleStr, ok := titleValue.(string); ok && titleStr != "" {
				renderer.writeHeading(baseLevel, titleStr)
			}
		}
	}

	// Render prominent paragraph if specified
	if prominentParagraphPath != "" {
		paragraphValue := renderer.getValueAtPath(config.JSONInstance, prominentParagraphPath)
		if paragraphValue != nil {
			if paragraphStr, ok := paragraphValue.(string); ok && paragraphStr != "" {
				renderer.writeParagraph(paragraphStr)
			}
		}
	}

	// Render the rest of the document
	renderer.renderNode(config.JSONInstance, config.Schema, "", baseLevel, documentTitlePath, prominentParagraphPath)

	return strings.TrimSpace(renderer.output.String())
}

func (r *MarkdownRenderer) renderNode(value interface{}, schema map[string]interface{}, path string, depth int, skipTitle string, skipParagraph string) {
	// Check for suppress override
	if override, ok := r.overrides[path]; ok && override.Role == "suppress" {
		return
	}

	// Check for render_as_json override
	if override, ok := r.overrides[path]; ok && override.Role == "render_as_json" {
		r.renderAsJSON(value)
		return
	}

	if value == nil {
		return
	}

	schemaType, _ := schema["type"].(string)

	switch schemaType {
	case "object":
		r.renderObject(value, schema, path, depth, skipTitle, skipParagraph)
	case "array":
		r.renderArray(value, schema, path, depth)
	case "string", "number", "integer", "boolean":
		r.renderScalar(value, schema, path, depth, skipTitle, skipParagraph)
	default:
		// Fallback for unknown types
		r.renderAsJSON(value)
	}
}

func (r *MarkdownRenderer) renderObject(value interface{}, schema map[string]interface{}, path string, depth int, skipTitle string, skipParagraph string) {
	obj, ok := value.(map[string]interface{})
	if !ok {
		r.renderAsJSON(value)
		return
	}

	properties, _ := schema["properties"].(map[string]interface{})
	if properties == nil {
		r.renderAsJSON(value)
		return
	}

	// Get required fields
	var requiredFields []string
	if req, ok := schema["required"].([]interface{}); ok {
		for _, r := range req {
			if s, ok := r.(string); ok {
				requiredFields = append(requiredFields, s)
			}
		}
	}

	// Determine property order
	orderedProps := r.getPropertyOrder(path, properties, requiredFields)

	// Render each property
	for _, propName := range orderedProps {
		propPath := path + "/" + propName
		
		// Skip if this is the document title or prominent paragraph (already rendered)
		if propPath == skipTitle || propPath == skipParagraph {
			continue
		}

		// Check for suppress override on this property
		if override, ok := r.overrides[propPath]; ok && override.Role == "suppress" {
			continue
		}

		propValue, exists := obj[propName]
		if !exists || propValue == nil {
			continue
		}

		propSchema, ok := properties[propName].(map[string]interface{})
		if !ok {
			continue
		}

		propType, _ := propSchema["type"].(string)
		label := r.getLabel(propSchema, propName)

		// For scalars, emit heading then value
		if propType == "string" || propType == "number" || propType == "integer" || propType == "boolean" {
			r.writeHeading(r.clampHeadingLevel(depth), label)
			if r.config.Plan.Settings.IncludeDescriptions {
				if desc, ok := propSchema["description"].(string); ok && desc != "" {
					r.writeParagraph(desc)
				}
			}
			r.writeParagraph(formatScalar(propValue))
		} else if propType == "object" || propType == "array" {
			// For objects and arrays, emit heading then recurse
			r.writeHeading(r.clampHeadingLevel(depth), label)
			if r.config.Plan.Settings.IncludeDescriptions {
				if desc, ok := propSchema["description"].(string); ok && desc != "" {
					r.writeParagraph(desc)
				}
			}
			r.renderNode(propValue, propSchema, propPath, depth+1, skipTitle, skipParagraph)
		} else {
			// Unknown type - fallback
			r.writeHeading(r.clampHeadingLevel(depth), label)
			r.renderAsJSON(propValue)
		}
	}
}

func (r *MarkdownRenderer) renderArray(value interface{}, schema map[string]interface{}, path string, depth int) {
	arr, ok := value.([]interface{})
	if !ok {
		r.renderAsJSON(value)
		return
	}

	if len(arr) == 0 {
		return
	}

	itemsSchema, _ := schema["items"].(map[string]interface{})
	if itemsSchema == nil {
		r.renderAsJSON(value)
		return
	}

	itemType, _ := itemsSchema["type"].(string)

	// Check for array_section override
	var arraySectionOverride *Override
	if override, ok := r.overrides[path]; ok && override.Role == "array_section" {
		arraySectionOverride = &override
	}

	switch itemType {
	case "object":
		// Array of objects - render each as subsection
		for i, item := range arr {
			itemTitle := r.getArrayItemTitle(item, i, arraySectionOverride)
			r.writeHeading(r.clampHeadingLevel(depth), itemTitle)
			
			itemPath := fmt.Sprintf("%s/%d", path, i)
			
			// Determine which field to skip (used for item title)
			skipItemTitleField := ""
			if arraySectionOverride != nil && arraySectionOverride.ItemTitleFrom != "" {
				// Convert relative path like "/name" to absolute path like "/members/0/name"
				skipItemTitleField = itemPath + arraySectionOverride.ItemTitleFrom
			}
			
			r.renderNode(item, itemsSchema, itemPath, depth+1, skipItemTitleField, "")
		}

	case "string", "number", "integer", "boolean":
		// Array of scalars - render each as paragraph
		for _, item := range arr {
			r.writeParagraph(formatScalar(item))
		}

	default:
		// Complex or unknown item type - fallback
		r.renderAsJSON(value)
	}
}

func (r *MarkdownRenderer) renderScalar(value interface{}, schema map[string]interface{}, path string, depth int, skipTitle string, skipParagraph string) {
	// Skip if already rendered as title or prominent paragraph
	if path == skipTitle || path == skipParagraph {
		return
	}

	r.writeParagraph(formatScalar(value))
}

func (r *MarkdownRenderer) renderAsJSON(value interface{}) {
	jsonBytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		r.output.WriteString("```json\n")
		r.output.WriteString(fmt.Sprintf("%v", value))
		r.output.WriteString("\n```\n\n")
		return
	}
	r.output.WriteString("```json\n")
	r.output.Write(jsonBytes)
	r.output.WriteString("\n```\n\n")
}

func (r *MarkdownRenderer) getPropertyOrder(path string, properties map[string]interface{}, requiredFields []string) []string {
	// Check for object_order override
	if override, ok := r.overrides[path]; ok && override.Role == "object_order" && len(override.Order) > 0 {
		// Start with specified order
		ordered := make([]string, 0)
		seen := make(map[string]bool)

		for _, prop := range override.Order {
			if _, exists := properties[prop]; exists {
				ordered = append(ordered, prop)
				seen[prop] = true
			}
		}

		// Add remaining properties using default order
		remaining := r.getDefaultPropertyOrder(properties, requiredFields)
		for _, prop := range remaining {
			if !seen[prop] {
				ordered = append(ordered, prop)
			}
		}

		return ordered
	}

	// Default ordering
	return r.getDefaultPropertyOrder(properties, requiredFields)
}

func (r *MarkdownRenderer) getDefaultPropertyOrder(properties map[string]interface{}, requiredFields []string) []string {
	requiredSet := make(map[string]bool)
	for _, f := range requiredFields {
		requiredSet[f] = true
	}

	var required, optional []string
	for name := range properties {
		if requiredSet[name] {
			required = append(required, name)
		} else {
			optional = append(optional, name)
		}
	}

	sort.Strings(required)
	sort.Strings(optional)

	return append(required, optional...)
}

func (r *MarkdownRenderer) getLabel(schema map[string]interface{}, propertyName string) string {
	if title, ok := schema["title"].(string); ok && title != "" {
		return title
	}
	return propertyName
}

func (r *MarkdownRenderer) getArrayItemTitle(item interface{}, index int, override *Override) string {
	if override != nil && override.ItemTitleFrom != "" {
		titleValue := r.getValueAtPath(item, override.ItemTitleFrom)
		if titleValue != nil {
			if titleStr, ok := titleValue.(string); ok && titleStr != "" {
				return titleStr
			}
		}
	}

	// Use fallback
	if override != nil && override.ItemTitleFallback != "" {
		return strings.ReplaceAll(override.ItemTitleFallback, "{{index}}", strconv.Itoa(index+1))
	}

	return fmt.Sprintf("Item %d", index+1)
}

func (r *MarkdownRenderer) getValueAtPath(value interface{}, path string) interface{} {
	if path == "" {
		return value
	}

	// Remove leading slash
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")

	current := value
	for _, part := range parts {
		if part == "" {
			continue
		}

		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		case []interface{}:
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 0 || idx >= len(v) {
				return nil
			}
			current = v[idx]
		default:
			return nil
		}
	}

	return current
}

func (r *MarkdownRenderer) clampHeadingLevel(depth int) int {
	level := r.config.Plan.Settings.BaseHeadingLevel + depth
	if level > 6 {
		return 6
	}
	if level < 1 {
		return 1
	}
	return level
}

func (r *MarkdownRenderer) writeHeading(level int, text string) {
	r.output.WriteString(strings.Repeat("#", level))
	r.output.WriteString(" ")
	r.output.WriteString(escapeMarkdown(text))
	r.output.WriteString("\n\n")
}

func (r *MarkdownRenderer) writeParagraph(text string) {
	r.output.WriteString(escapeMarkdown(text))
	r.output.WriteString("\n\n")
}

func formatScalar(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		// Check if it's an integer value
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

func escapeMarkdown(text string) string {
	// Basic markdown escaping for text content
	// Escape characters that have special meaning in markdown
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"`", "\\`",
		"*", "\\*",
		"_", "\\_",
		"{", "\\{",
		"}", "\\}",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		".", "\\.",
		"!", "\\!",
		"|", "\\|",
	)
	return replacer.Replace(text)
}

// ============================================================================
// Utility Functions
// ============================================================================

func getConfigValue(flagValue string, envVars ...string) string {
	if flagValue != "" {
		return flagValue
	}
	for _, envVar := range envVars {
		if val := os.Getenv(envVar); val != "" {
			return val
		}
	}
	return ""
}

func writeOutput(outFile string, content string) error {
	if outFile != "" {
		if err := os.WriteFile(outFile, []byte(content), 0644); err != nil {
			return &inputError{fmt.Sprintf("failed to write output file: %v", err)}
		}
	} else {
		fmt.Println(content)
	}
	return nil
}

// ============================================================================
// Error Types
// ============================================================================

type cliError struct {
	message string
}

func (e *cliError) Error() string {
	return e.message
}

type inputError struct {
	message string
}

func (e *inputError) Error() string {
	return e.message
}

type validationError struct {
	message string
}

func (e *validationError) Error() string {
	return e.message
}

type apiError struct {
	message string
}

func (e *apiError) Error() string {
	return e.message
}

func getExitCode(err error) int {
	switch err.(type) {
	case *cliError:
		return exitCLIUsageError
	case *inputError:
		return exitInputError
	case *validationError:
		return exitValidationError
	case *apiError:
		return exitAPIError
	default:
		return exitValidationError
	}
}
