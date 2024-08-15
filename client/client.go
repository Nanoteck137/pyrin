package client

import "github.com/nanoteck137/pyrin/api"

type Endpoint struct {
	Name         string          `json:"name"`
	Method       string          `json:"method"`
	Path         string          `json:"path"`
	ErrorTypes   []api.ErrorType `json:"errorTypes"`
	ResponseType string          `json:"responseType"`
	BodyType     string          `json:"bodyType"`
}

type TypeField struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Omit bool   `json:"omit"`
}

type Type struct {
	Name   string      `json:"name"`
	Extend string      `json:"extend"`
	Fields []TypeField `json:"fields"`
}

type Server struct {
	Types     []Type     `json:"types"`
	Endpoints []Endpoint `json:"endpoints"`
}
