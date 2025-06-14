// Code generated by muxt. DO NOT EDIT.
// muxt version: (devel)
//
// MIT License
//
// Copyright (c) 2025 Christopher Hunter
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package hypertext

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"strconv"
)

type RoutesReceiver interface {
	SubmitFormEditRow(fruitID int, form EditRow) (Row, error)
	GetFormEditRow(fruitID int) (Row, error)
	List(_ context.Context) []Row
}

func TemplateRoutes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("PATCH /fruits/{id}", func(response http.ResponseWriter, request *http.Request) {
		idParsed, err := strconv.Atoi(request.PathValue("id"))
		if err != nil {
			var zv Row
			rd := newTemplateData(receiver, response, request, zv, false, err)
			buf := bytes.NewBuffer(nil)
			if err := templates.ExecuteTemplate(buf, "PATCH /fruits/{id} SubmitFormEditRow(id, form)", rd); err != nil {
				slog.ErrorContext(request.Context(), "failed to render page", slog.String("path", request.URL.Path), slog.String("pattern", request.Pattern), slog.String("error", err.Error()))
				http.Error(response, "failed to render page", http.StatusInternalServerError)
				return
			}
			sc := cmp.Or(rd.statusCode, http.StatusBadRequest)
			if rd.redirectURL != "" {
				http.Redirect(response, request, rd.redirectURL, sc)
				return
			}
			if contentType := response.Header().Get("content-type"); contentType == "" {
				response.Header().Set("content-type", "text/html; charset=utf-8")
			}
			response.Header().Set("content-length", strconv.Itoa(buf.Len()))
			response.WriteHeader(sc)
			_, _ = buf.WriteTo(response)
			return
		}
		id := idParsed
		request.ParseForm()
		var form EditRow
		{
			value, err := strconv.Atoi(request.FormValue("count"))
			if err != nil {
				var zv Row
				rd := newTemplateData(receiver, response, request, zv, false, err)
				buf := bytes.NewBuffer(nil)
				if err := templates.ExecuteTemplate(buf, "PATCH /fruits/{id} SubmitFormEditRow(id, form)", rd); err != nil {
					slog.ErrorContext(request.Context(), "failed to render page", slog.String("path", request.URL.Path), slog.String("pattern", request.Pattern), slog.String("error", err.Error()))
					http.Error(response, "failed to render page", http.StatusInternalServerError)
					return
				}
				sc := cmp.Or(rd.statusCode, http.StatusBadRequest)
				if rd.redirectURL != "" {
					http.Redirect(response, request, rd.redirectURL, sc)
					return
				}
				if contentType := response.Header().Get("content-type"); contentType == "" {
					response.Header().Set("content-type", "text/html; charset=utf-8")
				}
				response.Header().Set("content-length", strconv.Itoa(buf.Len()))
				response.WriteHeader(sc)
				_, _ = buf.WriteTo(response)
				return
			}
			if value < 0 {
				var zv Row
				rd := newTemplateData(receiver, response, request, zv, false, errors.New("count must not be less than 0"))
				buf := bytes.NewBuffer(nil)
				if err := templates.ExecuteTemplate(buf, "PATCH /fruits/{id} SubmitFormEditRow(id, form)", rd); err != nil {
					slog.ErrorContext(request.Context(), "failed to render page", slog.String("path", request.URL.Path), slog.String("pattern", request.Pattern), slog.String("error", err.Error()))
					http.Error(response, "failed to render page", http.StatusInternalServerError)
					return
				}
				sc := cmp.Or(rd.statusCode, http.StatusBadRequest)
				if rd.redirectURL != "" {
					http.Redirect(response, request, rd.redirectURL, sc)
					return
				}
				if contentType := response.Header().Get("content-type"); contentType == "" {
					response.Header().Set("content-type", "text/html; charset=utf-8")
				}
				response.Header().Set("content-length", strconv.Itoa(buf.Len()))
				response.WriteHeader(sc)
				_, _ = buf.WriteTo(response)
				return
			}
			form.Value = value
		}
		result, err := receiver.SubmitFormEditRow(id, form)
		if err != nil {
			var zv Row
			rd := newTemplateData(receiver, response, request, zv, false, err)
			buf := bytes.NewBuffer(nil)
			if err := templates.ExecuteTemplate(buf, "PATCH /fruits/{id} SubmitFormEditRow(id, form)", rd); err != nil {
				slog.ErrorContext(request.Context(), "failed to render page", slog.String("path", request.URL.Path), slog.String("pattern", request.Pattern), slog.String("error", err.Error()))
				http.Error(response, "failed to render page", http.StatusInternalServerError)
				return
			}
			sc := cmp.Or(rd.statusCode, http.StatusInternalServerError)
			if rd.redirectURL != "" {
				http.Redirect(response, request, rd.redirectURL, sc)
				return
			}
			if contentType := response.Header().Get("content-type"); contentType == "" {
				response.Header().Set("content-type", "text/html; charset=utf-8")
			}
			response.Header().Set("content-length", strconv.Itoa(buf.Len()))
			response.WriteHeader(sc)
			_, _ = buf.WriteTo(response)
			return
		}
		td := newTemplateData(receiver, response, request, result, true, nil)
		buf := bytes.NewBuffer(nil)
		if err := templates.ExecuteTemplate(buf, "PATCH /fruits/{id} SubmitFormEditRow(id, form)", td); err != nil {
			slog.ErrorContext(request.Context(), "failed to render page", slog.String("path", request.URL.Path), slog.String("pattern", request.Pattern), slog.String("error", err.Error()))
			http.Error(response, "failed to render page", http.StatusInternalServerError)
			return
		}
		statusCode := cmp.Or(td.statusCode, http.StatusOK)
		if td.redirectURL != "" {
			http.Redirect(response, request, td.redirectURL, statusCode)
			return
		}
		if contentType := response.Header().Get("content-type"); contentType == "" {
			response.Header().Set("content-type", "text/html; charset=utf-8")
		}
		response.Header().Set("content-length", strconv.Itoa(buf.Len()))
		response.WriteHeader(statusCode)
		_, _ = buf.WriteTo(response)
	})
	mux.HandleFunc("GET /fruits/{id}/edit", func(response http.ResponseWriter, request *http.Request) {
		idParsed, err := strconv.Atoi(request.PathValue("id"))
		if err != nil {
			var zv Row
			rd := newTemplateData(receiver, response, request, zv, false, err)
			buf := bytes.NewBuffer(nil)
			if err := templates.ExecuteTemplate(buf, "GET /fruits/{id}/edit GetFormEditRow(id)", rd); err != nil {
				slog.ErrorContext(request.Context(), "failed to render page", slog.String("path", request.URL.Path), slog.String("pattern", request.Pattern), slog.String("error", err.Error()))
				http.Error(response, "failed to render page", http.StatusInternalServerError)
				return
			}
			sc := cmp.Or(rd.statusCode, http.StatusBadRequest)
			if rd.redirectURL != "" {
				http.Redirect(response, request, rd.redirectURL, sc)
				return
			}
			if contentType := response.Header().Get("content-type"); contentType == "" {
				response.Header().Set("content-type", "text/html; charset=utf-8")
			}
			response.Header().Set("content-length", strconv.Itoa(buf.Len()))
			response.WriteHeader(sc)
			_, _ = buf.WriteTo(response)
			return
		}
		id := idParsed
		result, err := receiver.GetFormEditRow(id)
		if err != nil {
			var zv Row
			rd := newTemplateData(receiver, response, request, zv, false, err)
			buf := bytes.NewBuffer(nil)
			if err := templates.ExecuteTemplate(buf, "GET /fruits/{id}/edit GetFormEditRow(id)", rd); err != nil {
				slog.ErrorContext(request.Context(), "failed to render page", slog.String("path", request.URL.Path), slog.String("pattern", request.Pattern), slog.String("error", err.Error()))
				http.Error(response, "failed to render page", http.StatusInternalServerError)
				return
			}
			sc := cmp.Or(rd.statusCode, http.StatusInternalServerError)
			if rd.redirectURL != "" {
				http.Redirect(response, request, rd.redirectURL, sc)
				return
			}
			if contentType := response.Header().Get("content-type"); contentType == "" {
				response.Header().Set("content-type", "text/html; charset=utf-8")
			}
			response.Header().Set("content-length", strconv.Itoa(buf.Len()))
			response.WriteHeader(sc)
			_, _ = buf.WriteTo(response)
			return
		}
		td := newTemplateData(receiver, response, request, result, true, nil)
		buf := bytes.NewBuffer(nil)
		if err := templates.ExecuteTemplate(buf, "GET /fruits/{id}/edit GetFormEditRow(id)", td); err != nil {
			slog.ErrorContext(request.Context(), "failed to render page", slog.String("path", request.URL.Path), slog.String("pattern", request.Pattern), slog.String("error", err.Error()))
			http.Error(response, "failed to render page", http.StatusInternalServerError)
			return
		}
		statusCode := cmp.Or(td.statusCode, http.StatusOK)
		if td.redirectURL != "" {
			http.Redirect(response, request, td.redirectURL, statusCode)
			return
		}
		if contentType := response.Header().Get("content-type"); contentType == "" {
			response.Header().Set("content-type", "text/html; charset=utf-8")
		}
		response.Header().Set("content-length", strconv.Itoa(buf.Len()))
		response.WriteHeader(statusCode)
		_, _ = buf.WriteTo(response)
	})
	mux.HandleFunc("GET /help", func(response http.ResponseWriter, request *http.Request) {
		result := struct {
		}{}
		buf := bytes.NewBuffer(nil)
		td := newTemplateData(receiver, response, request, result, true, nil)
		if err := templates.ExecuteTemplate(buf, "GET /help", td); err != nil {
			slog.ErrorContext(request.Context(), "failed to render page", slog.String("path", request.URL.Path), slog.String("pattern", request.Pattern), slog.String("error", err.Error()))
			http.Error(response, "failed to render page", http.StatusInternalServerError)
			return
		}
		statusCode := cmp.Or(td.statusCode, http.StatusOK)
		if td.redirectURL != "" {
			http.Redirect(response, request, td.redirectURL, statusCode)
			return
		}
		if contentType := response.Header().Get("content-type"); contentType == "" {
			response.Header().Set("content-type", "text/html; charset=utf-8")
		}
		response.Header().Set("content-length", strconv.Itoa(buf.Len()))
		response.WriteHeader(statusCode)
		_, _ = buf.WriteTo(response)
	})
	mux.HandleFunc("GET /{$}", func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		result := receiver.List(ctx)
		td := newTemplateData(receiver, response, request, result, true, nil)
		buf := bytes.NewBuffer(nil)
		if err := templates.ExecuteTemplate(buf, "GET /{$} List(ctx)", td); err != nil {
			slog.ErrorContext(request.Context(), "failed to render page", slog.String("path", request.URL.Path), slog.String("pattern", request.Pattern), slog.String("error", err.Error()))
			http.Error(response, "failed to render page", http.StatusInternalServerError)
			return
		}
		statusCode := cmp.Or(td.statusCode, http.StatusOK)
		if td.redirectURL != "" {
			http.Redirect(response, request, td.redirectURL, statusCode)
			return
		}
		if contentType := response.Header().Get("content-type"); contentType == "" {
			response.Header().Set("content-type", "text/html; charset=utf-8")
		}
		response.Header().Set("content-length", strconv.Itoa(buf.Len()))
		response.WriteHeader(statusCode)
		_, _ = buf.WriteTo(response)
	})
}

