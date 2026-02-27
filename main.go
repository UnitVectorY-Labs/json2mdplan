package main

import (
"bytes"
_ "embed"
"encoding/json"
"flag"
"fmt"
"io"
"math"
"os"
"runtime/debug"
"sort"
"strconv"
"strings"
"unicode"
)

// ──────────────────────────────────────────────
// Version & Embedded Files
// ──────────────────────────────────────────────

var Version = "dev" // This will be set by the build system to the release version

func getVersion() string {
if Version != "dev" {
return Version
}
if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
return info.Main.Version
}
return Version
}

//go:embed template-schema.json
var templateSchemaJSON string

// ──────────────────────────────────────────────
// Error Types
// ──────────────────────────────────────────────

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

func exitCodeForError(err error) int {
switch err.(type) {
case *cliError:
return 2
case *inputError:
return 3
case *validationError:
return 4
default:
return 1
}
}

// ──────────────────────────────────────────────
// Template Data Structures
// ──────────────────────────────────────────────

// TemplateNode describes how a JSON node should be rendered.
type TemplateNode struct {
Render     string                   `json:"render,omitempty"`
Title      string                   `json:"title,omitempty"`
Label      string                   `json:"label,omitempty"`
TitleKey   string                   `json:"title_key,omitempty"`
Order      []string                 `json:"order,omitempty"`
Properties map[string]*TemplateNode `json:"properties,omitempty"`
Items      *TemplateNode            `json:"items,omitempty"`
}

// TemplateFile is the top-level template document.
type TemplateFile struct {
Version  string        `json:"version"`
Template *TemplateNode `json:"template"`
}

// ──────────────────────────────────────────────
// Key-to-Label Utility
// ──────────────────────────────────────────────

// keyToLabel converts a JSON key to a human-friendly label.
//   - snake_case → split on "_", capitalize each word
//   - camelCase  → split on uppercase transitions
//   - ALL_CAPS   → keep as-is
//   - single word → capitalize first letter
func keyToLabel(key string) string {
if key == "" {
return ""
}

// Check if ALL_CAPS (letters+digits+underscores, all letters uppercase)
if isAllCaps(key) {
return key
}

// Check if snake_case (contains underscore)
if strings.Contains(key, "_") {
parts := strings.Split(key, "_")
for i, p := range parts {
if p != "" {
parts[i] = strings.ToUpper(p[:1]) + p[1:]
}
}
return strings.Join(parts, " ")
}

// camelCase split
words := splitCamelCase(key)
for i, w := range words {
if w != "" {
words[i] = strings.ToUpper(w[:1]) + w[1:]
}
}
return strings.Join(words, " ")
}

func isAllCaps(s string) bool {
hasLetter := false
for _, r := range s {
if unicode.IsLetter(r) {
hasLetter = true
if !unicode.IsUpper(r) {
return false
}
} else if r != '_' && !unicode.IsDigit(r) {
return false
}
}
return hasLetter
}

func splitCamelCase(s string) []string {
var words []string
runes := []rune(s)
start := 0
for i := 1; i < len(runes); i++ {
if unicode.IsUpper(runes[i]) && !unicode.IsUpper(runes[i-1]) {
words = append(words, string(runes[start:i]))
start = i
} else if i > 1 && unicode.IsUpper(runes[i-1]) && unicode.IsUpper(runes[i-2]) && !unicode.IsUpper(runes[i]) {
// Handle transitions like "XMLParser" → "XML", "Parser"
words = append(words, string(runes[start:i-1]))
start = i - 1
}
}
words = append(words, string(runes[start:]))
return words
}

// ──────────────────────────────────────────────
// Template Generation
// ──────────────────────────────────────────────

func generateTemplate(data interface{}) *TemplateFile {
return &TemplateFile{
Version:  "1",
Template: generateNode(data, true),
}
}

func generateNode(data interface{}, isRoot bool) *TemplateNode {
switch v := data.(type) {
case map[string]interface{}:
return generateObjectNode(v, isRoot)
case []interface{}:
return generateArrayNode(v)
default:
return &TemplateNode{Render: "labeled_value"}
}
}

