package domtest

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/crhntr/dom"
	"github.com/crhntr/dom/spec"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate --fake-name TestingT -o ../internal/fakes/t.go . TestingT

type TestingT interface {
	Helper()
	Error(...any)
	Log(...any)
	Errorf(format string, args ...interface{})
	FailNow()
	Failed() bool
	SkipNow()
}

func ParseResponseDocument(t TestingT, res *http.Response) spec.Document {
	t.Helper()
	buf, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
		if err := res.Body.Close(); err != nil {
			t.Error(err)
		}
		return nil
	}
	if err := res.Body.Close(); err != nil {
		t.Error(err)
		return nil
	}
	return ParseReaderDocument(t, bytes.NewReader(buf))
}

func ParseStringDocument(t TestingT, s string) spec.Document {
	t.Helper()
	return ParseReaderDocument(t, strings.NewReader(s))
}

func ParseReaderDocument(t TestingT, r io.Reader) spec.Document {
	t.Helper()
	node, err := html.Parse(r)
	if err != nil {
		t.Error(err)
		return nil
	}
	return dom.NewNode(node).(spec.Document)
}

func ParseReaderDocumentFragment(t TestingT, r io.Reader, parent atom.Atom) spec.DocumentFragment {
	t.Helper()

	body, err := io.ReadAll(r)
	if err != nil {
		t.Error(err)
		return nil
	}
	nodes, err := html.ParseFragment(bytes.NewReader(body), &html.Node{
		Type:     html.ElementNode,
		Data:     parent.String(),
		DataAtom: parent,
	})
	if err != nil {
		t.Error(err)
		return nil
	}
	return dom.NewDocumentFragment(nodes)
}

func ParseResponseDocumentFragment(t TestingT, res *http.Response, parent atom.Atom) spec.DocumentFragment {
	t.Helper()
	defer closeAndCheckError(t, res.Body)
	return ParseReaderDocumentFragment(t, res.Body, parent)
}

func closeAndCheckError(t TestingT, c io.Closer) {
	if err := c.Close(); err != nil {
		t.Error(err)
	}
}
