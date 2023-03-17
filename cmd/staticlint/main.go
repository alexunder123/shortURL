package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"

	"github.com/rs/zerolog/log"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"honnef.co/go/tools/staticcheck"
)

const (
	config   = `config.json`
	mainFile = `..\shortener\main.go`
)

type configData struct {
	Staticcheck []string
}

var exitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for direct call os.Exit()",
	Run:  run,
}

func main() {
	data, err := os.ReadFile(config)
	if err != nil {
		log.Fatal().Err(err).Msg("os.ReadFile error")
	}
	var cfg configData
	if err = json.Unmarshal(data, &cfg); err != nil {
		log.Fatal().Err(err).Msg("son.Unmarshal error")
	}
	mychecks := []*analysis.Analyzer{
		exitCheckAnalyzer,
		assign.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
	}
	checks := make(map[string]bool)
	for _, v := range cfg.Staticcheck {
		checks[v] = true
	}

	for _, v := range staticcheck.Analyzers {
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	multichecker.Main(
		mychecks...,
	)
}

func run(pass *analysis.Pass) (interface{}, error) {

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, mainFile, nil, 0)
	if err != nil {
		fmt.Println(err)
	}

	var stack []ast.Node
	ast.Inspect(f, func(n ast.Node) bool {
		if v, ok := n.(*ast.CallExpr); ok {
			if fun, ok := v.Fun.(*ast.SelectorExpr); ok && fun.Sel.Name == "Exit" {
				if len(stack) > 1 {
					isPack := stack[0]
					isFunc := stack[1]
					if isPack.(*ast.File).Name.Name == "main" && isFunc.(*ast.FuncDecl).Name.Name == "main" {
						fmt.Printf("Should not use os.Exit() direct call %v: ", fset.Position(v.Fun.Pos()))
						printer.Fprint(os.Stdout, fset, v)
						fmt.Println()
					}
				}
			}
		}

		if n == nil {
			stack = stack[:len(stack)-1]
		} else {
			stack = append(stack, n)
		}

		return true
	})
	return nil, nil
}
