package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/UnitVectorY-Labs/json2mdplan/internal/engine"
	"github.com/UnitVectorY-Labs/json2mdplan/internal/jsondoc"
	"github.com/UnitVectorY-Labs/json2mdplan/internal/plan"
)

func Run(args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("expected subcommand: plan or render")
	}

	switch args[0] {
	case "plan":
		return runPlan(args[1:], stdin, stdout)
	case "render":
		return runRender(args[1:], stdin, stdout)
	default:
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}

func runPlan(args []string, stdin io.Reader, stdout io.Writer) error {
	fs := flag.NewFlagSet("plan", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	inlineJSON := fs.String("json", "", "")
	jsonFile := fs.String("json-file", "", "")
	outFile := fs.String("out-file", "", "")

	if err := fs.Parse(args); err != nil {
		return err
	}

	jsonBytes, err := readJSONInput(stdin, *inlineJSON, *jsonFile)
	if err != nil {
		return err
	}

	root, err := jsondoc.Parse(jsonBytes)
	if err != nil {
		return err
	}

	generatedPlan, err := plan.Generate(root)
	if err != nil {
		return err
	}

	output, err := plan.Marshal(*generatedPlan)
	if err != nil {
		return err
	}

	return writeOutput(stdout, *outFile, output)
}

func runRender(args []string, stdin io.Reader, stdout io.Writer) error {
	fs := flag.NewFlagSet("render", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	inlineJSON := fs.String("json", "", "")
	jsonFile := fs.String("json-file", "", "")
	inlinePlan := fs.String("plan", "", "")
	planFile := fs.String("plan-file", "", "")
	outFile := fs.String("out-file", "", "")

	if err := fs.Parse(args); err != nil {
		return err
	}

	jsonBytes, err := readJSONInput(stdin, *inlineJSON, *jsonFile)
	if err != nil {
		return err
	}

	planBytes, err := readRequiredExplicitInput(*inlinePlan, *planFile, "plan")
	if err != nil {
		return err
	}

	root, err := jsondoc.Parse(jsonBytes)
	if err != nil {
		return err
	}

	parsedPlan, err := plan.Parse(planBytes)
	if err != nil {
		return err
	}

	output, err := engine.Render(root, parsedPlan)
	if err != nil {
		return err
	}

	return writeOutput(stdout, *outFile, []byte(output))
}

func readJSONInput(stdin io.Reader, inline string, file string) ([]byte, error) {
	if inline != "" && file != "" {
		return nil, errors.New("only one of --json or --json-file may be provided")
	}

	switch {
	case inline != "":
		return []byte(inline), nil
	case file != "":
		return os.ReadFile(file)
	default:
		data, err := io.ReadAll(stdin)
		if err != nil {
			return nil, err
		}
		if len(data) == 0 {
			return nil, errors.New("missing JSON input")
		}
		return data, nil
	}
}

func readRequiredExplicitInput(inline string, file string, name string) ([]byte, error) {
	if inline == "" && file == "" {
		return nil, fmt.Errorf("missing %s input", name)
	}
	if inline != "" && file != "" {
		return nil, fmt.Errorf("only one of --%s or --%s-file may be provided", name, name)
	}

	if inline != "" {
		return []byte(inline), nil
	}

	return os.ReadFile(file)
}

func writeOutput(stdout io.Writer, outFile string, data []byte) error {
	if outFile == "" {
		_, err := stdout.Write(data)
		return err
	}

	return os.WriteFile(outFile, data, 0o644)
}
