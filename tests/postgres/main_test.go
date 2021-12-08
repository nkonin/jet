package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-jet/jet/v2/tests/internal/utils/repo"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/stdlib"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/tests/dbconfig"
	_ "github.com/lib/pq"
	"github.com/pkg/profile"
	"github.com/stretchr/testify/require"

	_ "github.com/jackc/pgx/v4/stdlib"
)

var db *sql.DB
var testRoot string

func TestMain(m *testing.M) {
	rand.Seed(time.Now().Unix())
	defer profile.Start().Stop()

	setTestRoot()

	for _, driverName := range []string{"pgx", "postgres"} {
		fmt.Printf("\nRunning postgres tests for '%s' driver\n", driverName)

		func() {
			var err error
			db, err = sql.Open(driverName, dbconfig.PostgresConnectString)
			if err != nil {
				fmt.Println(err.Error())
				panic("Failed to connect to test db")
			}
			defer db.Close()

			ret := m.Run()

			if ret != 0 {
				os.Exit(ret)
			}
		}()
	}
}

func setTestRoot() {
	testRoot = repo.GetTestsDirPath()
}

var loggedSQL string
var loggedSQLArgs []interface{}
var loggedDebugSQL string

func init() {
	postgres.SetLogger(func(ctx context.Context, statement postgres.PrintableStatement) {
		loggedSQL, loggedSQLArgs = statement.Sql()
		loggedDebugSQL = statement.DebugSql()
	})
}

func requireLogged(t *testing.T, statement postgres.Statement) {
	query, args := statement.Sql()
	require.Equal(t, loggedSQL, query)
	require.Equal(t, loggedSQLArgs, args)
	require.Equal(t, loggedDebugSQL, statement.DebugSql())
}

func skipForPgxDriver(t *testing.T) {
	if isPgxDriver() {
		t.SkipNow()
	}
}

func isPgxDriver() bool {
	switch db.Driver().(type) {
	case *stdlib.Driver:
		return true
	}

	return false
}

func beginTx(t *testing.T) *sql.Tx {
	tx, err := db.Begin()
	require.NoError(t, err)
	return tx
}
