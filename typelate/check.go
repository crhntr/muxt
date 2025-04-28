package typelate

import (
	"fmt"
	"go/token"
	"go/types"
	"maps"
	"strconv"
	"strings"
	"text/template/parse"
)

// TreeFinder should wrap https://pkg.go.dev/html/template#Template.Lookup and return the Tree field from the Template
// If you are using text/template the lookup function from that package should also work.
type TreeFinder interface {
	FindTree(name string) (*parse.Tree, bool)
}

type FindTreeFunc func(name string) (*parse.Tree, bool)

func (fn FindTreeFunc) FindTree(name string) (*parse.Tree, bool) {
	return fn(name)
}

type CallChecker interface {
	CheckCall(string, []parse.Node, []types.Type) (types.Type, error)
}

func Check(tree *parse.Tree, data types.Type, pkg *types.Package, fileSet *token.FileSet, trees TreeFinder, fnChecker CallChecker) error {
	s := &scope{
		global: global{
			TreeFinder:  trees,
			CallChecker: fnChecker,
			pkg:         pkg,
			fileSet:     fileSet,
		},
		variables: map[string]types.Type{
			"$": data,
		},
	}
	_, err := s.walk(tree, data, nil, tree.Root)
	return err
}

type global struct {
	TreeFinder
	CallChecker

	pkg     *types.Package
	fileSet *token.FileSet
}

type scope struct {
	global
	variables map[string]types.Type
}

func (s *scope) child() *scope {
	return &scope{
		global:    s.global,
		variables: maps.Clone(s.variables),
	}
}

func (s *scope) walk(tree *parse.Tree, dot, prev types.Type, node parse.Node) (types.Type, error) {
	switch n := node.(type) {
	case *parse.DotNode:
		return dot, nil
	case *parse.ListNode:
		return nil, s.checkListNode(tree, dot, prev, n)
	case *parse.ActionNode:
		return nil, s.checkActionNode(tree, dot, prev, n)
	case *parse.CommandNode:
		return s.checkCommandNode(tree, dot, prev, n)
	case *parse.FieldNode:
		return s.checkFieldNode(tree, dot, n, nil)
	case *parse.PipeNode:
		return s.checkPipeNode(tree, dot, n)
	case *parse.IfNode:
		return nil, s.checkIfNode(tree, dot, n)
	case *parse.RangeNode:
		return nil, s.checkRangeNode(tree, dot, n)
	case *parse.TemplateNode:
		return nil, s.checkTemplateNode(tree, dot, n)
	case *parse.BoolNode:
		return types.Typ[types.Bool], nil
	case *parse.StringNode:
		return types.Typ[types.String], nil
	case *parse.NumberNode:
		return newNumberNodeType(n)
	case *parse.VariableNode:
		return s.checkVariableNode(tree, n, nil)
	case *parse.IdentifierNode:
		return s.checkIdentifierNode(tree, n)
	case *parse.TextNode:
		return nil, nil
	case *parse.WithNode:
		return nil, s.checkWithNode(tree, dot, n)
	case *parse.CommentNode:
		return nil, nil
	case *parse.NilNode:
		return types.Typ[types.UntypedNil], nil
	case *parse.ChainNode:
		return s.checkChainNode(tree, dot, prev, n, nil)
	case *parse.BranchNode:
		return nil, nil
	case *parse.BreakNode:
		return nil, nil
	case *parse.ContinueNode:
		return nil, nil
	default:
		return nil, fmt.Errorf("missing node type check %T", n)
	}
}

func (s *scope) checkChainNode(tree *parse.Tree, dot, prev types.Type, n *parse.ChainNode, args []types.Type) (types.Type, error) {
	x, err := s.walk(tree, dot, prev, n.Node)
	if err != nil {
		return nil, err
	}
	return s.checkIdentifiers(tree, x, n, n.Field, args)
}

func (s *scope) checkVariableNode(tree *parse.Tree, n *parse.VariableNode, args []types.Type) (types.Type, error) {
	tp, ok := s.variables[n.Ident[0]]
	if !ok {
		return nil, fmt.Errorf("variable %s not found", n.Ident[0])
	}
	return s.checkIdentifiers(tree, tp, n, n.Ident[1:], args)
}

