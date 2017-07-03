package backend

import (
	"bytes"
	"database/sql"
	"strings"

	"time"

	"github.com/pkg/errors"
)

type Measurement struct {
	Id        int64 `json:"id"`
	userid    int
	Systolic  int    `json:"systolic"`
	Diastolic int    `json:"diastolic"`
	Pulse     int    `json:"pulse"`
	Notes     string `json:"notes"`
	CreatedAt int64  `json:"created_at"`
}

var NothingToUpdate = errors.New("nothing to update")

func AddMeasurement(db *sql.DB, userid int, systolic, diastolic, pulse int, notes string) error {
	_, err := db.Exec("INSERT INTO measurements (user_id, systolic, diastolic, pulse, notes) VALUES (?, ?, ?, ?, ?)", userid, systolic, diastolic, pulse, notes)
	return err
}

func EditMeasurement(db *sql.DB, userid int, id int, systolic, diastolic, pulse *int, notes *string, createdAt *time.Time) error {

	var args []interface{}

	var stmt bytes.Buffer
	stmt.WriteString("UPDATE measurements SET ")

	if systolic != nil {
		if len(args) > 0 {
			stmt.WriteString(", ")
		}
		stmt.WriteString("systolic=? ")
		args = append(args, *systolic)
	}
	if diastolic != nil {

		if len(args) > 0 {
			stmt.WriteString(", ")
		}
		stmt.WriteString("diastolic=? ")
		args = append(args, *diastolic)
	}
	if pulse != nil {
		if len(args) > 0 {
			stmt.WriteString(", ")
		}
		stmt.WriteString("pulse=? ")
		args = append(args, *pulse)
	}
	if notes != nil {
		if len(args) > 0 {
			stmt.WriteString(", ")
		}
		stmt.WriteString("notes=? ")
		args = append(args, *notes)
	}
	if createdAt != nil {
		if len(args) > 0 {
			stmt.WriteString(", ")
		}
		stmt.WriteString("created_at=? ")
		args = append(args, *createdAt)
	}

	if len(args) == 0 {
		return NothingToUpdate
	}

	stmt.WriteString(" WHERE user_id=? AND id=?")
	args = append(args, userid)
	args = append(args, id)

	_, err := db.Exec(stmt.String(), args...)
	return err
}

func RemoveMeasurements(db *sql.DB, userid int, ids []int) error {

	if len(ids) == 0 {
		return nil
	}
	var args []interface{}
	args = append(args, userid)
	for _, id := range ids {
		args = append(args, id)
	}
	stmt := "DELETE FROM measurements WHERE user_id=? and id IN (?" + strings.Repeat(", ?", len(ids)-1) + ")"
	_, err := db.Exec(stmt, args...)
	return err

}
func GetMeasurements(db *sql.DB, userid int) ([]Measurement, error) {

	var ret []Measurement

	rows, err := db.Query("SELECT id, user_id, systolic, diastolic, pulse, notes, created_at FROM measurements WHERE user_id=? ORDER BY created_at DESC", userid)
	if err != nil {
		return ret, err
	}

	defer rows.Close()
	for rows.Next() {
		var m Measurement
		var createdAt time.Time
		err = rows.Scan(&m.Id, &m.userid, &m.Systolic, &m.Diastolic, &m.Pulse, &m.Notes, &createdAt)
		if err != nil {
			return ret, err
		}
		m.CreatedAt = createdAt.UTC().Unix()
		ret = append(ret, m)
	}

	err = rows.Err()
	if err != nil {
		return ret, err
	}
	return ret, nil
}
