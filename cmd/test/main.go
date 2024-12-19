package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/kr/pretty"
	"github.com/nanoteck137/pyrin"
	"github.com/nanoteck137/pyrin/tools/transform"
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

type Test2Body struct {
	Name     string `json:"name"`
	LastName string `json:"lastName"`
	Age      int    `json:"age"`
}

func (b *Test2Body) Transform() {
	b.Name = transform.String(b.Name)
	b.LastName = transform.String(b.LastName)
}

func (b Test2Body) Validate() error {
	return validate.ValidateStruct(&b,
		validate.Field(&b.Name, validate.Required),
		validate.Field(&b.LastName, validate.Required),
		validate.Field(&b.Age, validate.Min(18)),
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
				Name:       "Test",
				Method:     http.MethodPost,
				Path:       "/test/:id",
				ReturnType: nil,
				BodyType:   TestBody{},
				Errors:     []pyrin.ErrorType{},
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

			v1.Register(pyrin.FormApiHandler{
				Name:   "Test2",
				Method: http.MethodPost,
				Path:   "/test",
				Spec: pyrin.FormSpec{
					BodyType: Test2Body{},
					Files: map[string]pyrin.FormFileSpec{
						"files": {
							NumExpected: 2,
						},
					},
				},
				HandlerFunc: func(c pyrin.Context) (any, error) {
					body, err := pyrin.Body[Test2Body](c)
					if err != nil {
						return nil, err
					}

					pretty.Println(body)

					files, err := pyrin.FormFiles(c, "files")
					if err != nil {
						return nil, err
					}

					fmt.Printf("files: %v\n", files)

					for _, f := range files {
						fmt.Printf("f.Filename: %v\n", f.Filename)
						fmt.Printf("f.Size: %v\n", f.Size)
						fmt.Printf("f.Header: %v\n", f.Header)
					}

					return nil, nil
				},
			})
		},
	})

	err := server.Start(":1337")
	if err != nil {
		log.Fatal(err)
	}
}