type TemplateData[T any] struct {
	receiver    RoutesReceiver
	response    http.ResponseWriter
	request     *http.Request
	result      T
	statusCode  int
	okay        bool
	err         error
	redirectURL string
}

func newTemplateData[T any](receiver RoutesReceiver, response http.ResponseWriter, request *http.Request, result T, okay bool, err error) *TemplateData[T] {
	return &TemplateData[T]{receiver: receiver, response: response, request: request, result: result, okay: okay, err: err, redirectURL: ""}
}

func (data *TemplateData[T]) Path() TemplateRoutePaths {
	return TemplateRoutePaths{}
}

func (data *TemplateData[T]) Result() T {
	return data.result
}

func (data *TemplateData[T]) Request() *http.Request {
	return data.request
}

func (data *TemplateData[T]) StatusCode(statusCode int) *TemplateData[T] {
	data.statusCode = statusCode
	return data
}

func (data *TemplateData[T]) Header(key, value string) *TemplateData[T] {
	data.response.Header().Set(key, value)
	return data
}

func (data *TemplateData[T]) Ok() bool {
	return data.okay
}

func (data *TemplateData[T]) Err() error {
	return data.err
}

func (data *TemplateData[T]) Receiver() RoutesReceiver {
	return data.receiver
}

func (data *TemplateData[T]) Redirect(url string, code int) (*TemplateData[T], error) {
	if code < 300 || code >= 400 {
		return data, fmt.Errorf("invalid status code %d for redirect", code)
	}
	data.redirectURL = url
	return data.StatusCode(code), nil
}

type TemplateRoutePaths struct {
}

func (TemplateRoutePaths) SubmitFormEditRow(id int) string {
	return "/" + path.Join("fruits", strconv.Itoa(id))
}

func (TemplateRoutePaths) GetFormEditRow(id int) string {
	return "/" + path.Join("fruits", strconv.Itoa(id), "edit")
}

func (TemplateRoutePaths) ReadHelp() string {
	return "/help"
}

func (TemplateRoutePaths) List() string {
	return "/"
}
