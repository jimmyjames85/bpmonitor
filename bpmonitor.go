package bpmonitor

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/go-sql-driver/mysql"
	"github.com/justinas/alice"
	"github.com/pkg/errors"
)

type Server interface {
	// Serve  is a blocking function
	Serve() error
}

type bpserver struct {
	port          int
	adminPass     string
	mysqlCfg      mysql.Config
	db            *sql.DB
	endpoints     []string
	sslPemFileloc string
	sslKeyFileloc string
}

func NewServer(port int, adminPass string, dsn mysql.Config, sslPemFileloc, sslKeyFileloc string) (Server, error) {

	ret := &bpserver{
		port:          port,
		adminPass:     adminPass,
		sslPemFileloc: sslPemFileloc,
		sslKeyFileloc: sslKeyFileloc,
	}

	dsn.ParseTime = true
	scrubbedDSN := scrubDSN(dsn)
	log.Printf("DSN: %s\n", scrubbedDSN)
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
		"/measurements/remove":   authenticatedHandlers.ThenFunc(bp.handleRemoveMeasurements),
		"/measurements/edit":     authenticatedHandlers.ThenFunc(bp.handleEditMeasurements),
		"/measurements/graph":    authenticatedHandlers.ThenFunc(bp.handleGraphMeasurements),
		"/healthcheck":           commonHandlers.ThenFunc(bp.handleHealthcheck),
	}

	for ep, fn := range endpoints {
		http.Handle(ep, fn)
		bp.endpoints = append(bp.endpoints, ep)
	}

	if len(bp.sslKeyFileloc) > 0 && len(bp.sslPemFileloc) > 0 {
		log.Println("starting ssl")
		return http.ListenAndServeTLS(fmt.Sprintf(":%d", bp.port), bp.sslPemFileloc, bp.sslKeyFileloc, nil)
	}
	return http.ListenAndServe(fmt.Sprintf(":%d", bp.port), nil)
}

func scrubDSN(dsn mysql.Config) string {
	dsn.Passwd = "*****"
	return dsn.FormatDSN()
}
