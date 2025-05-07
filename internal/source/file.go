package source

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

type File struct {
	*ast.GenDecl
	fileSet            *token.FileSet
	typesCache         map[string]*types.Package
	files              map[string]*ast.File
	packages           []*packages.Package
	outPkg             *types.Package
	outputPackage      string
	packageIdentifiers map[string]string
}

func NewFile(decl *ast.GenDecl, fileSet *token.FileSet, list []*packages.Package) *File {
	if decl != nil {
		if got := decl.Tok; got != token.IMPORT {
			log.Panicf("expected decl to have token.IMPORT Tok got %s", got)
		}
	}
	if fileSet == nil {
		fileSet = token.NewFileSet()
	}
	file := &File{
		GenDecl:            decl,
		fileSet:            fileSet,
		typesCache:         make(map[string]*types.Package),
		files:              make(map[string]*ast.File),
		packages:           make([]*packages.Package, 0),
		packageIdentifiers: make(map[string]string),
	}
	file.addPackages(list)
	return file
}

func (file *File) Package(path string) (*packages.Package, bool) {
	for _, pkg := range file.packages {
		if pkg.PkgPath == path {
			return pkg, true
		}
	}
	return nil, false
}

func (file *File) addPackages(packages []*packages.Package) {
	file.packages = slices.Grow(file.packages, len(packages))
	for _, pkg := range packages {
		if pkg == nil {
			continue
		}
		file.typesCache[pkg.PkgPath] = pkg.Types
		file.packages = append(file.packages, pkg)
	}
}

func (file *File) PackageAtFilepath(p string) (*packages.Package, bool) {
	for _, pkg := range file.packages {
		if len(pkg.GoFiles) > 0 && filepath.Dir(pkg.GoFiles[0]) == p {
			return pkg, true
		}
	}
	return nil, false
}

func (file *File) SetOutputPackage(pkg *types.Package) {
	file.outPkg = pkg
	file.outputPackage = pkg.Path()
}

func (file *File) OutputPackage() string {
	return cmp.Or(file.outputPackage, "main")
}

func (file *File) OutputPackageType() *types.Package {
	return file.outPkg
}

func (file *File) SyntaxFile(pos token.Pos) (*ast.File, *token.FileSet, error) {
	position := file.fileSet.Position(pos)
	fSet := token.NewFileSet()
	f, err := parser.ParseFile(fSet, position.Filename, nil, parser.AllErrors|parser.ParseComments|parser.SkipObjectResolution)
	return f, fSet, err
}

func (file *File) TypeASTExpression(tp types.Type) (ast.Expr, error) {
	s := types.TypeString(tp, func(pkg *types.Package) string {
		return file.Add("", pkg.Path())
	})
	return parser.ParseExpr(s)
}