func (s *scope) checkListNode(tree *parse.Tree, dot, prev types.Type, n *parse.ListNode) error {
	for _, child := range n.Nodes {
		if _, err := s.walk(tree, dot, prev, child); err != nil {
			return err
		}
	}
	return nil
}

func (s *scope) checkActionNode(tree *parse.Tree, dot, prev types.Type, n *parse.ActionNode) error {
	_, err := s.walk(tree, dot, prev, n.Pipe)
	return err
}

func (s *scope) checkPipeNode(tree *parse.Tree, dot types.Type, n *parse.PipeNode) (types.Type, error) {
	var result types.Type
	for _, cmd := range n.Cmds {
		tp, err := s.walk(tree, dot, result, cmd)
		if err != nil {
			return nil, err
		}
		result = tp
	}
	if len(n.Decl) > 0 && len(n.Decl[0].Ident) > 0 {
		s.variables[n.Decl[0].Ident[0]] = result
	}
	return result, nil
}

func (s *scope) checkIfNode(tree *parse.Tree, dot types.Type, n *parse.IfNode) error {
	_, err := s.walk(tree, dot, nil, n.Pipe)
	if err != nil {
		return err
	}
	ifScope := s.child()
	if _, err := ifScope.walk(tree, dot, nil, n.List); err != nil {
		return err
	}
	if n.ElseList != nil {
		elseScope := s.child()
		if _, err := elseScope.walk(tree, dot, nil, n.ElseList); err != nil {
			return err
		}
	}
	return nil
}

func (s *scope) checkWithNode(tree *parse.Tree, dot types.Type, n *parse.WithNode) error {
	child := s.child()
	x, err := child.walk(tree, dot, nil, n.Pipe)
	if err != nil {
		return err
	}
	withScope := child.child()
	if _, err := withScope.walk(tree, x, nil, n.List); err != nil {
		return err
	}
	if n.ElseList != nil {
		elseScope := child.child()
		if _, err := elseScope.walk(tree, dot, nil, n.ElseList); err != nil {
			return err
		}
	}
	return nil
}

func newNumberNodeType(constant *parse.NumberNode) (types.Type, error) {
	switch {
	case constant.IsComplex:
		return types.Typ[types.UntypedComplex], nil

	case constant.IsFloat &&
		!isHexInt(constant.Text) && !isRuneInt(constant.Text) &&
		strings.ContainsAny(constant.Text, ".eEpP"):
		return types.Typ[types.UntypedFloat], nil

	case constant.IsInt:
		n := int(constant.Int64)
		if int64(n) != constant.Int64 {
			return nil, fmt.Errorf("%s overflows int", constant.Text)
		}
		return types.Typ[types.UntypedInt], nil

	case constant.IsUint:
		return nil, fmt.Errorf("%s overflows int", constant.Text)
	}
	return types.Typ[types.UntypedInt], nil
}

func isRuneInt(s string) bool {
	return len(s) > 0 && s[0] == '\''
}

func isHexInt(s string) bool {
	return len(s) > 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') && !strings.ContainsAny(s, "pP")
}

func (s *scope) checkTemplateNode(tree *parse.Tree, dot types.Type, n *parse.TemplateNode) error {
	x := dot
	if n.Pipe != nil {
		tp, err := s.walk(tree, x, nil, n.Pipe)
		if err != nil {
			return err
		}
		x = tp
		x = downgradeUntyped(x)
	} else {
		x = types.Typ[types.UntypedNil]
	}
	childTree, ok := s.FindTree(n.Name)
	if !ok {
		return fmt.Errorf("template %q not found", n.Name)
	}
	childScope := scope{
		global: s.global,
		variables: map[string]types.Type{
			"$": x,
		},
	}
	_, err := childScope.walk(childTree, x, nil, childTree.Root)
	return err
}

func downgradeUntyped(x types.Type) types.Type {
	if x == nil {
		return x
	}
	basic, ok := x.Underlying().(*types.Basic)
	if !ok {
		return x
	}
	switch k := basic.Kind(); k {
	case types.UntypedInt:
		return types.Typ[types.Int].Underlying()
	case types.UntypedRune:
		return types.Typ[types.Rune].Underlying()
	case types.UntypedFloat:
		return types.Typ[types.Float64].Underlying()
	case types.UntypedComplex:
		return types.Typ[types.Complex128].Underlying()
	case types.UntypedString:
		return types.Typ[types.String].Underlying()
	default:
		return x
	}
}

