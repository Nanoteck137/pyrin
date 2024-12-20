package spec

type ApiEndpoint struct {
	Name            string   `json:"name"`
	Method          string   `json:"method"`
	Path            string   `json:"path"`
	ResponseType    string   `json:"responseType"`
	BodyType        string   `json:"bodyType"`
}

type FormApiEndpoint struct {
	Name            string   `json:"name"`
	Method          string   `json:"method"`
	Path            string   `json:"path"`
	ResponseType    string   `json:"responseType"`
	BodyType        string   `json:"bodyType"`
}

type NormalEndpoint struct {
	Name            string   `json:"name"`
	Method          string   `json:"method"`
	Path            string   `json:"path"`
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
	ErrorTypes       []string          `json:"errorTypes"`
	Types            []Type            `json:"types"`
	ApiEndpoints     []ApiEndpoint     `json:"apiEndpoints"`
	FormApiEndpoints []FormApiEndpoint `json:"formApiEndpoints"`
	NormalEndpoints  []NormalEndpoint  `json:"normalEndpoints"`
}