func (file *File) StructField(pos token.Pos) (*ast.Field, error) {
	f, fileSet, err := file.SyntaxFile(pos)
	if err != nil {
		return nil, err
	}
	position := file.fileSet.Position(pos)
	for _, d := range f.Decls {
		switch decl := d.(type) {
		case *ast.GenDecl:
			for _, s := range decl.Specs {
				switch spec := s.(type) {
				case *ast.TypeSpec:
					tp, ok := spec.Type.(*ast.StructType)
					if !ok {
						continue
					}

					for _, field := range tp.Fields.List {
						for _, name := range field.Names {
							p := fileSet.Position(name.Pos())
							if p != position {
								continue
							}
							return field, nil
						}
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("failed to find field")
}

func (file *File) Types(pkgPath string) (*types.Package, bool) {
	if p, ok := file.typesCache[pkgPath]; ok {
		return p, true
	}
	for _, pkg := range file.packages {
		if pkg.Types.Path() == pkgPath {
			p := pkg.Types
			file.typesCache[pkgPath] = p
			return p, true
		}
	}
	for _, pkg := range file.packages {
		if p, ok := recursivelySearchImports(pkg.Types, pkgPath); ok {
			file.typesCache[pkgPath] = p
			return p, true
		}
	}
	return nil, false
}

func recursivelySearchImports(pt *types.Package, pkgPath string) (*types.Package, bool) {
	for _, pkg := range pt.Imports() {
		if pkg.Path() == pkgPath {
			return pkg, true
		}
	}
	for _, pkg := range pt.Imports() {
		if im, ok := recursivelySearchImports(pkg, pkgPath); ok {
			return im, true
		}
	}
	return nil, false
}

func (file *File) Add(pkgIdent, pkgPath string) string {
	if pkgPath == file.OutputPackage() {
		return ""
	}
	if ident, ok := file.packageIdentifiers[pkgPath]; ok {
		return ident
	}
	if file.GenDecl == nil {
		file.GenDecl = new(ast.GenDecl)
		file.GenDecl.Tok = token.IMPORT
	}
	if pkgIdent == "" {
		pkgIdent = path.Base(pkgPath)
	}
	if pkgPath != file.outputPackage {
		for _, s := range file.GenDecl.Specs {
			spec := s.(*ast.ImportSpec)
			pp, _ := strconv.Unquote(spec.Path.Value)
			if pp == pkgPath {
				if spec.Name != nil && spec.Name.Name != "" && spec.Name.Name != pkgIdent {
					n := spec.Name.Name
					file.packageIdentifiers[pkgPath] = n
					return n
				}
				n := path.Base(pp)
				file.packageIdentifiers[pkgPath] = n
				return n
			}
		}
		var pi *ast.Ident
		if path.Base(pkgPath) != pkgIdent {
			pi = ast.NewIdent(pkgIdent)
		}
		file.GenDecl.Specs = append(file.GenDecl.Specs, &ast.ImportSpec{
			Path: String(pkgPath),
			Name: pi,
		})
		slices.SortFunc(file.GenDecl.Specs, func(a, b ast.Spec) int {
			return strings.Compare(a.(*ast.ImportSpec).Path.Value, b.(*ast.ImportSpec).Path.Value)
		})
	}
	n := pkgIdent
	file.packageIdentifiers[pkgPath] = n
	return n
}

func (file *File) Call(pkgName, pkgPath, funcIdent string, args []ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(file.Add(pkgName, pkgPath)),
			Sel: ast.NewIdent(funcIdent),
		},
		Args: args,
	}
}

func (file *File) ImportSpecs() []*ast.ImportSpec {
	result := make([]*ast.ImportSpec, 0, len(file.GenDecl.Specs))
	for _, spec := range file.GenDecl.Specs {
		result = append(result, spec.(*ast.ImportSpec))
	}
	slices.SortFunc(result, func(a, b *ast.ImportSpec) int { return strings.Compare(a.Path.Value, b.Path.Value) })
	return slices.CompactFunc(result, func(a, b *ast.ImportSpec) bool { return a.Path.Value == b.Path.Value })
}

func (file *File) SortImports() {
	sorted := file.GenDecl.Specs[:0]
	for _, spec := range file.ImportSpecs() {
		sorted = append(sorted, spec)
	}
	file.GenDecl.Specs = sorted
}

func (file *File) AddNetHTTP() string      { return file.Add("", "net/http") }
func (file *File) AddPath() string         { return file.Add("", "path") }
func (file *File) AddHTMLTemplate() string { return file.Add("", "html/template") }
func (file *File) AddContext() string      { return file.Add("", "context") }

func (file *File) HTTPErrorCall(response, message ast.Expr, code int) *ast.CallExpr {
	return file.Call("", "net/http", "Error", []ast.Expr{
		response,
		message,
		HTTPStatusCode(file, code),
	})
}

func (file *File) StrconvAtoiCall(expr ast.Expr) *ast.CallExpr {
	return file.Call("", "strconv", "Atoi", []ast.Expr{expr})
}

func (file *File) StrconvItoaCall(expr ast.Expr) *ast.CallExpr {
	return file.Call("", "strconv", "Itoa", []ast.Expr{expr})
}

func (file *File) StrconvParseIntCall(expr ast.Expr, base, size int) *ast.CallExpr {
	return file.Call("", "strconv", "ParseInt", []ast.Expr{expr, Int(base), Int(size)})
}

func (file *File) StrconvParseUintCall(expr ast.Expr, base, size int) *ast.CallExpr {
	return file.Call("", "strconv", "ParseUint", []ast.Expr{expr, Int(base), Int(size)})
}

func (file *File) StrconvParseFloatCall(expr ast.Expr, size int) *ast.CallExpr {
	return file.Call("", "strconv", "ParseFloat", []ast.Expr{expr, Int(size)})
}

func (file *File) StrconvParseBoolCall(expr ast.Expr) *ast.CallExpr {
	return file.Call("", "strconv", "ParseBool", []ast.Expr{expr})
}

func (file *File) TimeParseCall(layout string, expr ast.Expr) *ast.CallExpr {
	return file.Call("", "time", "Parse", []ast.Expr{String(layout), expr})
}

func (file *File) BytesNewBuffer(expr ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(file.Add("", "bytes")),
			Sel: ast.NewIdent("NewBuffer"),
		},
		Args: []ast.Expr{expr},
	}
}

func (file *File) HTTPRequestPtr() *ast.StarExpr {
	return &ast.StarExpr{
		X: &ast.SelectorExpr{
			X:   ast.NewIdent(file.Add("http", "net/http")),
			Sel: ast.NewIdent("Request"),
		},
	}
}

func (file *File) HTTPResponseWriter() *ast.SelectorExpr {
	return &ast.SelectorExpr{
		X:   ast.NewIdent(file.Add("http", "net/http")),
		Sel: ast.NewIdent("ResponseWriter"),
	}
}

func (file *File) HTTPHeader() *ast.SelectorExpr {
	return &ast.SelectorExpr{
		X:   ast.NewIdent(file.Add("http", "net/http")),
		Sel: ast.NewIdent("Header"),
	}
}

func (file *File) StrconvParseInt8Call(in ast.Expr) *ast.CallExpr {
	return file.StrconvParseIntCall(in, 10, 8)
}

func (file *File) StrconvParseInt16Call(in ast.Expr) *ast.CallExpr {
	return file.StrconvParseIntCall(in, 10, 16)
}

func (file *File) StrconvParseInt32Call(in ast.Expr) *ast.CallExpr {
	return file.StrconvParseIntCall(in, 10, 32)
}

func (file *File) StrconvParseInt64Call(in ast.Expr) *ast.CallExpr {
	return file.StrconvParseIntCall(in, 10, 64)
}

func (file *File) StrconvParseUint0Call(in ast.Expr) *ast.CallExpr {
	return file.StrconvParseUintCall(in, 10, 0)
}

func (file *File) StrconvParseUint8Call(in ast.Expr) *ast.CallExpr {
	return file.StrconvParseUintCall(in, 10, 8)
}

func (file *File) StrconvParseUint16Call(in ast.Expr) *ast.CallExpr {
	return file.StrconvParseUintCall(in, 10, 16)
}

func (file *File) StrconvParseUint32Call(in ast.Expr) *ast.CallExpr {
	return file.StrconvParseUintCall(in, 10, 32)
}

func (file *File) StrconvParseUint64Call(in ast.Expr) *ast.CallExpr {
	return file.StrconvParseUintCall(in, 10, 64)
}

func (file *File) FormatInt(in ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  &ast.SelectorExpr{X: ast.NewIdent(file.Add("", "strconv")), Sel: ast.NewIdent("Itoa")},
		Args: []ast.Expr{in},
	}
}

