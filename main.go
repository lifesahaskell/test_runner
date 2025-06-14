package test_runner

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"testing"

	"github.com/ory/dockertest/v3"

	_ "github.com/sijms/go-ora/v2" // Oracle driver
)

var db *sql.DB

var dbParams = map[string]string{
	"server":          "localhost",
	"port":            "1521",
	"service":         "orclpdb1",
	"username":        "admin",
	"password":        "secret",
	"walletLocation":  "", // Optional, if using Oracle Wallet
	"ssl":             "enable", // Optional, if using SSL
}

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	// uses pool to try to connect to Docker
	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("oracle", "19c", []string{fmt.Sprintf("ORACLE_ADMIN_PASSWORD=%s",  dbParams["password"])})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {

		conn_string := buildConnectionString(dbParams);
		
		db, err = sql.Open("oracle", conn_string)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	// as of go1.15 testing.M returns the exit code of m.Run(), so it is safe to use defer here
    defer func() {
      if err := pool.Purge(resource); err != nil {
        log.Fatalf("Could not purge resource: %s", err)
      }

    }()

	m.Run()
}

func buildConnectionString(dbParams map[string]string) string {
	conn_string := "oracle://" + dbParams["username"] + ":" + dbParams["password"] + "@" + dbParams["server"] + ":" + dbParams["port"] + "/" + dbParams["service"]
	if val, ok := dbParams["walletLocation"]; ok && val != "" {
		conn_string += "?TRACE FILE=trace.log&SSL=enable&SSL Verify=false&WALLET=" + url.QueryEscape(dbParams["walletLocation"])
	}
	return conn_string
}

func TestSomething(t *testing.T) {
	// db.Query()
}