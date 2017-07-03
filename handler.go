package bpmonitor

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"runtime"
	"strconv"

	"github.com/jimmyjames85/bpmonitor/backend"
	"github.com/jimmyjames85/bpmonitor/backend/auth"
	"github.com/pkg/errors"
)

const passwordCookieName = "eWVrc2loV2hzYU1ydW9TZWVzc2VubmVUeXRpbGF1UWRuYXJCNy5vTmRsT2VtaXRkbE9zJ2xlaW5hRGtjYUoK"

type creds struct {
	username  *string
	password  *string
	apikey    *string
	sessionId *string
}

var (
	noUserInContext                            = fmt.Errorf("no user in context")
	defaultCustomerResponseInternalServerError = qm{"error": "Internal Server Error: please contact jimmyjames85"}
)

func (bp *bpserver) aliceParseIncomingRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			bp.handleInternalServerError(w, fmt.Errorf("failed to parse form data: %s", err), nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func validatePassword(pass string) error {
	if len(pass) < 5 {
		return errors.New("password must have at least five characters")
	}
	return nil
}

func validateUsername(username string) error {

	if len(username) < 3 {
		return errors.New("username must have at least three characters")
	}

	return nil
}

// todo this should be served on a different port
func (bp *bpserver) handleAdminCreateUser(w http.ResponseWriter, r *http.Request) {

	// allow cross domain AJAX requests
	w.Header().Set("Access-Control-Allow-Origin", "*")

	u, p, a := r.Form["user"], r.Form["pass"], r.Form["adminpass"]

	if len(a) == 0 || a[0] != bp.adminPass {
		bp.handleCustomerError(w, http.StatusUnauthorized, qm{"error": "Access Denied: Try this https://xkcd.com/538/"})
		return
	}

	if len(u) == 0 || len(p) == 0 {
		bp.handleCustomerError(w, http.StatusBadRequest, qm{"error": "username and password must be specified"})
		return
	}

	if err := validateUsername(u[0]); err != nil {
		bp.handleCustomerError(w, http.StatusBadRequest, qm{"error": err.Error()})
		return
	}

	if err := validatePassword(p[0]); err != nil {
		bp.handleCustomerError(w, http.StatusBadRequest, qm{"error": err.Error()})
		return
	}

	newUser := creds{username: &u[0], password: &p[0]}

	user, err := auth.CreateUser(bp.db, *newUser.username, *newUser.password)
	if err == auth.UserExists {
		bp.handleCustomerError(w, http.StatusConflict, qm{"error": "username is taken"})
		return
	} else if err != nil {
		bp.handleInternalServerError(w, err, qm{"error": "unknown error please contact the administrator"})
		return
	}

	io.WriteString(w, qm{"userid": user.ID}.toJSON())
}

func (bp *bpserver) handleGetMeasurements(w http.ResponseWriter, r *http.Request) {
	user := bp.mustGetUser(w, r)
	if user == nil {
		return
	}

	// TODO add date range
	measurements, err := backend.GetMeasurements(bp.db, user.ID)
	if err != nil {
		bp.handleInternalServerError(w, err, qm{"error": "retrieving measurements"})
		return
	}

	for _, m := range measurements {
		m.Notes = html.EscapeString(m.Notes)
	}

	io.WriteString(w, qm{"ok": true, "measurements": measurements}.toJSON())

}

func singleValue(values []string) (string, bool) {
	if len(values) > 0 {
		return values[0], true
	}
	return "", false
}

func (bp *bpserver) handleEditMeasurements(w http.ResponseWriter, r *http.Request) {
	user := bp.mustGetUser(w, r)
	if user == nil {
		return
	}

	var id int
	var sys, dia, pulse *int
	var notes *string

	if i, ok := singleValue(r.Form["id"]); !ok {
		bp.handleCustomerError(w, http.StatusBadRequest, qm{"error": "Must provide id of measurment to edit"})
		return
	} else {
		var err error
		id, err = strconv.Atoi(i)
		if err != nil {
			bp.handleCustomerError(w, http.StatusBadRequest, qm{"error": "Unable to parse measurement id"})
			return
		}
	}

	if systolic, ok := singleValue(r.Form["systolic"]); ok {
		s, err := strconv.Atoi(systolic)
		if err != nil {
			bp.handleCustomerError(w, http.StatusBadRequest, qm{"error": "Unable to parse systolic integer", "metric": "systolic", "id": id})
			return
		}
		sys = &s
	}

	if diastolic, ok := singleValue(r.Form["diastolic"]); ok {
		d, err := strconv.Atoi(diastolic)
		if err != nil {
			bp.handleCustomerError(w, http.StatusBadRequest, qm{"error": "Unable to parse diastolic integer", "metric": "diastolic", "id": id})
			return
		}
		dia = &d
	}

	if puls, ok := singleValue(r.Form["pulse"]); ok {
		p, err := strconv.Atoi(puls)
		if err != nil {
			bp.handleCustomerError(w, http.StatusBadRequest, qm{"error": "Unable to parse pulse integer", "metric": "pulse", "id": id})
			return
		}
		pulse = &p
	}

	if note, ok := singleValue(r.Form["notes"]); ok {
		notes = &note
	}

	err := backend.EditMeasurement(bp.db, user.ID, id, sys, dia, pulse, notes)
	if err == backend.NothingToUpdate {
		bp.handleCustomerError(w, http.StatusBadRequest, qm{"error": "Nothing to update"})
		return
	} else if err != nil {
		bp.handleInternalServerError(w, err, qm{"error": "internal server error editing measurement"})
		return
	}
	io.WriteString(w, qm{"ok": true, "id": id}.toJSON())

}

func (bp *bpserver) handleRemoveMeasurements(w http.ResponseWriter, r *http.Request) {
	user := bp.mustGetUser(w, r)
	if user == nil {
		return
	}

	var ids []int
	for _, i := range r.Form["id"] {
		id, err := strconv.Atoi(i)
		if err != nil {
			bp.handleCustomerError(w, http.StatusBadRequest, qm{"error": "could not parse integer"})
			return
		}
		ids = append(ids, id)
	}

	err := backend.RemoveMeasurements(bp.db, user.ID, ids)
	if err != nil {
		bp.handleInternalServerError(w, err, qm{"error": "removing measurements"})
		return
	}

	io.WriteString(w, qm{"ok": true}.toJSON())
}

func (bp *bpserver) handleAddMeasurement(w http.ResponseWriter, r *http.Request) {
	user := bp.mustGetUser(w, r)
	if user == nil {
		return
	}

	systolic := r.Form["systolic"]
	diastolic := r.Form["diastolic"]
	pulse := r.Form["pulse"]
	notes := r.Form["notes"]
	if len(systolic) == 0 || len(diastolic) == 0 || len(pulse) == 0 {
		bp.handleCustomerError(w, http.StatusBadRequest, qm{"error": "must provide systolic, diastolic, and pulse"})
		return
	}
	var note string
	if len(notes) > 0 {
		note = notes[0]
	}
	sys, err := strconv.Atoi(systolic[0])
	if err != nil {
		bp.handleCustomerError(w, http.StatusBadRequest, qm{"error": "unable to parse systolic"})
		return
	}
	dia, err := strconv.Atoi(diastolic[0])
	if err != nil {
		bp.handleCustomerError(w, http.StatusBadRequest, qm{"error": "unable to parse diastolic"})
		return
	}
	pul, err := strconv.Atoi(pulse[0])
	if err != nil {
		bp.handleCustomerError(w, http.StatusBadRequest, qm{"error": "unable to parse pulse"})
		return
	}

	err = backend.AddMeasurement(bp.db, user.ID, sys, dia, pul, note)
	if err != nil {
		bp.handleInternalServerError(w, err, nil)
		return
	}
	io.WriteString(w, qm{"ok": true}.toJSON())
}

func (bp *bpserver) parseUserCreds(r *http.Request) creds {

	ret := creds{}
	u, p := r.Form["user"], r.Form["pass"]
	if len(u) > 0 {
		ret.username = &u[0]
	}
	if len(p) > 0 {
		ret.password = &p[0]
	}

	a := r.Header["Authorization"]
	if len(a) > 0 {
		ret.apikey = &a[0]
	}

	if sid, err := r.Cookie(passwordCookieName); err == nil {
		ret.sessionId = &sid.Value
	} else {
		s := r.Form["session_id"]
		if len(s) > 0 {
			ret.sessionId = &s[0]
		}
	}

	return ret
}

func (bp *bpserver) handleHealthcheck(w http.ResponseWriter, r *http.Request) {
	err := bp.db.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, qm{"ok": false, "errer": err}.toJSON())
		return
	}
	io.WriteString(w, qm{"ok": true, "endpoints": bp.endpoints}.toJSON())
}