func (file *File) FormatInt8(in ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  &ast.SelectorExpr{X: ast.NewIdent(file.Add("", "strconv")), Sel: ast.NewIdent("FormatInt")},
		Args: []ast.Expr{&ast.CallExpr{Fun: ast.NewIdent("int64"), Args: []ast.Expr{in}}, Int(10)},
	}
}

func (file *File) FormatInt16(in ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  &ast.SelectorExpr{X: ast.NewIdent(file.Add("", "strconv")), Sel: ast.NewIdent("FormatInt")},
		Args: []ast.Expr{&ast.CallExpr{Fun: ast.NewIdent("int64"), Args: []ast.Expr{in}}, Int(10)},
	}
}

func (file *File) FormatInt32(in ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  &ast.SelectorExpr{X: ast.NewIdent(file.Add("", "strconv")), Sel: ast.NewIdent("FormatInt")},
		Args: []ast.Expr{&ast.CallExpr{Fun: ast.NewIdent("int64"), Args: []ast.Expr{in}}, Int(10)},
	}
}

func (file *File) FormatInt64(in ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  &ast.SelectorExpr{X: ast.NewIdent(file.Add("", "strconv")), Sel: ast.NewIdent("FormatInt")},
		Args: []ast.Expr{&ast.CallExpr{Fun: ast.NewIdent("int64"), Args: []ast.Expr{in}}, Int(10)},
	}
}