func generateObjectNode(obj map[string]interface{}, isRoot bool) *TemplateNode {
node := &TemplateNode{}
if isRoot {
node.Render = "inline"
} else {
node.Render = "section"
}

keys := sortedKeys(obj)
node.Order = keys
node.Properties = make(map[string]*TemplateNode, len(keys))

for _, k := range keys {
child := generateNode(obj[k], false)
child.Label = keyToLabel(k)
if child.Render == "section" {
child.Title = keyToLabel(k)
}
node.Properties[k] = child
}
return node
}

func generateArrayNode(arr []interface{}) *TemplateNode {
if len(arr) == 0 {
return &TemplateNode{Render: "bullet_list"}
}

// Check first element type
first := arr[0]
switch first.(type) {
case map[string]interface{}:
if isArrayOfFlatObjects(arr) {
return generateTableNode(arr)
}
return generateSectionsNode(arr)
default:
return &TemplateNode{Render: "bullet_list"}
}
}

func isArrayOfFlatObjects(arr []interface{}) bool {
for _, item := range arr {
obj, ok := item.(map[string]interface{})
if !ok {
return false
}
for _, v := range obj {
switch v.(type) {
case map[string]interface{}, []interface{}:
return false
}
}
}
return true
}

func generateTableNode(arr []interface{}) *TemplateNode {
// Collect all keys across all objects
keySet := make(map[string]bool)
for _, item := range arr {
obj, ok := item.(map[string]interface{})
if !ok {
continue
}
for k := range obj {
keySet[k] = true
}
}

keys := make([]string, 0, len(keySet))
for k := range keySet {
keys = append(keys, k)
}
sort.Strings(keys)

props := make(map[string]*TemplateNode, len(keys))
for _, k := range keys {
props[k] = &TemplateNode{
Render: "labeled_value",
Label:  keyToLabel(k),
}
}

return &TemplateNode{
Render: "table",
Items: &TemplateNode{
Order:      keys,
Properties: props,
},
}
}

func generateSectionsNode(arr []interface{}) *TemplateNode {
titleKey := detectTitleKey(arr)

// Generate items template from first object
var itemTemplate *TemplateNode
if first, ok := arr[0].(map[string]interface{}); ok {
itemTemplate = generateObjectNode(first, false)
itemTemplate.Render = "inline"
itemTemplate.Title = ""
if titleKey != "" {
itemTemplate.TitleKey = titleKey
}
} else {
itemTemplate = &TemplateNode{Render: "inline"}
}

return &TemplateNode{
Render: "sections",
Items:  itemTemplate,
}
}

var titleKeyPriority = []string{"title", "name", "label", "id"}

func detectTitleKey(arr []interface{}) string {
if len(arr) == 0 {
return ""
}
first, ok := arr[0].(map[string]interface{})
if !ok {
return ""
}
for _, candidate := range titleKeyPriority {
if _, exists := first[candidate]; exists {
return candidate
}
}
return ""
}

// ──────────────────────────────────────────────
// Conversion (Template → Markdown)
// ──────────────────────────────────────────────

type converter struct {
buf     bytes.Buffer
verbose bool
}

func newConverter(verbose bool) *converter {
return &converter{verbose: verbose}
}

func (c *converter) log(format string, args ...interface{}) {
if c.verbose {
fmt.Fprintf(os.Stderr, "[verbose] "+format+"\n", args...)
}
}

func (c *converter) convert(data interface{}, tmpl *TemplateNode) string {
c.renderNode(data, tmpl, 1, true)
return trimTrailingBlankLines(c.buf.String()) + "\n"
}

func (c *converter) renderNode(data interface{}, tmpl *TemplateNode, level int, isRoot bool) {
if data == nil {
return
}

renderMode := c.resolveRenderMode(data, tmpl, isRoot)
c.log("renderNode: mode=%s level=%d", renderMode, level)

if renderMode == "hidden" {
return
}

switch renderMode {
case "section":
c.renderSection(data, tmpl, level)
case "inline":
c.renderInline(data, tmpl, level)
case "table":
c.renderTable(data, tmpl, level)
case "bullet_list":
c.renderBulletList(data, tmpl, level)
case "sections":
c.renderSections(data, tmpl, level)
case "labeled_value":
c.renderLabeledValue(data, tmpl)
case "text":
c.renderText(data)
case "heading":
c.renderHeading(data, level)
default:
// Fallback: render with defaults
c.renderNode(data, &TemplateNode{}, level, isRoot)
}
}