// mustGetUser will return the user in `r`'s context if it exists
// If the context does not have a user this function will call ts.handleInternalServerError and return nil.
// To avoid multiple http header writes, the calling function should not write to the header in the case of a nil user
func (bp *bpserver) mustGetUser(w http.ResponseWriter, r *http.Request) *auth.User {

	u := r.Context().Value("user")
	if u == nil {
		bp.handleInternalServerError(w, noUserInContext, nil)
		return nil
	}
	user, ok := u.(*auth.User)
	if !ok {
		bp.handleInternalServerError(w, noUserInContext, nil)
		return nil
	}
	return user
}
func (bp *bpserver) handleUserCreateSessionID(w http.ResponseWriter, r *http.Request) {
	user := bp.mustGetUser(w, r)
	if user == nil {
		return
	}

	//todo detect if creds are invalid vs internal error and return http.StatusUnauthorized
	sid, err := auth.CreateNewSessionID(bp.db, user)
	if err != nil {
		bp.handleInternalServerError(w, err, nil)
		return
	}

	// TODO duplicate code? If you change e.g. session_id to sessionID then you have to update web/handler.go:submitLogin to know it is sessionID
	io.WriteString(w, qm{"ok": true, "session_id": sid}.String())
}

func (bp *bpserver) handleUserCreateApikey(w http.ResponseWriter, r *http.Request) {

	user := bp.mustGetUser(w, r)
	if user == nil {
		return
	}

	apikey, err := auth.CreateNewApikey(bp.db, user)
	if err != nil {
		bp.handleInternalServerError(w, err, nil)
		return
	}
	io.WriteString(w, qm{"apikey": apikey}.toJSON())
}

