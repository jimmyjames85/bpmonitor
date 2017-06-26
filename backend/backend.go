package backend

import (
	"database/sql"
	"fmt"
)

type Measurement struct {
	Id        int64 `json:"id"`
	userid    int64
	Systolic  int `json:"systolic"`
	Diastolic int `json:"diastolic"`
	Pulse     int `json:"pulse"`
	Notes     string `json:"notes"`
	CreatedAt string `json:"created_at"`
}

func AddMeasurement(db *sql.DB, userid int64, systolic, diastolic, pulse int, notes string) error {
	stmt := fmt.Sprintf("INSERT INTO measurements (user_id, systolic, diastolic, pulse, notes) VALUES (?, ?, ?, ?, ?)")
	_, err := db.Exec(stmt, userid, systolic, diastolic, pulse, notes)
	return err
}

func GetMeasurements(db *sql.DB, userid int64) ([]Measurement, error) {

	var ret []Measurement

	rows, err := db.Query("SELECT id, user_id, systolic, diastolic, pulse, notes, created_at FROM measurements WHERE user_id=?", userid)
	if err != nil {
		return ret, err
	}

	defer rows.Close()
	for rows.Next() {
		var m Measurement
		err = rows.Scan(&m.Id, &m.userid, &m.Systolic, &m.Diastolic, &m.Pulse, &m.Notes, &m.CreatedAt)
		if err != nil {
			return ret, err
		}
		ret = append(ret, m)
	}

	err = rows.Err()
	if err != nil {
		return ret, err
	}
	return ret, nil
}
