package main

import (
	"fmt"
	"github.com/jimmyjames85/bpmonitor"
	"github.com/kelseyhightower/envconfig"
	"github.com/go-sql-driver/mysql"
	"log"
)

type config struct {
	AdminPass string `envconfig:"ADMIN_PASS" required:"true"`                // is used to add new users
	Port      int `envconfig:"PORT" required:"false" default:"1234"`         // port to run on
	Host      string `envconfig:"HOST" required:"false" default:"localhost"` // for the web form data to know which host to hit
	DBuser    string `envconfig:"DB_USER" required "true"`
	DBPswd    string `envconfig:"DB_PASS" required "true"`
	DBHost    string `envconfig:"DB_HOST" required "true"`
	DBPort    int    `envconfig:"DB_PORT" required "true"`
	DBName    string `envconfig:"DB_NAME" required "true"`
}

func main() {
	c := &config{}
	envconfig.MustProcess("BP", c)

	dsn := mysql.Config{}
	dsn.Addr = fmt.Sprintf("%s:%d", c.DBHost, c.DBPort)
	dsn.Passwd = c.DBPswd
	dsn.User = c.DBuser
	dsn.DBName = c.DBName
	dsn.Net = "tcp"

	bp, err := bpmonitor.NewServer(c.Host, c.Port, c.AdminPass, dsn)
	if err != nil {
		log.Fatalf("unable to start server: %v\n", err)
	}
	err = bp.Serve()
	if err != nil {
		log.Fatalf("Error occured while serving: %v\n", err)
	}

}
