package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kr/pretty"
	"github.com/nanoteck137/pyrin"
	"github.com/nanoteck137/pyrin/api"
)

type TestBody struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

var usernameRegex = regexp.MustCompile("^[a-zA-Z0-9-]+$")

func (b TestBody) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.Username, validation.Required, validation.Length(4, 32), validation.Match(usernameRegex).Error("not valid username")),
		validation.Field(&b.Password, validation.Required, validation.Length(8, 32)),
		validation.Field(&b.ConfirmPassword, validation.Required, validation.Length(8, 32), validation.By(func(value interface{}) error {
			s, _ := value.(string)

			if s != b.Password {
				return errors.New("password mismatch")
			}

			return nil
		})),
	)
}

func main() {
	server := pyrin.NewServer()

	v1 := server.Group("/api/v1")
	v1.Register(pyrin.Handler{
		Name:     "Test",
		Method:   http.MethodPost,
		Path:     "/test/:id",
		DataType: nil,
		BodyType: TestBody{},
		Errors:   []api.ErrorType{},
		HandlerFunc: func(c pyrin.Context) (any, error) {
			id := c.Param("id")

			body := c.Body().(TestBody)
			pretty.Println(body)

			if id == "123" {
				return struct {
					Value int `json:"value"`
				}{
					Value: 123,
				}, nil
			}

			return nil, &api.Error{
				Code:    404,
				Type:    "NOT_FOUND_TEST",
				Message: "Testing",
				Extra:   nil,
			}
		},
	})

	err := server.Start(":1337")
	if err != nil {
		log.Fatal(err)
	}
}
