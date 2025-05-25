package ember

import (
	"context"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/kr/pretty"
)

var dialect = SqliteDialect()

type DbMigration struct {
	Version int   `db:"version"`
	Applied int64 `db:"applied"`
}

var (
	migrationTable   = goqu.T("ember_migrations")
	migrationVersion = migrationTable.Col("version")
	migrationApplied = migrationTable.Col("applied")
)

func MigrationQuery(db DB) *goqu.SelectDataset {
	return dialect.Select(migrationTable).
		Select(
			migrationVersion,
			migrationApplied,
		)
}

func GetAllMigrations(ctx context.Context, db DB) ([]DbMigration, error) {
	query := MigrationQuery(db).
		Order(migrationVersion.Asc())

	var res []DbMigration
	err := db.Multiple(ctx, query, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

type CreateMigrationParams struct {
	Version int
	Applied int64
}

func CreateMigration(ctx context.Context, db ember.DB, params CreateMigrationParams) error {
	query := dialect.Insert(migrationTable).
		Rows(goqu.Record{
			"version": params.Version,
			"applied": params.Applied,
		})

	_, err := db.Exec(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

type Migration struct {
	Title   string
	Version int
	Done    bool

	Up   func(ctx context.Context, db DB) error
	Down func(ctx context.Context, db DB) error
}

func SetupMigrations(ctx context.Context, db DB) error {
	_, err := db.Exec(ctx, RawQuery{
		Sql: `
		CREATE TABLE IF NOT EXISTS ember_migrations(
			version INTEGER PRIMARY KEY,
			applied INTEGER NOT NULL
		);
		`,
		Params: []any{},
	})
	if err != nil {
		return err
	}

	return nil
}

func ApplyMigrations(ctx context.Context, db *Database, migrations []Migration) error {
	versions, err := GetAllMigrations(ctx, db)
	if err != nil {
		return err
	}

	applied := make(map[int]bool)
	for _, v := range versions {
		applied[v.Version] = true
	}

	currentVersion := 0
	if len(versions) > 0 {
		currentVersion = versions[len(versions)-1].Version
	}

	fmt.Printf("currentVersion: %v\n", currentVersion)

	// TODO(patrik):
	//  - Get all the app registerd migrations
	//  	- Sort
	//  - Get all migrations applied inside the database
	//  - Find missing migrations
	//  - Check for duplicated migrations
	//  - Check app migrations
	//    - Version != 0
	//  - Transactions
	//  - Apply the missing migrations

	var migrationsToApply []Migration

	for _, migration := range migrations {
		if applied[migration.Version] {
			continue
		}

		if migration.Version > currentVersion {
			migrationsToApply = append(migrationsToApply, migration)
		}
	}

	pretty.Println(migrationsToApply)

	for _, migration := range migrationsToApply {
		if migration.Up != nil {
			err := migration.Up(ctx, db)
			if err != nil {
				return err
			}
		}

		err = CreateMigration(ctx, db, CreateMigrationParams{
			Version: migration.Version,
			Applied: time.Now().UnixMilli(),
		})
		if err != nil {
			return err
		}
	}

	return nil
}