func (c *converter) resolveRenderMode(data interface{}, tmpl *TemplateNode, isRoot bool) string {
if tmpl != nil && tmpl.Render != "" {
// Type mismatch check: if template mode doesn't match data type, use defaults
if !isRenderModeCompatible(tmpl.Render, data) {
c.log("type mismatch: render=%s, using defaults", tmpl.Render)
return c.defaultRenderMode(data, isRoot)
}
return tmpl.Render
}
return c.defaultRenderMode(data, isRoot)
}

func isRenderModeCompatible(mode string, data interface{}) bool {
switch mode {
case "hidden":
return true
case "section", "inline":
_, ok := data.(map[string]interface{})
return ok
case "table", "bullet_list", "sections":
_, ok := data.([]interface{})
return ok
case "labeled_value", "text", "heading":
switch data.(type) {
case map[string]interface{}, []interface{}:
// scalars expected but got complex type - still allow rendering
return true
default:
return true
}
}
return true
}

func (c *converter) defaultRenderMode(data interface{}, isRoot bool) string {
switch v := data.(type) {
case map[string]interface{}:
if isRoot {
return "inline"
}
return "section"
case []interface{}:
if len(v) == 0 {
return "bullet_list"
}
if _, ok := v[0].(map[string]interface{}); ok {
if isArrayOfFlatObjects(v) {
return "table"
}
return "sections"
}
return "bullet_list"
default:
return "labeled_value"
}
}

// ──────────────────────────────────────────────
// Render: Section (object with heading)
// ──────────────────────────────────────────────

func (c *converter) renderSection(data interface{}, tmpl *TemplateNode, level int) {
obj, ok := data.(map[string]interface{})
if !ok {
return
}

title := "Section"
if tmpl != nil && tmpl.Title != "" {
title = tmpl.Title
}

c.writeHeading(level, title)
c.renderObjectProperties(obj, tmpl, level)
}

// ──────────────────────────────────────────────
// Render: Inline (object without heading)
// ──────────────────────────────────────────────

func (c *converter) renderInline(data interface{}, tmpl *TemplateNode, level int) {
obj, ok := data.(map[string]interface{})
if !ok {
return
}
c.renderObjectProperties(obj, tmpl, level)
}

// ──────────────────────────────────────────────
// Object Property Rendering
// ──────────────────────────────────────────────

func (c *converter) renderObjectProperties(obj map[string]interface{}, tmpl *TemplateNode, level int) {
keys := c.orderedKeys(obj, tmpl)

// Group consecutive labeled_value items
type pendingItem struct {
label string
value string
}
var labeledGroup []pendingItem

flushLabeled := func() {
if len(labeledGroup) == 0 {
return
}
for _, item := range labeledGroup {
c.buf.WriteString(fmt.Sprintf("- **%s**: %s\n", item.label, item.value))
}
c.buf.WriteString("\n")
labeledGroup = nil
}

for _, key := range keys {
val, exists := obj[key]
if !exists || val == nil {
continue
}

childTmpl := c.getChildTemplate(tmpl, key)
childRender := c.resolveRenderMode(val, childTmpl, false)

if childRender == "hidden" {
continue
}

if childRender == "labeled_value" {
label := c.getLabelForKey(childTmpl, key)
formatted := formatScalar(val)
if formatted == "" {
continue
}
labeledGroup = append(labeledGroup, pendingItem{label: label, value: formatted})
} else {
flushLabeled()
c.renderNode(val, childTmpl, level+1, false)
}
}
flushLabeled()
}

func (c *converter) orderedKeys(obj map[string]interface{}, tmpl *TemplateNode) []string {
var ordered []string
seen := make(map[string]bool)

// Keys from template order
if tmpl != nil && len(tmpl.Order) > 0 {
for _, k := range tmpl.Order {
if _, exists := obj[k]; exists {
ordered = append(ordered, k)
seen[k] = true
}
}
}

// Remaining keys alphabetically
remaining := make([]string, 0)
for k := range obj {
if !seen[k] {
remaining = append(remaining, k)
}
}
sort.Strings(remaining)
ordered = append(ordered, remaining...)

return ordered
}

func (c *converter) getChildTemplate(tmpl *TemplateNode, key string) *TemplateNode {
if tmpl != nil && tmpl.Properties != nil {
if child, ok := tmpl.Properties[key]; ok {
return child
}
}
return nil
}

