package bpmonitor

import (
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/justinas/alice"
	"net/http"
)

type Server interface {
	// Serve  is a blocking function
	Serve() error
}

type bpserver struct {
	host      string
	port      int
	adminPass string
	mysqlCfg  mysql.Config
	db        *sql.DB
	endpoints []string
}

func NewServer(host string, port int, adminPass string, dsn mysql.Config) (Server, error) {
	ret := &bpserver{
		host:      host,
		port:      port,
		adminPass: adminPass,
	}

	scrubbedDSN := scrubDSN(dsn)
	fmt.Printf("DSN: %s\n", scrubbedDSN)
	db, err := sql.Open("mysql", dsn.FormatDSN())
	if err != nil {
		return nil, errors.Wrapf(err, "could not open %s", scrubbedDSN)
	}

	ret.db = db
	return ret, nil
}

// Serve is a blocking function
func (bp *bpserver) Serve() error {
	defer bp.db.Close()

	commonHandlers := alice.New(bp.aliceParseIncomingRequest)
	authenticatedHandlers := alice.New(bp.aliceParseIncomingRequest, bp.aliceParseIncomingUser)

	endpoints := map[string]http.Handler{
		"/admin/create/user":     commonHandlers.ThenFunc(bp.handleAdminCreateUser),
		"/user/create/sessionid": authenticatedHandlers.ThenFunc(bp.handleUserCreateSessionID),
		"/user/create/apikey":    authenticatedHandlers.ThenFunc(bp.handleUserCreateApikey),
		"/measurements/add":      authenticatedHandlers.ThenFunc(bp.handleAddMeasurement),
		"/measurements/get":      authenticatedHandlers.ThenFunc(bp.handleGetMeasurements),
		"/healthcheck":           commonHandlers.ThenFunc(bp.handleHealthcheck),
		//"/remove":              authenticatedHandlers.ThenFunc(bp.handleRemoveMeasurements),
		//"/edit":                authenticatedHandlers.ThenFunc(bp.handleEditMeasurements),
	}

	for ep, fn := range endpoints {
		http.Handle(ep, fn)
		bp.endpoints = append(bp.endpoints, ep)
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", bp.port), nil)
}

func scrubDSN(dsn mysql.Config) string {
	dsn.Passwd = "*****"
	return dsn.FormatDSN()
}
