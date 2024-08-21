package source

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"strings"
)

func IterateGenDecl(files []*ast.File, tok token.Token) func(func(*ast.File, *ast.GenDecl) bool) {
	return func(yield func(*ast.File, *ast.GenDecl) bool) {
		for _, file := range files {
			for _, decl := range file.Decls {
				d, ok := decl.(*ast.GenDecl)
				if !ok || d.Tok != tok {
					continue
				}
				if !yield(file, d) {
					return
				}
			}
		}
	}
}

func IterateValueSpecs(files []*ast.File) func(func(*ast.File, *ast.ValueSpec) bool) {
	return func(yield func(*ast.File, *ast.ValueSpec) bool) {
		for file, decl := range IterateGenDecl(files, token.VAR) {
			for _, s := range decl.Specs {
				if !yield(file, s.(*ast.ValueSpec)) {
					return
				}
			}
		}
	}
}

//func IterateTypes(files []*ast.File) func(func(*ast.File, *ast.TypeSpec) bool) {
//	return func(yield func(*ast.File, *ast.TypeSpec) bool) {
//		for _, file := range files {
//			for _, decl := range file.Decls {
//				spec, ok := decl.(*ast.GenDecl)
//				if !ok || spec.Tok != token.TYPE {
//					continue
//				}
//				for _, s := range spec.Specs {
//					t, ok := s.(*ast.TypeSpec)
//					if !ok {
//						continue
//					}
//					if !yield(file, t) {
//						return
//					}
//				}
//			}
//		}
//	}
//}

func IterateFunctions(files []*ast.File) func(func(*ast.File, *ast.FuncDecl) bool) {
	return func(yield func(*ast.File, *ast.FuncDecl) bool) {
		for _, file := range files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				if !yield(file, fn) {
					return
				}
			}
		}
	}
}

//func IterateImports(files []*ast.File) func(func(*ast.File, *ast.ImportSpec) bool) {
//	return func(yield func(*ast.File, *ast.ImportSpec) bool) {
//		for _, file := range files {
//			for _, decl := range file.Decls {
//				genDecl, ok := decl.(*ast.GenDecl)
//				if !ok || genDecl.Tok != token.IMPORT {
//					continue
//				}
//				for _, s := range genDecl.Specs {
//					if !yield(file, s.(*ast.ImportSpec)) {
//						return
//					}
//				}
//			}
//		}
//	}
//}

func Format(node ast.Node) string {
	var buf strings.Builder
	if err := printer.Fprint(&buf, token.NewFileSet(), node); err != nil {
		return fmt.Sprintf("formatting error: %v", err)
	}
	return buf.String()
}