func (c *converter) getLabelForKey(tmpl *TemplateNode, key string) string {
if tmpl != nil && tmpl.Label != "" {
return tmpl.Label
}
return keyToLabel(key)
}

// ──────────────────────────────────────────────
// Render: Table (array of flat objects)
// ──────────────────────────────────────────────

func (c *converter) renderTable(data interface{}, tmpl *TemplateNode, level int) {
arr, ok := data.([]interface{})
if !ok || len(arr) == 0 {
return
}

if tmpl != nil && tmpl.Title != "" {
c.writeHeading(level, tmpl.Title)
}

columns := c.getTableColumns(arr, tmpl)
if len(columns) == 0 {
return
}

// Header
headers := make([]string, len(columns))
for i, col := range columns {
headers[i] = c.getColumnHeader(tmpl, col)
}
c.buf.WriteString("| " + strings.Join(headers, " | ") + " |\n")

// Separator
seps := make([]string, len(columns))
for i := range seps {
seps[i] = "---"
}
c.buf.WriteString("| " + strings.Join(seps, " | ") + " |\n")

// Rows
for _, item := range arr {
obj, ok := item.(map[string]interface{})
if !ok {
continue
}
cells := make([]string, len(columns))
for i, col := range columns {
cells[i] = escapeTableCell(formatScalar(obj[col]))
}
c.buf.WriteString("| " + strings.Join(cells, " | ") + " |\n")
}
c.buf.WriteString("\n")
}

func (c *converter) getTableColumns(arr []interface{}, tmpl *TemplateNode) []string {
// From items template order
if tmpl != nil && tmpl.Items != nil && len(tmpl.Items.Order) > 0 {
return tmpl.Items.Order
}

// Collect all keys alphabetically
keySet := make(map[string]bool)
for _, item := range arr {
obj, ok := item.(map[string]interface{})
if !ok {
continue
}
for k := range obj {
keySet[k] = true
}
}
keys := make([]string, 0, len(keySet))
for k := range keySet {
keys = append(keys, k)
}
sort.Strings(keys)
return keys
}

func (c *converter) getColumnHeader(tmpl *TemplateNode, col string) string {
if tmpl != nil && tmpl.Items != nil && tmpl.Items.Properties != nil {
if prop, ok := tmpl.Items.Properties[col]; ok && prop.Label != "" {
return prop.Label
}
}
return keyToLabel(col)
}

func escapeTableCell(s string) string {
return strings.ReplaceAll(s, "|", "\\|")
}

// ──────────────────────────────────────────────
// Render: Bullet List (array of scalars)
// ──────────────────────────────────────────────

func (c *converter) renderBulletList(data interface{}, tmpl *TemplateNode, level int) {
arr, ok := data.([]interface{})
if !ok || len(arr) == 0 {
return
}

if tmpl != nil && tmpl.Title != "" {
c.writeHeading(level, tmpl.Title)
}

for _, item := range arr {
formatted := formatScalar(item)
if formatted != "" {
c.buf.WriteString("- " + formatted + "\n")
}
}
c.buf.WriteString("\n")
}

// ──────────────────────────────────────────────
// Render: Sections (array of objects as sub-sections)
// ──────────────────────────────────────────────

