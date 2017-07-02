package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int    `json:"id"`
	Username  string `json:"user"`
	password  string
	apikey    *string
	sessionId *string
}

func GetUserBySessionId(db *sql.DB, sessionId string) (*User, error) {
	row := db.QueryRow("select id, username , password , apikey from users where session_id=?", sessionId)
	var apikey sql.NullString
	ret := &User{sessionId: &sessionId}
	err := row.Scan(&ret.ID, &ret.Username, &ret.password, &apikey)
	if err != nil {
		return nil, err
	}
	if apikey.Valid {
		ret.apikey = &apikey.String
	}
	return ret, nil
}

func GetUserByApikey(db *sql.DB, apikey string) (*User, error) {
	row := db.QueryRow("select id, username , password , session_id from users where apikey=?", apikey)
	var sessionId sql.NullString
	ret := &User{sessionId: &apikey}
	err := row.Scan(&ret.ID, &ret.Username, &ret.password, &sessionId)
	if err != nil {
		return nil, err
	}
	if sessionId.Valid {
		ret.sessionId = &sessionId.String
	}
	return ret, nil
}

func GetUserByLogin(db *sql.DB, username, password string) (*User, error) {

	var sessionId, apikey sql.NullString
	ret := &User{Username: username}

	row := db.QueryRow("select id, session_id, apikey, password from users where username=?", username)
	err := row.Scan(&ret.ID, &sessionId, &apikey, &ret.password)
	if err != nil {
		return nil, err
	}
	if sessionId.Valid {
		ret.sessionId = &sessionId.String
	}
	if apikey.Valid {
		ret.apikey = &apikey.String
	}
	err = bcrypt.CompareHashAndPassword([]byte(ret.password), []byte(password))
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func ClearSessionID(db *sql.DB, user *User) error {
	_, err := db.Exec("UPDATE users set session_id=NULL WHERE username=? AND id=?", user.Username, user.ID)
	return err
}

func CreateNewSessionID(db *sql.DB, user *User) (string, error) {
	sessionID, err := newUUID()
	if err != nil {
		return "", fmt.Errorf("uuid creation err: %v", err)
	}

	s64 := base64.StdEncoding.EncodeToString([]byte(sessionID))
	uid64 := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", user.ID)))

	sessionID = fmt.Sprintf("BP.%s.%s", s64, uid64)
	sessionID = strings.Replace(sessionID, "=", "", -1)
	sessionID = strings.Replace(sessionID, "\n", "", -1)

	_, err = db.Exec("UPDATE users SET session_id=? WHERE username=? AND id=?", sessionID, user.Username, user.ID)
	if err != nil {
		return "", err
	}
	return sessionID, nil
}

func CreateNewApikey(db *sql.DB, user *User) (string, error) {

	apikey, err := newUUID()
	if err != nil {
		return "", fmt.Errorf("uuid creation err: %v", err)
	}

	api64 := base64.StdEncoding.EncodeToString([]byte(apikey))
	uid64 := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", user.ID)))

	apikey = fmt.Sprintf("BP.%s.%s", api64, uid64)
	apikey = strings.Replace(apikey, "=", "", -1)
	apikey = strings.Replace(apikey, "\n", "", -1)

	_, err = db.Exec("UPDATE users SET apikey=? WHERE username=? AND id=?", apikey, user.Username, user.ID)
	if err != nil {
		return "", err
	}

	return apikey, nil
}

var UserExists = errors.New("User already exists")

func CreateUser(db *sql.DB, username, password string) (*User, error) {

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, string(hashed))
	if err != nil {
		if strings.Index(err.Error(), "Duplicate entry") == -1 {
			return nil, err
		}
		return nil, UserExists
	}

	// TODO what if the insert works, but getuser returns err?
	return GetUserByLogin(db, username, password)
}

// newUUID generates a random UUID according to RFC 4122
func newUUID() (string, error) {
	// copy pasta from a google search
	// https://play.golang.org/p/uEIKweC-kp
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}
