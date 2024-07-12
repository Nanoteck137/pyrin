package client

type Endpoint struct {
	Name         string `json:"name"`
	Method       string `json:"method"`
	Path         string `json:"path"`
	ResponseType string `json:"responseType"`
	BodyType     string `json:"bodyType"`
}

type TypeField struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Omit bool   `json:"omit"`
}

type MetadataType struct {
	Name   string              `json:"name"`
	Extend string              `json:"extend"`
	Fields []TypeField `json:"fields"`
}

type Server struct {
	Types     []MetadataType     `json:"types"`
	Endpoints []Endpoint `json:"endpoints"`
}