func (s *scope) checkFieldNode(tree *parse.Tree, dot types.Type, n *parse.FieldNode, args []types.Type) (types.Type, error) {
	return s.checkIdentifiers(tree, dot, n, n.Ident, args)
}

func (s *scope) checkCommandNode(tree *parse.Tree, dot, prev types.Type, cmd *parse.CommandNode) (types.Type, error) {
	first := cmd.Args[0]
	switch n := first.(type) {
	case *parse.FieldNode:
		argTypes, err := s.argumentTypes(tree, dot, prev, cmd.Args[1:])
		if err != nil {
			return nil, err
		}
		return s.checkFieldNode(tree, dot, n, argTypes)
	case *parse.ChainNode:
		argTypes, err := s.argumentTypes(tree, dot, prev, cmd.Args[1:])
		if err != nil {
			return nil, err
		}
		return s.checkChainNode(tree, dot, prev, n, argTypes)
	case *parse.IdentifierNode:
		argTypes, err := s.argumentTypes(tree, dot, prev, cmd.Args[1:])
		if err != nil {
			return nil, err
		}
		tp, err := s.CallChecker.CheckCall(n.Ident, cmd.Args[1:], argTypes)
		if err != nil {
			return nil, s.error(tree, cmd, err)
		}
		return tp, nil
	case *parse.PipeNode:
		if err := s.notAFunction(cmd.Args, prev); err != nil {
			return nil, err
		}
		return s.checkPipeNode(tree, dot, n)
	case *parse.VariableNode:
		argTypes, err := s.argumentTypes(tree, dot, prev, cmd.Args[1:])
		if err != nil {
			return nil, err
		}
		return s.checkVariableNode(tree, n, argTypes)
	}

	if err := s.notAFunction(cmd.Args, prev); err != nil {
		return nil, err
	}

	switch n := first.(type) {
	case *parse.BoolNode:
		return types.Typ[types.UntypedBool], nil
	case *parse.StringNode:
		return types.Typ[types.UntypedString], nil
	case *parse.NumberNode:
		return newNumberNodeType(n)
	case *parse.DotNode:
		return dot, nil
	case *parse.NilNode:
		return nil, s.error(tree, n, fmt.Errorf("nil is not a command"))
	default:
		return nil, s.error(tree, first, fmt.Errorf("can't evaluate command %q", first))
	}
}

func (s *scope) argumentTypes(tree *parse.Tree, dot types.Type, prev types.Type, args []parse.Node) ([]types.Type, error) {
	argTypes := make([]types.Type, 0, len(args)+1)
	for _, arg := range args {
		argType, err := s.walk(tree, dot, prev, arg)
		if err != nil {
			return nil, err
		}
		argTypes = append(argTypes, argType)
	}
	if prev != nil {
		argTypes = append(argTypes, prev)
	}
	return argTypes, nil
}

func (s *scope) notAFunction(args []parse.Node, final types.Type) error {
	if len(args) > 1 || final != nil {
		return fmt.Errorf("can't give argument to non-function %s", args[0])
	}
	return nil
}

