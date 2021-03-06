package bpmonitor

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"image/color"
	"io"
	"log"
	"math"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"
	"github.com/jimmyjames85/bpmonitor/backend"
	"github.com/jimmyjames85/bpmonitor/backend/auth"
	"github.com/pkg/errors"
)

var (
	noUserInContext                            = fmt.Errorf("no user in context")
	defaultCustomerResponseInternalServerError = qm{"error": "Internal Server Error: please contact jimmyjames85"}
)

type creds struct {
	username  *string
	password  *string
	apikey    *string
	sessionId *string
}

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

// aliceParseIncomingUser attempts to parse the incoming user credentials and
// validate authorization. If successful, this function passes along the context "user"
// set to a valid *auth.User object, in r.Context
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

// mustGetUser will retrieve and return the *auth.user in r.context
// If the user does not exits this function calls ts.handleInternalServerError and returns nil.
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

func (bp *bpserver) handleAdminCreateUser(w http.ResponseWriter, r *http.Request) {

	// todo should this method be served on a different port?

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

	if err := validateUsernameRequirements(u[0]); err != nil {
		bp.handleCustomerError(w, http.StatusBadRequest, qm{"error": err.Error()})
		return
	}

	if err := validatePasswordRequirements(p[0]); err != nil {
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

	io.WriteString(w, qm{"ok": true, "user": user.Username}.toJSON())
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

func (bp *bpserver) handleEditMeasurements(w http.ResponseWriter, r *http.Request) {
	user := bp.mustGetUser(w, r)
	if user == nil {
		return
	}

	var id int
	var sys, dia, pulse *int
	var notes *string
	var created_at *time.Time

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

	if ca, ok := singleValue(r.Form["created_at"]); ok {
		ts, err := strconv.ParseInt(ca, 10, 64)
		if err != nil {
			bp.handleCustomerError(w, http.StatusBadRequest, qm{"error": "Unable to parse created_at date", "metric": "created_at", "id": id})
			return
		}
		ca := time.Unix(ts, 0)
		created_at = &ca
	}

	err := backend.EditMeasurement(bp.db, user.ID, id, sys, dia, pulse, notes, created_at)
	if err == backend.NothingToUpdate {
		bp.handleCustomerError(w, http.StatusBadRequest, qm{"error": "Nothing to update"})
		return
	} else if err != nil {
		bp.handleInternalServerError(w, err, qm{"error": "internal server error editing measurement"})
		return
	}
	io.WriteString(w, qm{"ok": true, "id": id}.toJSON())

}

func (bp *bpserver) handleGetMeasurements(w http.ResponseWriter, r *http.Request) {
	user := bp.mustGetUser(w, r)
	if user == nil {
		return
	}

	// TODO add date range
	from := time.Unix(0, 0)
	to := time.Now()

	measurements, err := backend.GetMeasurements(bp.db, user.ID, from, to)
	if err != nil {
		bp.handleInternalServerError(w, err, qm{"error": "retrieving measurements"})
		return
	}

	for _, m := range measurements {
		m.Notes = html.EscapeString(m.Notes)
	}

	io.WriteString(w, qm{"ok": true, "measurements": measurements}.toJSON())

}

func (bp *bpserver) handleGraphMeasurements(w http.ResponseWriter, r *http.Request) {

	// TODO add date range

	from := time.Unix(0, 0)
	to := time.Now()

	user := bp.mustGetUser(w, r)
	if user == nil {
		return
	}

	// tz_offset should be in hours, and is a way to graph in different timezones
	// this function graphs times in UTC by default
	tzOffset := 0
	if len(r.Form["tz_offset"]) > 0 {
		o, err := strconv.Atoi(r.Form["tz_offset"][0])
		if err != nil {
			bp.handleCustomerError(w, http.StatusBadRequest, qm{"error": "unable to parse tz_offset"})
			return
		}
		tzOffset = o
	}

	measurements, err := backend.GetMeasurements(bp.db, user.ID, from, to)
	if err != nil {
		bp.handleInternalServerError(w, err, qm{"error": "retrieving measurements"})
		return
	}

	if len(measurements) == 0 {
		bp.handleInternalServerError(w, errors.New("aint nothin to plot"), qm{"error": "there are no values to plot"})
	}

	sys := make(plotter.XYs, len(measurements))
	dia := make(plotter.XYs, len(measurements))
	pulse := make(plotter.XYs, len(measurements))
	notes := make(plotter.XYs, len(measurements))

	minX := float64(measurements[0].CreatedAt)
	maxX := minX
	minY := float64(measurements[0].Pulse) // pulse is arbitrary, but we must use an actual measurement.
	maxY := minY                           // We can't assume zero is the min or max

	for i := range measurements {
		// we get measurements in reverse order from the DB,
		// so must we iterate over measurements in reverse
		m := measurements[len(measurements)-i-1]

		x := float64(time.Unix(m.CreatedAt, 0).Add(time.Duration(tzOffset) * time.Hour).Unix())
		s, d, p := float64(m.Systolic), float64(m.Diastolic), float64(m.Pulse)

		minX = math.Min(minX, x)
		maxX = math.Max(maxX, x)
		minY = math.Min(math.Min(math.Min(minY, s), d), p)
		maxY = math.Max(math.Max(math.Max(maxY, s), d), p)

		sys = append(sys, point{x, s})
		dia = append(dia, point{x, d})
		pulse = append(pulse, point{x, p})

		if len(m.Notes) > 0 {
			notes = append(notes, point{x, s})
		}

	}

	// grid
	g := plotter.NewGrid()

	// systolic
	sl, sp, _ := plotter.NewLinePoints(sys)
	sl.Color = color.RGBA{0, 0, 165, 255}
	sp.GlyphStyle.Radius *= 0.5

	// notes
	np, _ := plotter.NewScatter(notes)
	np.GlyphStyle.Shape = draw.CircleGlyph{}
	np.GlyphStyle.Radius *= 1.5
	np.GlyphStyle.Color = color.RGBA{225, 225, 0, 0}

	// diastolic
	dl, dp, _ := plotter.NewLinePoints(dia)
	dl.Color = color.RGBA{0, 165, 0, 255}
	dp.GlyphStyle.Radius *= 0.5

	// pulse
	pl, pp, _ := plotter.NewLinePoints(pulse)
	pl.Color = color.RGBA{165, 0, 0, 200}
	pl.Width *= 0.75
	pp.GlyphStyle.Radius *= 0.5

	p, err := plot.New()
	if err != nil {
		bp.handleInternalServerError(w, err, qm{"error": "internal error with gonum/plot"})
		return
	}
	p.Add(g, dl, dp, pl, pp, sl, sp, np)

	pl.Dashes = []vg.Length{vg.Points(5), vg.Points(5)}

	if err != nil {
		bp.handleInternalServerError(w, err, qm{"error": "internal error with gonum/plot"})
		return
	}
	p.Title.Text = fmt.Sprintf("BPMonitor: %s", user.Username)
	p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-02\n15:04"}
	imgWidth := 12 * vg.Inch // 20
	imgHeigth := 4 * vg.Inch // 6

	// padding is weird
	padX := (maxX - minX) / 10
	maxX += padX
	padY := (maxY - minY) / 20
	maxY += padY
	p.X.Padding = imgWidth / 20
	p.Y.Padding = imgWidth / 20
	p.Y.Min, p.Y.Max = minY, maxY
	p.X.Min, p.X.Max = minX, maxX
	p.Legend.Add("Systolic", sl)
	p.Legend.Add("Diastolic", dl)
	p.Legend.Add("Pulse", pl)
	p.Legend.Add("Note", np)

	wr, err := p.WriterTo(imgWidth, imgHeigth, "png")
	if err != nil {
		bp.handleInternalServerError(w, err, qm{"error": "internal error with gonum/plot"})
		return
	}

	_, err = wr.WriteTo(w)
	if err != nil {
		bp.handleInternalServerError(w, err, qm{"error": "internal error with gonum/plot"})
		return
	}

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

func (bp *bpserver) handleHealthcheck(w http.ResponseWriter, r *http.Request) {
	err := bp.db.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, qm{"ok": false, "errer": err}.toJSON())
		return
	}
	io.WriteString(w, qm{"ok": true, "endpoints": bp.endpoints}.toJSON())
}

func (bp *bpserver) handleUserCreateSessionID(w http.ResponseWriter, r *http.Request) {
	user := bp.mustGetUser(w, r)
	if user == nil {
		return
	}

	sid, err := auth.CreateNewSessionID(bp.db, user)
	if err != nil {
		bp.handleInternalServerError(w, err, nil)
		return
	}

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

func validatePasswordRequirements(pass string) error {
	if len(pass) < 5 {
		return errors.New("password must have at least five characters")
	}
	return nil
}

func validateUsernameRequirements(username string) error {

	if len(username) < 3 {
		return errors.New("username must have at least three characters")
	}

	return nil
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

//TODO there has GOT to be an equivalent type in github.com/gonum/plot
type point struct{ X, Y float64 }

func singleValue(values []string) (string, bool) {
	if len(values) > 0 {
		return values[0], true
	}
	return "", false
}

// I may not need this
const passwordCookieName = "eWVrc2loV2hzYU1ydW9TZWVzc2VubmVUeXRpbGF1UWRuYXJCNy5vTmRsT2VtaXRkbE9zJ2xlaW5hRGtjYUoK"
