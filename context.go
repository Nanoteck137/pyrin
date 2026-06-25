package pyrin

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/nanoteck137/validate"
)

var _ Context = (*wrapperContext)(nil)

type wrapperContext struct {
	w http.ResponseWriter
	r *http.Request

	formSpec *FormSpec
}

func (w *wrapperContext) Response() http.ResponseWriter {
	return w.w
}

func (w *wrapperContext) Request() *http.Request {
	return w.r
}

func (w *wrapperContext) Param(name string) string {
	return chi.URLParam(w.r, name)
}

func (w *wrapperContext) checkContentType(expected string) error {
	contentType := w.r.Header.Get("Content-Type")
	if contentType == "" {
		return BadContentType(expected)
	}

	typ, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return BadContentType(expected)
	}

	if typ != expected {
		return BadContentType(expected)
	}

	return nil
}

func FormFiles(c Context, key string) ([]*multipart.FileHeader, error) {
	wrapperContext := c.(*wrapperContext)

	spec := wrapperContext.formSpec

	if spec == nil {
		return nil, errors.New("handler cannot use forms use 'FormApiHandler'")
	}

	_, exists := spec.Files[key]
	if !exists {
		return nil, fmt.Errorf("%s: is not valid, key is not defined in spec", key)
	}

	form := c.Request().MultipartForm
	files := form.File[key]

	return files, nil
}

func Body[T any](c Context) (T, error) {
	var res T

	wrapperContext := c.(*wrapperContext)

	var body io.Reader
	if wrapperContext.formSpec == nil {
		body = c.Request().Body
	} else {
		data := c.Request().FormValue(formBodyKey)
		body = strings.NewReader(data)
	}

	decoder := json.NewDecoder(body)

	if !decoder.More() {
		return res, EmptyBody()
	}

	err := decoder.Decode(&res)
	if err != nil {
		return res, err
	}

	var p any = &res
	if t, ok := p.(Transformable); ok {
		t.Transform()
	}

	if v, ok := p.(validate.Validatable); ok {
		err = v.Validate()
		if err != nil {
			extra := make(map[string]string)

			if e, ok := err.(validate.Errors); ok {
				for k, v := range e {
					extra[k] = v.Error()
				}
			}

			return res, ValidationError(extra)
		}
	}

	return res, nil
}