func (c *converter) renderSections(data interface{}, tmpl *TemplateNode, level int) {
arr, ok := data.([]interface{})
if !ok || len(arr) == 0 {
return
}

if tmpl != nil && tmpl.Title != "" {
c.writeHeading(level, tmpl.Title)
}

titleKey := ""
if tmpl != nil && tmpl.Items != nil && tmpl.Items.TitleKey != "" {
titleKey = tmpl.Items.TitleKey
}

for i, item := range arr {
obj, ok := item.(map[string]interface{})
if !ok {
continue
}

// Determine heading
heading := fmt.Sprintf("Item %d", i+1)
if titleKey != "" {
if v, exists := obj[titleKey]; exists && v != nil {
heading = formatScalar(v)
}
}

subLevel := level + 1
if tmpl != nil && tmpl.Title == "" {
subLevel = level
}

c.writeHeading(subLevel, heading)

// Build a modified template that excludes title_key from labeled values
itemTmpl := tmpl.Items
if itemTmpl == nil {
itemTmpl = &TemplateNode{Render: "inline"}
}

// Render object properties, skipping title_key
keys := c.orderedKeys(obj, itemTmpl)

type pendingItem struct {
label string
value string
}
var labeledGroup []pendingItem

flushLabeled := func() {
if len(labeledGroup) == 0 {
return
}
for _, li := range labeledGroup {
c.buf.WriteString(fmt.Sprintf("- **%s**: %s\n", li.label, li.value))
}
c.buf.WriteString("\n")
labeledGroup = nil
}

for _, key := range keys {
// Skip title_key to avoid duplication
if key == titleKey {
continue
}

val, exists := obj[key]
if !exists || val == nil {
continue
}

childTmpl := c.getChildTemplate(itemTmpl, key)
childRender := c.resolveRenderMode(val, childTmpl, false)

if childRender == "hidden" {
continue
}

if childRender == "labeled_value" {
label := c.getLabelForKey(childTmpl, key)
formatted := formatScalar(val)
if formatted == "" {
continue
}
labeledGroup = append(labeledGroup, pendingItem{label: label, value: formatted})
} else {
flushLabeled()
c.renderNode(val, childTmpl, subLevel+1, false)
}
}
flushLabeled()
}
}

// ──────────────────────────────────────────────
// Render: Labeled Value
// ──────────────────────────────────────────────

func (c *converter) renderLabeledValue(data interface{}, tmpl *TemplateNode) {
formatted := formatScalar(data)
if formatted == "" {
return
}

label := "Value"
if tmpl != nil && tmpl.Label != "" {
label = tmpl.Label
}

c.buf.WriteString(fmt.Sprintf("- **%s**: %s\n", label, formatted))
c.buf.WriteString("\n")
}

// ──────────────────────────────────────────────
// Render: Text
// ──────────────────────────────────────────────

func (c *converter) renderText(data interface{}) {
formatted := formatScalar(data)
if formatted == "" {
return
}
c.buf.WriteString(formatted + "\n\n")
}

// ──────────────────────────────────────────────
// Render: Heading
// ──────────────────────────────────────────────

func (c *converter) renderHeading(data interface{}, level int) {
formatted := formatScalar(data)
if formatted == "" {
return
}
c.writeHeading(level, formatted)
}

// ──────────────────────────────────────────────
// Formatting Helpers
// ──────────────────────────────────────────────

func (c *converter) writeHeading(level int, text string) {
if level < 1 {
level = 1
}
if level > 6 {
level = 6
}
c.buf.WriteString(strings.Repeat("#", level) + " " + text + "\n\n")
}

func formatScalar(v interface{}) string {
if v == nil {
return ""
}
switch val := v.(type) {
case string:
return val
case float64:
if math.IsInf(val, 0) || math.IsNaN(val) {
return fmt.Sprintf("%v", val)
}
if val == math.Trunc(val) {
return strconv.FormatInt(int64(val), 10)
}
return strconv.FormatFloat(val, 'f', -1, 64)
case bool:
if val {
return "true"
}
return "false"
case json.Number:
return val.String()
case map[string]interface{}, []interface{}:
b, err := json.Marshal(val)
if err != nil {
return fmt.Sprintf("%v", val)
}
return string(b)
default:
return fmt.Sprintf("%v", val)
}
}

func trimTrailingBlankLines(s string) string {
return strings.TrimRight(s, "\n")
}

func sortedKeys(m map[string]interface{}) []string {
keys := make([]string, 0, len(m))
for k := range m {
keys = append(keys, k)
}
sort.Strings(keys)
return keys
}

// ──────────────────────────────────────────────
// CLI
// ──────────────────────────────────────────────

