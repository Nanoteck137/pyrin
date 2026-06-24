package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/kr/pretty"
	"github.com/nanoteck137/pyrin"
	"github.com/nanoteck137/pyrin/spark"
	"github.com/nanoteck137/pyrin/spark/dart"
	"github.com/nanoteck137/pyrin/spark/golang"
	"github.com/nanoteck137/pyrin/spark/typescript"
	"github.com/nanoteck137/validate"
)

type TestBody struct {
	Username        string `json:"username,omitempty"`
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

type MapTest struct {
	A string `json:"a"`
	B int    `json:"b"`
}

type Test2Body struct {
	Type     string              `json:"type"`
	Number   int                 `json:"number"`
	Id       string              `json:"id"`
	Name     string              `json:"name"`
	LastName string              `json:"lastName"`
	Age      int                 `json:"age"`
	Map      map[string]*MapTest `json:"mapTest"`
}

func (b *Test2Body) Transform() {
	b.Name = strings.TrimSpace(b.Name)
	b.LastName = strings.TrimSpace(b.LastName)
}

func (b Test2Body) Validate() error {
	return validate.ValidateStruct(&b,
		validate.Field(&b.Name, validate.Required),
		validate.Field(&b.LastName, validate.Required),
		validate.Field(&b.Age, validate.Min(18)),
	)
}

func registerRoutes(router pyrin.Router) {
	root := router.Group("/")
	root.Register(pyrin.NormalHandler{
		Method: http.MethodGet,
		Path:   "/file",
		HandlerFunc: func(c pyrin.Context) error {
			fs := os.DirFS(".")
			return pyrin.ServeFile(c, fs, "pyrin.go")
		},
	})

	root.Register(pyrin.ApiHandler{
		Name:   "Test123",
		Method: http.MethodPost,
		Path:   "/test/123",
		HandlerFunc: func(c pyrin.Context) (any, error) {
			fmt.Println("123 hit")
			return nil, errors.New("test error")
		},
	})

	v1 := router.Group("/api/v1")
	v1.Register(pyrin.ApiHandler{
		Name:   "Test123",
		Method: http.MethodPost,
		Path:   "/test/123",
		HandlerFunc: func(c pyrin.Context) (any, error) {
			fmt.Println("123 hit")
			return nil, errors.New("test error")
		},
	})

	v1.Register(pyrin.ApiHandler{
		Name:     "Test",
		Method:   http.MethodPost,
		Path:     "/test/:id",
		BodyType: TestBody{},
		HandlerFunc: func(c pyrin.Context) (any, error) {
			id := c.Param("id")

			pretty.Println(c.Request().Header)

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
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func loggerMiddleware(logName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sr, r)

			slog.LogAttrs(r.Context(), slog.LevelInfo, logName,
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", sr.status),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	if true {
		router := spark.Router{}
		registerRoutes(&router)

		fieldNameFilter := spark.NameFilter{}

		serverDef, err := spark.CreateServerDef(&router, fieldNameFilter)
		if err != nil {
			slog.Error("failed", "err", err)
			return
		}

		pretty.Println(serverDef)

		err = serverDef.SaveToFile("./work/pyrin.json")
		if err != nil {
			slog.Error("failed", "err", err)
			return
		}

		resolver, err := spark.CreateResolverFromServerDef(&serverDef)
		if err != nil {
			slog.Error("failed", "err", err)
			return
		}

		{
			gen := typescript.TypescriptGenerator{
				NameMapping: map[string]string{
					"TestBody.@username": "username",
				},
			}

			err = gen.Generate(&serverDef, resolver, "../pyrin-test-projects/typescript/api")
			if err != nil {
				slog.Error("failed typescript", "err", err)
				return
			}
		}

		{
			gen := golang.GolangGenerator{
				NameMapping: map[string]string{
					"TestBody.@username": "username",
				},
			}

			err = gen.Generate(&serverDef, resolver, "../pyrin-test-projects/golang/api")
			if err != nil {
				slog.Error("failed golang", "err", err)
				return
			}
		}

		{
			gen := dart.DartGenerator{
				NameMapping: map[string]string{
					"TestBody.@username": "username",
				},
			}

			err = gen.Generate(&serverDef, resolver, "../pyrin-test-projects/dart/lib/api")
			if err != nil {
				slog.Error("failed dart", "err", err)
				return
			}
		}
	}

	server := pyrin.NewServer(&pyrin.ServerConfig{
		ErrorCallback: func(err error) {
			slog.Error("API Error", "err", err)
		},
		RegisterHandlers: registerRoutes,
		Middlewares: []pyrin.MiddlewareFunc{
			loggerMiddleware("Test"),
			corsMiddleware,
			middleware.Recoverer,
		},
	})

	registerRoutes(server)

	fmt.Println("Starting on :1337")
	err := server.Start(":1337")
	if err != nil {
		slog.Error("failed", "err", err)
	}
}