func (s *scope) checkIdentifiers(tree *parse.Tree, dot types.Type, n parse.Node, idents []string, args []types.Type) (types.Type, error) {
	x := dot
	for i, ident := range idents {
		x = dereference(x)
		switch xx := x.(type) {
		case *types.Map:
			switch key := xx.Key().Underlying().(type) {
			case *types.Basic:
				switch key.Kind() {
				// case types.Int, types.Int64, types.Int32, types.Int16, types.Int8,
				//	types.Uint, types.Uint64, types.Uint32, types.Uint16, types.Uint8:
				case types.Int:
					x = xx.Elem()
					_, err := strconv.Atoi(ident)
					if err != nil {
						loc, _ := tree.ErrorContext(n)
						return nil, fmt.Errorf(`%s: executing %q at <%s>: can't evaluate field one in type %s`, loc, tree.Name, n.String(), xx.String())
					}
				case types.String:
					x = xx.Elem()
				default:
				}
				continue
			default:
				x = xx.Elem()
			}
			continue
		default:
			if !token.IsExported(ident) {
				return nil, s.error(tree, n, fmt.Errorf("field or method %s is not exported", ident))
			}
			obj, _, _ := types.LookupFieldOrMethod(x, true, s.pkg, ident)
			if obj == nil {
				loc, _ := tree.ErrorContext(n)
				return nil, fmt.Errorf("type check failed: %s: %s not found on %s", loc, ident, x)
			}
			switch o := obj.(type) {
			default:
				x = obj.Type()
			case *types.Func:
				sig := o.Signature()
				resultLen := sig.Results().Len()
				if resultLen < 1 || resultLen > 2 {
					loc, _ := tree.ErrorContext(n)
					methodPos := s.fileSet.Position(o.Pos())
					return nil, fmt.Errorf("type check failed: %s: function %s has %d return values; should be 1 or 2: incorrect signature at %s", loc, ident, resultLen, methodPos)
				}
				if resultLen > 1 {
					loc, _ := tree.ErrorContext(n)
					methodPos := s.fileSet.Position(obj.Pos())
					finalResult := sig.Results().At(sig.Results().Len() - 1)
					errorType := types.Universe.Lookup("error")
					if !types.Identical(errorType.Type(), finalResult.Type()) {
						return nil, fmt.Errorf("type check failed: %s: invalid function signature for %s: second return value should be error; is %s: incorrect signature at %s", loc, ident, finalResult.Type(), methodPos)
					}
				}
				if i == len(idents)-1 {
					res, err := checkCallArguments(sig, args)
					if err != nil {
						return nil, s.error(tree, n, err)
					}
					return res, nil
				}
				x = sig.Results().At(0).Type()
			}
			if _, ok := x.(*types.Signature); ok && i < len(idents)-1 {
				return nil, s.error(tree, n, fmt.Errorf("identifier chain not supported for type %s", x.String()))
			}
		}
	}
	if len(args) > 0 {
		sig, ok := x.(*types.Signature)
		if !ok {
			return nil, s.error(tree, n, fmt.Errorf("expected method or function"))
		}
		tp, err := checkCallArguments(sig, args)
		if err != nil {
			return nil, s.error(tree, n, err)
		}
		return tp, nil
	}
	return x, nil
}

func (s *scope) error(tree *parse.Tree, n parse.Node, err error) error {
	loc, _ := tree.ErrorContext(n)
	return fmt.Errorf("type check failed: %s: executing %q at <%s>: %w", loc, tree.Name, n, err)
}