func main() {
os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
fs := flag.NewFlagSet("json2mdplan", flag.ContinueOnError)
fs.SetOutput(stderr)

var (
generateMode bool
convertMode  bool
inputFile    string
templateFile string
outputFile   string
pretty       bool
verbose      bool
showVersion  bool
)

fs.BoolVar(&generateMode, "generate", false, "Generate template from JSON")
fs.BoolVar(&convertMode, "convert", false, "Convert JSON to Markdown using template")
fs.StringVar(&inputFile, "input", "", "Input JSON file (default: STDIN)")
fs.StringVar(&templateFile, "template", "", "Template file (required for --convert)")
fs.StringVar(&outputFile, "output", "", "Output file (default: STDOUT)")
fs.BoolVar(&pretty, "pretty", false, "Pretty-print JSON (--generate mode)")
fs.BoolVar(&verbose, "verbose", false, "Enable verbose logging to STDERR")
fs.BoolVar(&showVersion, "version", false, "Show version")

if err := fs.Parse(args); err != nil {
return 2
}

if showVersion {
fmt.Fprintf(stdout, "json2mdplan %s\n", getVersion())
return 0
}

if !generateMode && !convertMode {
fmt.Fprintln(stderr, "error: specify either --generate or --convert")
fs.Usage()
return exitCodeForError(&cliError{message: "no mode specified"})
}

if generateMode && convertMode {
fmt.Fprintln(stderr, "error: specify only one of --generate or --convert")
return exitCodeForError(&cliError{message: "conflicting modes"})
}

if convertMode && templateFile == "" {
fmt.Fprintln(stderr, "error: --template is required for --convert mode")
return exitCodeForError(&cliError{message: "missing template"})
}

// Read input JSON
inputData, err := readInput(inputFile, stdin)
if err != nil {
fmt.Fprintf(stderr, "error: %v\n", err)
return exitCodeForError(err)
}

var parsed interface{}
if err := json.Unmarshal(inputData, &parsed); err != nil {
e := &inputError{message: fmt.Sprintf("invalid JSON input: %v", err)}
fmt.Fprintf(stderr, "error: %v\n", e)
return exitCodeForError(e)
}

// Determine output writer
out := stdout
if outputFile != "" {
f, err := os.Create(outputFile)
if err != nil {
e := &cliError{message: fmt.Sprintf("cannot create output file: %v", err)}
fmt.Fprintf(stderr, "error: %v\n", e)
return exitCodeForError(e)
}
defer f.Close()
out = f
}

if generateMode {
return runGenerate(parsed, out, stderr, pretty, verbose)
}
return runConvert(parsed, templateFile, out, stderr, verbose)
}

func readInput(inputFile string, stdin io.Reader) ([]byte, error) {
if inputFile != "" {
data, err := os.ReadFile(inputFile)
if err != nil {
return nil, &inputError{message: fmt.Sprintf("cannot read input file: %v", err)}
}
return data, nil
}
data, err := io.ReadAll(stdin)
if err != nil {
return nil, &inputError{message: fmt.Sprintf("cannot read stdin: %v", err)}
}
return data, nil
}

func runGenerate(data interface{}, out io.Writer, stderr io.Writer, pretty bool, verbose bool) int {
if verbose {
fmt.Fprintln(stderr, "[verbose] generating template from JSON input")
}

tmpl := generateTemplate(data)

var output []byte
var err error
if pretty {
output, err = json.MarshalIndent(tmpl, "", "  ")
} else {
output, err = json.Marshal(tmpl)
}
if err != nil {
fmt.Fprintf(stderr, "error: failed to marshal template: %v\n", err)
return 1
}

fmt.Fprintln(out, string(output))
return 0
}

func runConvert(data interface{}, templateFile string, out io.Writer, stderr io.Writer, verbose bool) int {
if verbose {
fmt.Fprintln(stderr, "[verbose] converting JSON to Markdown using template")
}

tmplData, err := os.ReadFile(templateFile)
if err != nil {
e := &inputError{message: fmt.Sprintf("cannot read template file: %v", err)}
fmt.Fprintf(stderr, "error: %v\n", e)
return exitCodeForError(e)
}

var tmplFile TemplateFile
if err := json.Unmarshal(tmplData, &tmplFile); err != nil {
e := &validationError{message: fmt.Sprintf("invalid template JSON: %v", err)}
fmt.Fprintf(stderr, "error: %v\n", e)
return exitCodeForError(e)
}

if tmplFile.Version != "1" {
e := &validationError{message: fmt.Sprintf("unsupported template version: %q", tmplFile.Version)}
fmt.Fprintf(stderr, "error: %v\n", e)
return exitCodeForError(e)
}

if tmplFile.Template == nil {
e := &validationError{message: "template is missing 'template' field"}
fmt.Fprintf(stderr, "error: %v\n", e)
return exitCodeForError(e)
}

conv := newConverter(verbose)
result := conv.convert(data, tmplFile.Template)
fmt.Fprint(out, result)
return 0
}
