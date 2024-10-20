package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/kr/pretty"
	"github.com/labstack/echo/v4"
	"github.com/nanoteck137/pyrin"
	"github.com/nanoteck137/pyrin/tools/validate"
	"github.com/nanoteck137/pyrin/tools/validate/rules"
)

type TestBody struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

var usernameRegex = regexp.MustCompile("^[a-zA-Z0-9-]+$")

func (b TestBody) Validate(v validate.Validator) error {
	return v.Struct(&b,
		v.Field(&b.Username, rules.Required, rules.Length(4, 32), rules.Match(usernameRegex).Error("not valid username")),
		v.Field(&b.Password, rules.Required, rules.Length(8, 32)),
		v.Field(&b.ConfirmPassword, rules.Required, rules.By(func(value interface{}) error {
			s, _ := value.(string)

			if s != b.Password {
				return errors.New("password mismatch")
			}

			return nil
		})),
	)
}

// TODO(patrik): Move to pyrin.go
func Body[T validate.Validatable](c pyrin.Context) (T, error) {
	var res T

	decoder := json.NewDecoder(c.Request().Body)

	err := decoder.Decode(&res)
	if err != nil {
		pretty.Println(err)
		return res, err
	}

	validator := validate.NormalValidator{}
	err = res.Validate(&validator)
	if err != nil {
		extra := make(map[string]string)

		if e, ok := err.(validate.Errors); ok {
			for k, v := range e {
				extra[k] = v.Error()
			}
		}

		return res, pyrin.ValidationError(extra)
	}

	return res, nil
}

// TODO(patrik): Add to pyrin helpers
func fsFile(w http.ResponseWriter, r *http.Request, file string) error {
	f, err := os.Open(file)
	if err != nil {
		// TODO(patrik): Add NoContentError to pyrin
		return echo.ErrNotFound
	}
	defer f.Close()

	fi, _ := f.Stat()

	http.ServeContent(w, r, fi.Name(), fi.ModTime(), f)

	return nil
}

func main() {
	server := pyrin.NewServer(&pyrin.ServerConfig{
		RegisterHandlers: func(router pyrin.Router) {
			root := router.Group("")
			root.Register(pyrin.NormalHandler{
				Method: http.MethodGet,
				Path:   "/file",
				HandlerFunc: func(c pyrin.Context) error {
					fs := os.DirFS(".")
					return pyrin.ServeFile(c.Response(), c.Request(), fs, "pyrin.go")
				},
			})

			v1 := router.Group("/api/v1")
			v1.Register(pyrin.ApiHandler{
				Name:     "Test",
				Method:   http.MethodPost,
				Path:     "/test/:id",
				DataType: nil,
				BodyType: TestBody{},
				Errors:   []pyrin.ErrorType{},
				HandlerFunc: func(c pyrin.Context) (any, error) {
					id := c.Param("id")

					// body := c.Body().(TestBody)
					body, err := Body[TestBody](c)
					if err != nil {
						return nil, err
					}
					pretty.Println(body)

					if id == "123" {
						return struct {
							Value int `json:"value"`
						}{
							Value: 123,
						}, nil
					}

					return nil, &pyrin.Error{
						Code:    404,
						Type:    "NOT_FOUND_TEST",
						Message: "Testing",
						Extra:   nil,
					}
				},
			})
		},
	})

	err := server.Start(":1337")
	if err != nil {
		log.Fatal(err)
	}
}