func (s *scope) checkRangeNode(tree *parse.Tree, dot types.Type, n *parse.RangeNode) error {
	child := s.child()
	pipeType, err := child.walk(tree, dot, nil, n.Pipe)
	if err != nil {
		return err
	}
	pipeType = dereference(pipeType).Underlying()
	var x types.Type
	switch pt := pipeType.(type) {
	case *types.Slice:
		x = pt.Elem()
		if len(n.Pipe.Decl) == 1 {
			child.variables[n.Pipe.Decl[0].Ident[0]] = x
		} else if len(n.Pipe.Decl) > 1 {
			child.variables[n.Pipe.Decl[0].Ident[0]] = types.Typ[types.Int]
			child.variables[n.Pipe.Decl[1].Ident[0]] = x
		}
	case *types.Array:
		x = pt.Elem()
		if len(n.Pipe.Decl) == 1 {
			child.variables[n.Pipe.Decl[0].Ident[0]] = x
		} else if len(n.Pipe.Decl) > 1 {
			child.variables[n.Pipe.Decl[0].Ident[0]] = types.Typ[types.Int]
			child.variables[n.Pipe.Decl[1].Ident[0]] = x
		}
	case *types.Map:
		x = pt.Elem()
		if len(n.Pipe.Decl) == 1 {
			child.variables[n.Pipe.Decl[0].Ident[0]] = pt.Elem()
		} else if len(n.Pipe.Decl) > 1 {
			child.variables[n.Pipe.Decl[0].Ident[0]] = pt.Key()
			child.variables[n.Pipe.Decl[1].Ident[0]] = pt.Elem()
		}
	case *types.Chan:
		x = pt.Elem()
		if len(n.Pipe.Decl) > 1 {
			child.variables[n.Pipe.Decl[0].Ident[0]] = types.Typ[types.Int] // TODO: this looks odd, I don't think I should permit an index here
			child.variables[n.Pipe.Decl[1].Ident[0]] = pt.Elem()
		}
	case *types.Basic:
		switch {
		case pt.Info()&types.IsInteger != 0:
			tp := types.Universe.Lookup(strings.TrimPrefix(pipeType.String(), "untyped "))
			x = tp.Type()
			if len(n.Pipe.Decl) > 0 {
				child.variables[n.Pipe.Decl[0].Ident[0]] = x
			}
			return nil
		default:
			return s.error(tree, n.Pipe, fmt.Errorf("range can't iterate over %s", strings.TrimPrefix(pipeType.String(), "untyped ")))
		}
	case *types.Signature:
		if v1, v2, ok := isIter2(pt); ok {
			x = v1
			if len(n.Pipe.Decl) > 0 {
				child.variables[n.Pipe.Decl[0].Ident[0]] = v1
			}
			if len(n.Pipe.Decl) > 1 {
				child.variables[n.Pipe.Decl[1].Ident[0]] = v2
			}
			return nil
		}
		if val, ok := isIter(pt); ok {
			x = val
			if len(n.Pipe.Decl) == 1 {
				child.variables[n.Pipe.Decl[0].Ident[0]] = val
			}
			if len(n.Pipe.Decl) > 1 {
				return s.error(tree, n.Pipe, fmt.Errorf("iter.Seq[T] must not iterate over more than one variable"))
			}
			return nil
		}
		return s.error(tree, n.Pipe, fmt.Errorf("failed to range over function %s", pipeType))
	default:
		return s.error(tree, n.Pipe, fmt.Errorf("failed to range over %s", pipeType))
	}
	if _, err := child.walk(tree, x, nil, n.List); err != nil {
		return err
	}
	if n.ElseList != nil {
		if _, err := child.walk(tree, x, nil, n.ElseList); err != nil {
			return err
		}
	}
	return nil
}

func isIter(signature *types.Signature) (types.Type, bool) {
	if signature == nil || signature.Variadic() || signature.Results().Len() != 0 || signature.Params().Len() != 1 {
		return nil, false
	}
	yield, ok := signature.Params().At(0).Type().(*types.Signature)
	if !ok || yield.Results().Len() != 1 || yield.Params().Len() != 1 {
		return nil, false
	}
	if !types.Identical(yield.Results().At(0).Type(), types.Universe.Lookup("bool").Type()) {
		return nil, false
	}
	return yield.Params().At(0).Type(), true
}

func isIter2(signature *types.Signature) (types.Type, types.Type, bool) {
	if signature == nil || signature.Variadic() || signature.Results().Len() != 0 || signature.Params().Len() != 1 {
		return nil, nil, false
	}
	yield, ok := signature.Params().At(0).Type().(*types.Signature)
	if !ok || yield.Results().Len() != 1 || yield.Params().Len() != 2 {
		return nil, nil, false
	}
	if !types.Identical(yield.Results().At(0).Type(), types.Universe.Lookup("bool").Type()) {
		return nil, nil, false
	}
	yp := yield.Params()
	return yp.At(0).Type(), yp.At(1).Type(), true
}

func (s *scope) checkIdentifierNode(tree *parse.Tree, n *parse.IdentifierNode) (types.Type, error) {
	if !strings.HasPrefix(n.Ident, "$") {
		tp, err := s.CheckCall(n.Ident, nil, nil)
		if err != nil {
			return nil, s.error(tree, n, err)
		}
		return tp, err
	}
	tp, ok := s.variables[n.Ident]
	if !ok {
		return nil, fmt.Errorf("failed to find identifier %s", n.Ident)
	}
	return tp, nil
}

func dereference(tp types.Type) types.Type {
	for {
		ptr, ok := tp.(*types.Pointer)
		if !ok {
			return tp
		}
		tp = ptr.Elem()
	}
}