func (bp *bpserver) aliceParseIncomingUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// allow cross domain AJAX requests
		w.Header().Set("Access-Control-Allow-Origin", "*")

		var errs []error

		c := bp.parseUserCreds(r)
		if c.sessionId != nil {
			user, err := auth.GetUserBySessionId(bp.db, *c.sessionId)
			if err == nil {
				r = r.WithContext(context.WithValue(r.Context(), "user", user))
				next.ServeHTTP(w, r)
				return
			}
			errs = append(errs, err)
		}

		if c.apikey != nil {
			user, err := auth.GetUserByApikey(bp.db, *c.apikey)
			if err == nil {
				r = r.WithContext(context.WithValue(r.Context(), "user", user))
				next.ServeHTTP(w, r)
				return
			}
			errs = append(errs, err)
		}

		if c.username != nil && c.password != nil {
			user, err := auth.GetUserByLogin(bp.db, *c.username, *c.password)
			if err == nil {
				r = r.WithContext(context.WithValue(r.Context(), "user", user))
				next.ServeHTTP(w, r)
				return
			}
			errs = append(errs, err)
		}

		if len(errs) == 0 {
			bp.handleCustomerError(w, http.StatusBadRequest, qm{"error": "no credentials were supplied"})
			return
		}

		errString := ""
		for _, e := range errs {
			errString += e.Error() + ": "
		}

		bp.handleInternalServerError(w, fmt.Errorf(errString), qm{"ok": false, "error": "unable to authorize"})
		return

	})
}

// handleCustomerError does not log
func (bp *bpserver) handleCustomerError(w http.ResponseWriter, httpCode int, customerResponse qm) {
	w.WriteHeader(httpCode)
	if customerResponse != nil {
		if _, ok := customerResponse["ok"]; !ok {
			customerResponse["ok"] = false
		}
		io.WriteString(w, customerResponse.toJSON())
	}
}

// handleInternalServerError logs
func (bp *bpserver) handleInternalServerError(w http.ResponseWriter, err error, customerResponse qm) {

	pc, file, line, ok := runtime.Caller(1) // 0 is _this_ func. 1 is one up the stack
	logErr := qm{"ok": false, "error": err.Error(), "customer_response": customerResponse, "caller": qm{"pc": pc, "file": file, "line": line, "ok": ok}}
	log.Println(logErr.toJSON())

	w.WriteHeader(http.StatusInternalServerError)
	if customerResponse == nil {
		customerResponse = defaultCustomerResponseInternalServerError
	}
	if _, ok := customerResponse["ok"]; !ok {
		customerResponse["ok"] = false
	}
	io.WriteString(w, customerResponse.toJSON())
}

type qm map[string]interface{}

func (q qm) String() string {
	return q.toJSON()
}

func (q qm) toJSON() string {
	return ToJSON(q)
}

// ToJSON returns a the JSON form of obj. If unable to Marshal obj, a JSON error message is returned
// with the %#v formatted string of the object
func ToJSON(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		return fmt.Sprintf(`{"error":"failed to marshal into JSON","obj":%q}`, fmt.Sprintf("%#v", obj))
	}
	return string(b)
}
