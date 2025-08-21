package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/kr/pretty"
	"github.com/nanoteck137/pyrin"
	"github.com/nanoteck137/pyrin/anvil"
	"github.com/nanoteck137/pyrin/ember"
	"github.com/nanoteck137/pyrin/spark"
	"github.com/nanoteck137/pyrin/spark/dart"
	"github.com/nanoteck137/pyrin/spark/golang"
	"github.com/nanoteck137/pyrin/spark/typescript"
	"github.com/nanoteck137/pyrin/trail"
	"github.com/nanoteck137/validate"

	_ "github.com/mattn/go-sqlite3"
)

var logger = trail.NewLogger(&trail.Options{Debug: true, Level: slog.LevelInfo})

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
	Id       string             `json:"id"`
	Name     string             `json:"name"`
	LastName string             `json:"lastName"`
	Age      int                `json:"age"`
	Map      map[string]*MapTest `json:"mapTest"`
}

func (b *Test2Body) Body(c anvil.Context) {
	c.Field("id", &b.Id, c.Trim(), c.Required())
	c.Field("name", &b.Name, c.Trim(), c.Required())
	c.Field("lastName", &b.LastName, c.Trim(), c.Required())
	c.Field("age", &b.Age, c.Min(18))
}

func (b *Test2Body) Transform() {
	b.Name = anvil.String(b.Name)
	b.LastName = anvil.String(b.LastName)
}

func (b Test2Body) Validate() error {
	return validate.ValidateStruct(&b,
		validate.Field(&b.Name, validate.Required),
		validate.Field(&b.LastName, validate.Required),
		validate.Field(&b.Age, validate.Min(18)),
	)
}

func registerRoutes(router pyrin.Router) {
	root := router.Group("")
	root.Register(pyrin.NormalHandler{
		Name:   "GetFile",
		Method: http.MethodGet,
		Path:   "/file",
		HandlerFunc: func(c pyrin.Context) error {
			fs := os.DirFS(".")
			return pyrin.ServeFile(c, fs, "pyrin.go")
		},
	})

	v1 := router.Group("/api/v1")
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

type TestLel struct {
	Id   string  `json:"id"`
	Lel  *string `json:"lel,omitempty"`
	Test float32 `json:"test"`
}

type Testing struct {
	TestLel

	Lel int `json:"lel"`
}

func test(registry *spark.StructRegistry) error {
	return registry.Register(TestLel{})
}

func main() {

	{
		router := spark.Router{}
		registerRoutes(&router)

		fieldNameFilter := spark.NameFilter{}
		fieldNameFilter.LoadDefault()

		serverDef, err := spark.CreateServerDef(&router, fieldNameFilter)
		if err != nil {
			logger.Fatal("failed", "err", err)
		}

		pretty.Println(serverDef)

		err = serverDef.SaveToFile("./work/pyrin.json")
		if err != nil {
			logger.Fatal("failed", "err", err)
		}

		resolver, err := spark.CreateResolverFromServerDef(&serverDef)
		if err != nil {
			logger.Fatal("failed", "err", err)
		}

		{
			gen := typescript.TypescriptGenerator{
				NameMapping: map[string]string{
					"TestBody": "LelBody",
					"id":       "bid",
				},
			}

			err = gen.Generate(&serverDef, resolver, "../pyrin-test-projects/typescript/api")
			if err != nil {
				logger.Fatal("failed typescript", "err", err)
			}
		}

		{
			gen := golang.GolangGenerator{
				NameMapping: map[string]string{
					"TestBody": "LelBody",
					"id":       "bid",
				},
			}

			err = gen.Generate(&serverDef, resolver, "../pyrin-test-projects/golang/api")
			if err != nil {
				logger.Fatal("failed golang", "err", err)
			}
		}

		{
			gen := dart.DartGenerator{
				NameMapping: map[string]string{
					"TestBody": "LelBody",
					"id":       "bid",
				},
			}

			err = gen.Generate(&serverDef, resolver, "../pyrin-test-projects/dart/lib/api")
			if err != nil {
				logger.Fatal("failed dart", "err", err)
			}
		}
	}

	// return
	// err := os.Remove("./work/test.db")
	// if err != nil {
	// 	fmt.Println("err:", err)
	// }

	if false {
		dbFile := "./work/test.db"
		dbUrl := fmt.Sprintf("file:%s?_foreign_keys=true", dbFile)
		db, err := ember.OpenDatabase("sqlite3", dbUrl)
		if err != nil {
			logger.Fatal("failed", "err", err)
		}

		migrations := []ember.Migration{
			{
				Title:   "init",
				Version: 1,
				Done:    false,
				Up: func(ctx context.Context, db ember.DB) error {
					_, err = db.Exec(ctx, ember.RawQuery{
						Sql: `
					CREATE TABLE tracks(
						id TEXT PRIMARY KEY,

						title TEXT NOT NULL,
						album_name TEXT NOT NULL,
						aritst_name TEXT NOT NULL
					);
					`,
						Params: []any{},
					})
					if err != nil {
						return err
					}

					return nil
				},
				Down: func(ctx context.Context, db ember.DB) error {
					_, err = db.Exec(ctx, ember.RawQuery{
						Sql: `
					DROP TABLE tracks;
					`,
						Params: []any{},
					})
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Title:   "test",
				Version: 3,
				Done:    false,
				Up:      nil,
				Down:    nil,
			},
			{
				Title:   "test",
				Version: 2,
				Done:    false,
				Up:      nil,
				Down:    nil,
			},
			{
				Title:   "test",
				Version: 4,
				Done:    false,
				Up:      nil,
				Down:    nil,
			},
		}

		sort.SliceStable(migrations, func(i, j int) bool {
			return migrations[i].Version < migrations[j].Version
		})

		ctx := context.Background()

		err = ember.SetupMigrations(ctx, db)
		if err != nil {
			logger.Fatal("failed", "err", err)
		}

		err = ember.ApplyMigrations(ctx, db, migrations)
		if err != nil {
			logger.Fatal("failed", "err", err)
		}
	}

	server := pyrin.NewServer(&pyrin.ServerConfig{
		RegisterHandlers: registerRoutes,
	})

	// router := spec.Router{}
	// registerRoutes(&router)
	//
	// s, err := spec.GenerateSpec(router.Routes)
	// if err != nil {
	// 	logger.Fatal("Failed to generate spec", "err", err)
	// }
	//
	// d, err := json.MarshalIndent(s, "", "  ")
	// if err != nil {
	// 	logger.Fatal("Failed to marshal server", "err", err)
	// }
	//
	// fmt.Printf("string(d): %v\n", string(d))
	//
	// err = gen.GenerateTypescript(s, "./work/ts")
	// if err != nil {
	// 	logger.Fatal("Failed", "err", err)
	// }
	//
	// _ = server
	err := server.Start(":1337")
	if err != nil {
		logger.Fatal("failed", "err", err)
	}
}

func init() {
	slog.SetDefault(logger.Logger)
}