func (file *File) FormatUint(in ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  &ast.SelectorExpr{X: ast.NewIdent(file.Add("", "strconv")), Sel: ast.NewIdent("FormatUint")},
		Args: []ast.Expr{&ast.CallExpr{Fun: ast.NewIdent("uint64"), Args: []ast.Expr{in}}, Int(10)},
	}
}

func (file *File) FormatUint8(in ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  &ast.SelectorExpr{X: ast.NewIdent(file.Add("", "strconv")), Sel: ast.NewIdent("FormatUint")},
		Args: []ast.Expr{&ast.CallExpr{Fun: ast.NewIdent("uint64"), Args: []ast.Expr{in}}, Int(10)},
	}
}

func (file *File) FormatUint16(in ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  &ast.SelectorExpr{X: ast.NewIdent(file.Add("", "strconv")), Sel: ast.NewIdent("FormatUint")},
		Args: []ast.Expr{&ast.CallExpr{Fun: ast.NewIdent("uint64"), Args: []ast.Expr{in}}, Int(10)},
	}
}

func (file *File) FormatUint32(in ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  &ast.SelectorExpr{X: ast.NewIdent(file.Add("", "strconv")), Sel: ast.NewIdent("FormatUint")},
		Args: []ast.Expr{&ast.CallExpr{Fun: ast.NewIdent("uint64"), Args: []ast.Expr{in}}, Int(10)},
	}
}

func (file *File) FormatUint64(in ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  &ast.SelectorExpr{X: ast.NewIdent(file.Add("", "strconv")), Sel: ast.NewIdent("FormatUint")},
		Args: []ast.Expr{in, Int(10)},
	}
}

func (file *File) FormatBool(in ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  &ast.SelectorExpr{X: ast.NewIdent(file.Add("", "strconv")), Sel: ast.NewIdent("FormatBool")},
		Args: []ast.Expr{&ast.CallExpr{Fun: ast.NewIdent("bool"), Args: []ast.Expr{in}}},
	}
}

func (file *File) Format(variable ast.Expr, kind types.BasicKind) (ast.Expr, error) {
	switch kind {
	case types.Bool, types.UntypedBool:
		return file.FormatBool(variable), nil
	case types.Int, types.UntypedInt:
		return file.FormatInt(variable), nil
	case types.Int8:
		return file.FormatInt8(variable), nil
	case types.Int16:
		return file.FormatInt16(variable), nil
	case types.Int32:
		return file.FormatInt32(variable), nil
	case types.Int64:
		return file.FormatInt64(variable), nil
	case types.Uint:
		return file.FormatUint(variable), nil
	case types.Uint8:
		return file.FormatUint8(variable), nil
	case types.Uint16:
		return file.FormatUint16(variable), nil
	case types.Uint32:
		return file.FormatUint32(variable), nil
	case types.Uint64:
		return file.FormatUint64(variable), nil
	case types.String:
		return variable, nil
	default:
		return nil, fmt.Errorf("unsupported basic type for path parameters")
	}
}
