// Code generated by counterfeiter. DO NOT EDIT.
package fake

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"sync"

	"github.com/crhntr/muxt/internal/example"
)

type Receiver struct {
	CheckAuthStub        func(*http.Request) (string, error)
	checkAuthMutex       sync.RWMutex
	checkAuthArgsForCall []struct {
		arg1 *http.Request
	}
	checkAuthReturns struct {
		result1 string
		result2 error
	}
	checkAuthReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	GetCommentStub        func(context.Context, int, int) (string, error)
	getCommentMutex       sync.RWMutex
	getCommentArgsForCall []struct {
		arg1 context.Context
		arg2 int
		arg3 int
	}
	getCommentReturns struct {
		result1 string
		result2 error
	}
	getCommentReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	HandlerStub        func(http.ResponseWriter, *http.Request) template.HTML
	handlerMutex       sync.RWMutex
	handlerArgsForCall []struct {
		arg1 http.ResponseWriter
		arg2 *http.Request
	}
	handlerReturns struct {
		result1 template.HTML
	}
	handlerReturnsOnCall map[int]struct {
		result1 template.HTML
	}
	ListArticlesStub        func(context.Context) ([]example.Article, error)
	listArticlesMutex       sync.RWMutex
	listArticlesArgsForCall []struct {
		arg1 context.Context
	}
	listArticlesReturns struct {
		result1 []example.Article
		result2 error
	}
	listArticlesReturnsOnCall map[int]struct {
		result1 []example.Article
		result2 error
	}
	LogLinesStub        func(*slog.Logger) int
	logLinesMutex       sync.RWMutex
	logLinesArgsForCall []struct {
		arg1 *slog.Logger
	}
	logLinesReturns struct {
		result1 int
	}
	logLinesReturnsOnCall map[int]struct {
		result1 int
	}
	NumAuthorsStub        func() int
	numAuthorsMutex       sync.RWMutex
	numAuthorsArgsForCall []struct {
	}
	numAuthorsReturns struct {
		result1 int
	}
	numAuthorsReturnsOnCall map[int]struct {
		result1 int
	}
	ParseStub        func(string) []string
	parseMutex       sync.RWMutex
	parseArgsForCall []struct {
		arg1 string
	}
	parseReturns struct {
		result1 []string
	}
	parseReturnsOnCall map[int]struct {
		result1 []string
	}
	SomeStringStub        func(context.Context, string) (string, error)
	someStringMutex       sync.RWMutex
	someStringArgsForCall []struct {
		arg1 context.Context
		arg2 string
	}
	someStringReturns struct {
		result1 string
		result2 error
	}
	someStringReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	TemplateStub        func(*template.Template) template.HTML
	templateMutex       sync.RWMutex
	templateArgsForCall []struct {
		arg1 *template.Template
	}
	templateReturns struct {
		result1 template.HTML
	}
	templateReturnsOnCall map[int]struct {
		result1 template.HTML
	}
	ToUpperStub        func(...rune) string
	toUpperMutex       sync.RWMutex
	toUpperArgsForCall []struct {
		arg1 []rune
	}
	toUpperReturns struct {
		result1 string
	}
	toUpperReturnsOnCall map[int]struct {
		result1 string
	}
	TooManyResultsStub        func() (int, int, int)
	tooManyResultsMutex       sync.RWMutex
	tooManyResultsArgsForCall []struct {
	}
	tooManyResultsReturns struct {
		result1 int
		result2 int
		result3 int
	}
	tooManyResultsReturnsOnCall map[int]struct {
		result1 int
		result2 int
		result3 int
	}
	TupleStub        func() (string, string)
	tupleMutex       sync.RWMutex
	tupleArgsForCall []struct {
	}
	tupleReturns struct {
		result1 string
		result2 string
	}
	tupleReturnsOnCall map[int]struct {
		result1 string
		result2 string
	}
	TypeStub        func(any) string
	typeMutex       sync.RWMutex
	typeArgsForCall []struct {
		arg1 any
	}
	typeReturns struct {
		result1 string
	}
	typeReturnsOnCall map[int]struct {
		result1 string
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *Receiver) CheckAuth(arg1 *http.Request) (string, error) {
	fake.checkAuthMutex.Lock()
	ret, specificReturn := fake.checkAuthReturnsOnCall[len(fake.checkAuthArgsForCall)]
	fake.checkAuthArgsForCall = append(fake.checkAuthArgsForCall, struct {
		arg1 *http.Request
	}{arg1})
	stub := fake.CheckAuthStub
	fakeReturns := fake.checkAuthReturns
	fake.recordInvocation("CheckAuth", []interface{}{arg1})
	fake.checkAuthMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *Receiver) CheckAuthCallCount() int {
	fake.checkAuthMutex.RLock()
	defer fake.checkAuthMutex.RUnlock()
	return len(fake.checkAuthArgsForCall)
}

func (fake *Receiver) CheckAuthCalls(stub func(*http.Request) (string, error)) {
	fake.checkAuthMutex.Lock()
	defer fake.checkAuthMutex.Unlock()
	fake.CheckAuthStub = stub
}

func (fake *Receiver) CheckAuthArgsForCall(i int) *http.Request {
	fake.checkAuthMutex.RLock()
	defer fake.checkAuthMutex.RUnlock()
	argsForCall := fake.checkAuthArgsForCall[i]
	return argsForCall.arg1
}

func (fake *Receiver) CheckAuthReturns(result1 string, result2 error) {
	fake.checkAuthMutex.Lock()
	defer fake.checkAuthMutex.Unlock()
	fake.CheckAuthStub = nil
	fake.checkAuthReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *Receiver) CheckAuthReturnsOnCall(i int, result1 string, result2 error) {
	fake.checkAuthMutex.Lock()
	defer fake.checkAuthMutex.Unlock()
	fake.CheckAuthStub = nil
	if fake.checkAuthReturnsOnCall == nil {
		fake.checkAuthReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.checkAuthReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *Receiver) GetComment(arg1 context.Context, arg2 int, arg3 int) (string, error) {
	fake.getCommentMutex.Lock()
	ret, specificReturn := fake.getCommentReturnsOnCall[len(fake.getCommentArgsForCall)]
	fake.getCommentArgsForCall = append(fake.getCommentArgsForCall, struct {
		arg1 context.Context
		arg2 int
		arg3 int
	}{arg1, arg2, arg3})
	stub := fake.GetCommentStub
	fakeReturns := fake.getCommentReturns
	fake.recordInvocation("GetComment", []interface{}{arg1, arg2, arg3})
	fake.getCommentMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *Receiver) GetCommentCallCount() int {
	fake.getCommentMutex.RLock()
	defer fake.getCommentMutex.RUnlock()
	return len(fake.getCommentArgsForCall)
}

func (fake *Receiver) GetCommentCalls(stub func(context.Context, int, int) (string, error)) {
	fake.getCommentMutex.Lock()
	defer fake.getCommentMutex.Unlock()
	fake.GetCommentStub = stub
}

func (fake *Receiver) GetCommentArgsForCall(i int) (context.Context, int, int) {
	fake.getCommentMutex.RLock()
	defer fake.getCommentMutex.RUnlock()
	argsForCall := fake.getCommentArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *Receiver) GetCommentReturns(result1 string, result2 error) {
	fake.getCommentMutex.Lock()
	defer fake.getCommentMutex.Unlock()
	fake.GetCommentStub = nil
	fake.getCommentReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *Receiver) GetCommentReturnsOnCall(i int, result1 string, result2 error) {
	fake.getCommentMutex.Lock()
	defer fake.getCommentMutex.Unlock()
	fake.GetCommentStub = nil
	if fake.getCommentReturnsOnCall == nil {
		fake.getCommentReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.getCommentReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *Receiver) Handler(arg1 http.ResponseWriter, arg2 *http.Request) template.HTML {
	fake.handlerMutex.Lock()
	ret, specificReturn := fake.handlerReturnsOnCall[len(fake.handlerArgsForCall)]
	fake.handlerArgsForCall = append(fake.handlerArgsForCall, struct {
		arg1 http.ResponseWriter
		arg2 *http.Request
	}{arg1, arg2})
	stub := fake.HandlerStub
	fakeReturns := fake.handlerReturns
	fake.recordInvocation("Handler", []interface{}{arg1, arg2})
	fake.handlerMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *Receiver) HandlerCallCount() int {
	fake.handlerMutex.RLock()
	defer fake.handlerMutex.RUnlock()
	return len(fake.handlerArgsForCall)
}

func (fake *Receiver) HandlerCalls(stub func(http.ResponseWriter, *http.Request) template.HTML) {
	fake.handlerMutex.Lock()
	defer fake.handlerMutex.Unlock()
	fake.HandlerStub = stub
}

func (fake *Receiver) HandlerArgsForCall(i int) (http.ResponseWriter, *http.Request) {
	fake.handlerMutex.RLock()
	defer fake.handlerMutex.RUnlock()
	argsForCall := fake.handlerArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *Receiver) HandlerReturns(result1 template.HTML) {
	fake.handlerMutex.Lock()
	defer fake.handlerMutex.Unlock()
	fake.HandlerStub = nil
	fake.handlerReturns = struct {
		result1 template.HTML
	}{result1}
}

func (fake *Receiver) HandlerReturnsOnCall(i int, result1 template.HTML) {
	fake.handlerMutex.Lock()
	defer fake.handlerMutex.Unlock()
	fake.HandlerStub = nil
	if fake.handlerReturnsOnCall == nil {
		fake.handlerReturnsOnCall = make(map[int]struct {
			result1 template.HTML
		})
	}
	fake.handlerReturnsOnCall[i] = struct {
		result1 template.HTML
	}{result1}
}

func (fake *Receiver) ListArticles(arg1 context.Context) ([]example.Article, error) {
	fake.listArticlesMutex.Lock()
	ret, specificReturn := fake.listArticlesReturnsOnCall[len(fake.listArticlesArgsForCall)]
	fake.listArticlesArgsForCall = append(fake.listArticlesArgsForCall, struct {
		arg1 context.Context
	}{arg1})
	stub := fake.ListArticlesStub
	fakeReturns := fake.listArticlesReturns
	fake.recordInvocation("ListArticles", []interface{}{arg1})
	fake.listArticlesMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *Receiver) ListArticlesCallCount() int {
	fake.listArticlesMutex.RLock()
	defer fake.listArticlesMutex.RUnlock()
	return len(fake.listArticlesArgsForCall)
}

func (fake *Receiver) ListArticlesCalls(stub func(context.Context) ([]example.Article, error)) {
	fake.listArticlesMutex.Lock()
	defer fake.listArticlesMutex.Unlock()
	fake.ListArticlesStub = stub
}

func (fake *Receiver) ListArticlesArgsForCall(i int) context.Context {
	fake.listArticlesMutex.RLock()
	defer fake.listArticlesMutex.RUnlock()
	argsForCall := fake.listArticlesArgsForCall[i]
	return argsForCall.arg1
}

func (fake *Receiver) ListArticlesReturns(result1 []example.Article, result2 error) {
	fake.listArticlesMutex.Lock()
	defer fake.listArticlesMutex.Unlock()
	fake.ListArticlesStub = nil
	fake.listArticlesReturns = struct {
		result1 []example.Article
		result2 error
	}{result1, result2}
}

func (fake *Receiver) ListArticlesReturnsOnCall(i int, result1 []example.Article, result2 error) {
	fake.listArticlesMutex.Lock()
	defer fake.listArticlesMutex.Unlock()
	fake.ListArticlesStub = nil
	if fake.listArticlesReturnsOnCall == nil {
		fake.listArticlesReturnsOnCall = make(map[int]struct {
			result1 []example.Article
			result2 error
		})
	}
	fake.listArticlesReturnsOnCall[i] = struct {
		result1 []example.Article
		result2 error
	}{result1, result2}
}

func (fake *Receiver) LogLines(arg1 *slog.Logger) int {
	fake.logLinesMutex.Lock()
	ret, specificReturn := fake.logLinesReturnsOnCall[len(fake.logLinesArgsForCall)]
	fake.logLinesArgsForCall = append(fake.logLinesArgsForCall, struct {
		arg1 *slog.Logger
	}{arg1})
	stub := fake.LogLinesStub
	fakeReturns := fake.logLinesReturns
	fake.recordInvocation("LogLines", []interface{}{arg1})
	fake.logLinesMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *Receiver) LogLinesCallCount() int {
	fake.logLinesMutex.RLock()
	defer fake.logLinesMutex.RUnlock()
	return len(fake.logLinesArgsForCall)
}

func (fake *Receiver) LogLinesCalls(stub func(*slog.Logger) int) {
	fake.logLinesMutex.Lock()
	defer fake.logLinesMutex.Unlock()
	fake.LogLinesStub = stub
}

func (fake *Receiver) LogLinesArgsForCall(i int) *slog.Logger {
	fake.logLinesMutex.RLock()
	defer fake.logLinesMutex.RUnlock()
	argsForCall := fake.logLinesArgsForCall[i]
	return argsForCall.arg1
}

func (fake *Receiver) LogLinesReturns(result1 int) {
	fake.logLinesMutex.Lock()
	defer fake.logLinesMutex.Unlock()
	fake.LogLinesStub = nil
	fake.logLinesReturns = struct {
		result1 int
	}{result1}
}

func (fake *Receiver) LogLinesReturnsOnCall(i int, result1 int) {
	fake.logLinesMutex.Lock()
	defer fake.logLinesMutex.Unlock()
	fake.LogLinesStub = nil
	if fake.logLinesReturnsOnCall == nil {
		fake.logLinesReturnsOnCall = make(map[int]struct {
			result1 int
		})
	}
	fake.logLinesReturnsOnCall[i] = struct {
		result1 int
	}{result1}
}

func (fake *Receiver) NumAuthors() int {
	fake.numAuthorsMutex.Lock()
	ret, specificReturn := fake.numAuthorsReturnsOnCall[len(fake.numAuthorsArgsForCall)]
	fake.numAuthorsArgsForCall = append(fake.numAuthorsArgsForCall, struct {
	}{})
	stub := fake.NumAuthorsStub
	fakeReturns := fake.numAuthorsReturns
	fake.recordInvocation("NumAuthors", []interface{}{})
	fake.numAuthorsMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *Receiver) NumAuthorsCallCount() int {
	fake.numAuthorsMutex.RLock()
	defer fake.numAuthorsMutex.RUnlock()
	return len(fake.numAuthorsArgsForCall)
}

func (fake *Receiver) NumAuthorsCalls(stub func() int) {
	fake.numAuthorsMutex.Lock()
	defer fake.numAuthorsMutex.Unlock()
	fake.NumAuthorsStub = stub
}

func (fake *Receiver) NumAuthorsReturns(result1 int) {
	fake.numAuthorsMutex.Lock()
	defer fake.numAuthorsMutex.Unlock()
	fake.NumAuthorsStub = nil
	fake.numAuthorsReturns = struct {
		result1 int
	}{result1}
}

func (fake *Receiver) NumAuthorsReturnsOnCall(i int, result1 int) {
	fake.numAuthorsMutex.Lock()
	defer fake.numAuthorsMutex.Unlock()
	fake.NumAuthorsStub = nil
	if fake.numAuthorsReturnsOnCall == nil {
		fake.numAuthorsReturnsOnCall = make(map[int]struct {
			result1 int
		})
	}
	fake.numAuthorsReturnsOnCall[i] = struct {
		result1 int
	}{result1}
}

func (fake *Receiver) Parse(arg1 string) []string {
	fake.parseMutex.Lock()
	ret, specificReturn := fake.parseReturnsOnCall[len(fake.parseArgsForCall)]
	fake.parseArgsForCall = append(fake.parseArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.ParseStub
	fakeReturns := fake.parseReturns
	fake.recordInvocation("Parse", []interface{}{arg1})
	fake.parseMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *Receiver) ParseCallCount() int {
	fake.parseMutex.RLock()
	defer fake.parseMutex.RUnlock()
	return len(fake.parseArgsForCall)
}

func (fake *Receiver) ParseCalls(stub func(string) []string) {
	fake.parseMutex.Lock()
	defer fake.parseMutex.Unlock()
	fake.ParseStub = stub
}

func (fake *Receiver) ParseArgsForCall(i int) string {
	fake.parseMutex.RLock()
	defer fake.parseMutex.RUnlock()
	argsForCall := fake.parseArgsForCall[i]
	return argsForCall.arg1
}

func (fake *Receiver) ParseReturns(result1 []string) {
	fake.parseMutex.Lock()
	defer fake.parseMutex.Unlock()
	fake.ParseStub = nil
	fake.parseReturns = struct {
		result1 []string
	}{result1}
}

func (fake *Receiver) ParseReturnsOnCall(i int, result1 []string) {
	fake.parseMutex.Lock()
	defer fake.parseMutex.Unlock()
	fake.ParseStub = nil
	if fake.parseReturnsOnCall == nil {
		fake.parseReturnsOnCall = make(map[int]struct {
			result1 []string
		})
	}
	fake.parseReturnsOnCall[i] = struct {
		result1 []string
	}{result1}
}

func (fake *Receiver) SomeString(arg1 context.Context, arg2 string) (string, error) {
	fake.someStringMutex.Lock()
	ret, specificReturn := fake.someStringReturnsOnCall[len(fake.someStringArgsForCall)]
	fake.someStringArgsForCall = append(fake.someStringArgsForCall, struct {
		arg1 context.Context
		arg2 string
	}{arg1, arg2})
	stub := fake.SomeStringStub
	fakeReturns := fake.someStringReturns
	fake.recordInvocation("SomeString", []interface{}{arg1, arg2})
	fake.someStringMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *Receiver) SomeStringCallCount() int {
	fake.someStringMutex.RLock()
	defer fake.someStringMutex.RUnlock()
	return len(fake.someStringArgsForCall)
}

func (fake *Receiver) SomeStringCalls(stub func(context.Context, string) (string, error)) {
	fake.someStringMutex.Lock()
	defer fake.someStringMutex.Unlock()
	fake.SomeStringStub = stub
}

func (fake *Receiver) SomeStringArgsForCall(i int) (context.Context, string) {
	fake.someStringMutex.RLock()
	defer fake.someStringMutex.RUnlock()
	argsForCall := fake.someStringArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *Receiver) SomeStringReturns(result1 string, result2 error) {
	fake.someStringMutex.Lock()
	defer fake.someStringMutex.Unlock()
	fake.SomeStringStub = nil
	fake.someStringReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *Receiver) SomeStringReturnsOnCall(i int, result1 string, result2 error) {
	fake.someStringMutex.Lock()
	defer fake.someStringMutex.Unlock()
	fake.SomeStringStub = nil
	if fake.someStringReturnsOnCall == nil {
		fake.someStringReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.someStringReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *Receiver) Template(arg1 *template.Template) template.HTML {
	fake.templateMutex.Lock()
	ret, specificReturn := fake.templateReturnsOnCall[len(fake.templateArgsForCall)]
	fake.templateArgsForCall = append(fake.templateArgsForCall, struct {
		arg1 *template.Template
	}{arg1})
	stub := fake.TemplateStub
	fakeReturns := fake.templateReturns
	fake.recordInvocation("Template", []interface{}{arg1})
	fake.templateMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *Receiver) TemplateCallCount() int {
	fake.templateMutex.RLock()
	defer fake.templateMutex.RUnlock()
	return len(fake.templateArgsForCall)
}

func (fake *Receiver) TemplateCalls(stub func(*template.Template) template.HTML) {
	fake.templateMutex.Lock()
	defer fake.templateMutex.Unlock()
	fake.TemplateStub = stub
}

func (fake *Receiver) TemplateArgsForCall(i int) *template.Template {
	fake.templateMutex.RLock()
	defer fake.templateMutex.RUnlock()
	argsForCall := fake.templateArgsForCall[i]
	return argsForCall.arg1
}

func (fake *Receiver) TemplateReturns(result1 template.HTML) {
	fake.templateMutex.Lock()
	defer fake.templateMutex.Unlock()
	fake.TemplateStub = nil
	fake.templateReturns = struct {
		result1 template.HTML
	}{result1}
}

func (fake *Receiver) TemplateReturnsOnCall(i int, result1 template.HTML) {
	fake.templateMutex.Lock()
	defer fake.templateMutex.Unlock()
	fake.TemplateStub = nil
	if fake.templateReturnsOnCall == nil {
		fake.templateReturnsOnCall = make(map[int]struct {
			result1 template.HTML
		})
	}
	fake.templateReturnsOnCall[i] = struct {
		result1 template.HTML
	}{result1}
}

func (fake *Receiver) ToUpper(arg1 ...rune) string {
	fake.toUpperMutex.Lock()
	ret, specificReturn := fake.toUpperReturnsOnCall[len(fake.toUpperArgsForCall)]
	fake.toUpperArgsForCall = append(fake.toUpperArgsForCall, struct {
		arg1 []rune
	}{arg1})
	stub := fake.ToUpperStub
	fakeReturns := fake.toUpperReturns
	fake.recordInvocation("ToUpper", []interface{}{arg1})
	fake.toUpperMutex.Unlock()
	if stub != nil {
		return stub(arg1...)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *Receiver) ToUpperCallCount() int {
	fake.toUpperMutex.RLock()
	defer fake.toUpperMutex.RUnlock()
	return len(fake.toUpperArgsForCall)
}

func (fake *Receiver) ToUpperCalls(stub func(...rune) string) {
	fake.toUpperMutex.Lock()
	defer fake.toUpperMutex.Unlock()
	fake.ToUpperStub = stub
}

func (fake *Receiver) ToUpperArgsForCall(i int) []rune {
	fake.toUpperMutex.RLock()
	defer fake.toUpperMutex.RUnlock()
	argsForCall := fake.toUpperArgsForCall[i]
	return argsForCall.arg1
}

func (fake *Receiver) ToUpperReturns(result1 string) {
	fake.toUpperMutex.Lock()
	defer fake.toUpperMutex.Unlock()
	fake.ToUpperStub = nil
	fake.toUpperReturns = struct {
		result1 string
	}{result1}
}

func (fake *Receiver) ToUpperReturnsOnCall(i int, result1 string) {
	fake.toUpperMutex.Lock()
	defer fake.toUpperMutex.Unlock()
	fake.ToUpperStub = nil
	if fake.toUpperReturnsOnCall == nil {
		fake.toUpperReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.toUpperReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *Receiver) TooManyResults() (int, int, int) {
	fake.tooManyResultsMutex.Lock()
	ret, specificReturn := fake.tooManyResultsReturnsOnCall[len(fake.tooManyResultsArgsForCall)]
	fake.tooManyResultsArgsForCall = append(fake.tooManyResultsArgsForCall, struct {
	}{})
	stub := fake.TooManyResultsStub
	fakeReturns := fake.tooManyResultsReturns
	fake.recordInvocation("TooManyResults", []interface{}{})
	fake.tooManyResultsMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3
}

func (fake *Receiver) TooManyResultsCallCount() int {
	fake.tooManyResultsMutex.RLock()
	defer fake.tooManyResultsMutex.RUnlock()
	return len(fake.tooManyResultsArgsForCall)
}

func (fake *Receiver) TooManyResultsCalls(stub func() (int, int, int)) {
	fake.tooManyResultsMutex.Lock()
	defer fake.tooManyResultsMutex.Unlock()
	fake.TooManyResultsStub = stub
}

func (fake *Receiver) TooManyResultsReturns(result1 int, result2 int, result3 int) {
	fake.tooManyResultsMutex.Lock()
	defer fake.tooManyResultsMutex.Unlock()
	fake.TooManyResultsStub = nil
	fake.tooManyResultsReturns = struct {
		result1 int
		result2 int
		result3 int
	}{result1, result2, result3}
}

func (fake *Receiver) TooManyResultsReturnsOnCall(i int, result1 int, result2 int, result3 int) {
	fake.tooManyResultsMutex.Lock()
	defer fake.tooManyResultsMutex.Unlock()
	fake.TooManyResultsStub = nil
	if fake.tooManyResultsReturnsOnCall == nil {
		fake.tooManyResultsReturnsOnCall = make(map[int]struct {
			result1 int
			result2 int
			result3 int
		})
	}
	fake.tooManyResultsReturnsOnCall[i] = struct {
		result1 int
		result2 int
		result3 int
	}{result1, result2, result3}
}

func (fake *Receiver) Tuple() (string, string) {
	fake.tupleMutex.Lock()
	ret, specificReturn := fake.tupleReturnsOnCall[len(fake.tupleArgsForCall)]
	fake.tupleArgsForCall = append(fake.tupleArgsForCall, struct {
	}{})
	stub := fake.TupleStub
	fakeReturns := fake.tupleReturns
	fake.recordInvocation("Tuple", []interface{}{})
	fake.tupleMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *Receiver) TupleCallCount() int {
	fake.tupleMutex.RLock()
	defer fake.tupleMutex.RUnlock()
	return len(fake.tupleArgsForCall)
}

func (fake *Receiver) TupleCalls(stub func() (string, string)) {
	fake.tupleMutex.Lock()
	defer fake.tupleMutex.Unlock()
	fake.TupleStub = stub
}

func (fake *Receiver) TupleReturns(result1 string, result2 string) {
	fake.tupleMutex.Lock()
	defer fake.tupleMutex.Unlock()
	fake.TupleStub = nil
	fake.tupleReturns = struct {
		result1 string
		result2 string
	}{result1, result2}
}

func (fake *Receiver) TupleReturnsOnCall(i int, result1 string, result2 string) {
	fake.tupleMutex.Lock()
	defer fake.tupleMutex.Unlock()
	fake.TupleStub = nil
	if fake.tupleReturnsOnCall == nil {
		fake.tupleReturnsOnCall = make(map[int]struct {
			result1 string
			result2 string
		})
	}
	fake.tupleReturnsOnCall[i] = struct {
		result1 string
		result2 string
	}{result1, result2}
}

func (fake *Receiver) Type(arg1 any) string {
	fake.typeMutex.Lock()
	ret, specificReturn := fake.typeReturnsOnCall[len(fake.typeArgsForCall)]
	fake.typeArgsForCall = append(fake.typeArgsForCall, struct {
		arg1 any
	}{arg1})
	stub := fake.TypeStub
	fakeReturns := fake.typeReturns
	fake.recordInvocation("Type", []interface{}{arg1})
	fake.typeMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *Receiver) TypeCallCount() int {
	fake.typeMutex.RLock()
	defer fake.typeMutex.RUnlock()
	return len(fake.typeArgsForCall)
}

func (fake *Receiver) TypeCalls(stub func(any) string) {
	fake.typeMutex.Lock()
	defer fake.typeMutex.Unlock()
	fake.TypeStub = stub
}

func (fake *Receiver) TypeArgsForCall(i int) any {
	fake.typeMutex.RLock()
	defer fake.typeMutex.RUnlock()
	argsForCall := fake.typeArgsForCall[i]
	return argsForCall.arg1
}

func (fake *Receiver) TypeReturns(result1 string) {
	fake.typeMutex.Lock()
	defer fake.typeMutex.Unlock()
	fake.TypeStub = nil
	fake.typeReturns = struct {
		result1 string
	}{result1}
}

func (fake *Receiver) TypeReturnsOnCall(i int, result1 string) {
	fake.typeMutex.Lock()
	defer fake.typeMutex.Unlock()
	fake.TypeStub = nil
	if fake.typeReturnsOnCall == nil {
		fake.typeReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.typeReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *Receiver) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.checkAuthMutex.RLock()
	defer fake.checkAuthMutex.RUnlock()
	fake.getCommentMutex.RLock()
	defer fake.getCommentMutex.RUnlock()
	fake.handlerMutex.RLock()
	defer fake.handlerMutex.RUnlock()
	fake.listArticlesMutex.RLock()
	defer fake.listArticlesMutex.RUnlock()
	fake.logLinesMutex.RLock()
	defer fake.logLinesMutex.RUnlock()
	fake.numAuthorsMutex.RLock()
	defer fake.numAuthorsMutex.RUnlock()
	fake.parseMutex.RLock()
	defer fake.parseMutex.RUnlock()
	fake.someStringMutex.RLock()
	defer fake.someStringMutex.RUnlock()
	fake.templateMutex.RLock()
	defer fake.templateMutex.RUnlock()
	fake.toUpperMutex.RLock()
	defer fake.toUpperMutex.RUnlock()
	fake.tooManyResultsMutex.RLock()
	defer fake.tooManyResultsMutex.RUnlock()
	fake.tupleMutex.RLock()
	defer fake.tupleMutex.RUnlock()
	fake.typeMutex.RLock()
	defer fake.typeMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *Receiver) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}