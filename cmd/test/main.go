package main

import (
	"encoding/json"
	"errors"
	"fmt"
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

func Body[T validation.Validatable](c pyrin.Context) (T, error) {
	var res T

	decoder := json.NewDecoder(c.Request().Body)

	err := decoder.Decode(&res)
	if err != nil {
		return res, err
	}

	err = res.Validate()
	if err != nil {
		extra := make(map[string]string)

		if e, ok := err.(validation.Errors); ok {
			for k, v := range e {
				extra[k] = v.Error()
			}
		}

		return res, &api.Error{
			Code:    400,
			Type:    "VALIDATION_ERROR",
			Message: "Body Validation error",
			Extra:   extra,
		}
	}

	return res, nil
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

			body, err := Body[TestBody](c)
			if err != nil {
				fmt.Printf("err: %v\n", err)
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
