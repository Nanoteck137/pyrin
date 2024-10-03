package main

import (
	"log"
	"net/http"

	"github.com/nanoteck137/pyrin"
	"github.com/nanoteck137/pyrin/api"
)

func main() {
	server := pyrin.NewServer()

	v1 := server.Group("/api/v1")
	v1.Register(pyrin.Handler{
		Name:        "Test",
		Method:      http.MethodGet,
		Path:        "/test/:id",
		DataType:    nil,
		BodyType:    nil,
		Errors:      []api.ErrorType{},
		HandlerFunc: func(c pyrin.Context) (any, error) {
			id := c.Param("id")

			var t any = id

			l := t.(int)

			_ = l

			if id == "123" {
				return struct{
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
