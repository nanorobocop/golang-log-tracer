package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"go/ast"

	"github.com/sirupsen/logrus"
)

const (
	logrusImportPath = `"github.com/sirupsen/logrus"`
	prefixCall       = "CALL>>>"
	prefixReturn     = "RET>>>"
)

type paths []string

func (p *paths) String() string {
	return strings.Join(*p, ",")
}

func (p *paths) Set(value string) error {
	splitted := strings.Split(value, " ")
	*p = splitted
	return nil
}

var dry = flag.Bool("dry", true, "dry run")

func main() {

	var pathsFlag paths

	flag.Var(&pathsFlag, "paths", "paths")
	flag.Parse()

	logrus.Infof("paths::: %+v", pathsFlag)

	for i, file := range pathsFlag {
		logrus.Infof("Checking path: %s", file)
		filepath.Walk(file, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			logrus.Infof("Found %d: %s", i, path)
			err = modify(path)
			if err != nil {
				logrus.Errorf("Error found while parsing %s: %+v", path, err)
			}
			return nil
		})
	}
}

func modify(f string) error {

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, f, nil, 0)
	if err != nil {
		return err
	}

	logrus.Infof("File: %+v", file)

	changed := addLogs(f, file)

	if changed {
		addImport(file)
	}

	if *dry {
		printer.Fprint(os.Stdout, fset, file)
	} else {
		outWriter, err := os.OpenFile(f, os.O_WRONLY|os.O_TRUNC, 0640)
		if err != nil {
			return err
		}
		defer outWriter.Close()

		printer.Fprint(outWriter, fset, file)
	}

	return err
}

func addLogs(path string, f *ast.File) (changed bool) {

	var curFunc *ast.FuncDecl

	ast.Inspect(f, func(f ast.Node) bool {

		// logrus.Debugf("Inspecting: %+v", f)

		if fn, ok := f.(*ast.FuncDecl); ok {

			curFunc = fn

			changed = true

			// parsing func args
			var argsString string
			var argsValues string

			for _, arg := range fn.Type.Params.List {
				for _, name := range arg.Names {
					if name.Name == "_" {
						continue
					}
					argsString += fmt.Sprintf("%s:%%+v, ", name.Name)
					argsValues += fmt.Sprintf("%s, ", name.Name)
				}
			}

			// parsing func returns
			var retString string
			var retValues string
			if fn.Type.Results != nil {

				for _, ret := range fn.Type.Results.List {
					for _, name := range ret.Names {
						retString += fmt.Sprintf("%%+v")
						retValues += fmt.Sprintf("%s, ", name.Name)
					}
				}
			}

			loglineCall := fmt.Sprintf("logrus.Infof(\"%s %s(%s)\", %s)", prefixCall, path+":"+fn.Name.Name, argsString, argsValues)
			call, err := parser.ParseExpr(loglineCall)
			if err != nil {
				panic(err)
			}
			callExpr := call.(*ast.CallExpr)
			expr := &ast.ExprStmt{}
			expr.X = callExpr

			loglineReturn := fmt.Sprintf("logrus.Infof(\"%s %s\", %s)", prefixReturn, retString, retValues)
			ret, err := parser.ParseExpr(loglineReturn)
			if err != nil {
				panic(err)
			}
			retExpr := ret.(*ast.CallExpr)
			// expr = &ast.ExprStmt{}
			// expr.X = retExpr

			def := &ast.DeferStmt{Call: retExpr}

			fn.Body.List = append([]ast.Stmt{expr, def}, fn.Body.List...)

		}
		if ret, ok := f.(*ast.ReturnStmt); ok {
			var retString string
			var retValues string
			for _, res := range ret.Results {
				retString += fmt.Sprintf("%%+v")
				retValues += fmt.Sprintf("%+v, ", res)
			}
		}

		return true
	})

	return changed
}

func addImport(f *ast.File) {
	hasLogrusImport := false
	for _, im := range f.Imports {
		if im.Path.Value == logrusImportPath && im.Name == nil {
			hasLogrusImport = true
			break
		}
	}

	if !hasLogrusImport {
		importSectionFound := false
		newImport := &ast.ImportSpec{Path: &ast.BasicLit{Value: logrusImportPath}}

		ast.Inspect(f, func(f ast.Node) bool {
			// logrus.Debugf("Inspecting: %+v", f)
			if im, ok := f.(*ast.GenDecl); !hasLogrusImport && ok {
				if im.Tok == token.IMPORT {
					importSectionFound = true
					im.Specs = append(im.Specs, newImport)
				}
			}

			return true
		})
		if !importSectionFound {
			importDecl := &ast.GenDecl{
				Tok:    token.IMPORT,
				Lparen: 1,
				Rparen: 1,
				Specs:  []ast.Spec{newImport},
			}
			f.Decls = append([]ast.Decl{importDecl}, f.Decls...)
		}
	}

}
