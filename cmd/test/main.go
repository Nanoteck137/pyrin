package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/kr/pretty"
	"github.com/nanoteck137/pyrin"
	"github.com/nanoteck137/validate"
)

type TestBody struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

var usernameRegex = regexp.MustCompile("^[a-zA-Z0-9-]+$")

func (b *TestBody) Transform() {
	b.Username = strings.TrimSpace(b.Username)
}

func (b TestBody) Validate() error {
	return validate.ValidateStruct(&b,
		validate.Field(&b.Username, validate.Required, validate.Length(4, 32), validate.Match(usernameRegex).Error("not valid username")),
		validate.Field(&b.Password, validate.Required, validate.Length(8, 32)),
		validate.Field(&b.ConfirmPassword, validate.Required, validate.By(func(value interface{}) error {
			s, _ := value.(string)

			if s != b.Password {
				return errors.New("password mismatch")
			}

			return nil
		})),
	)
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

					body, err := pyrin.Body[TestBody](c)
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
