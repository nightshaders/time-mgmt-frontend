package authenticator

import (
	"encoding/csv"
	"errors"
	"os"
)

type User struct {
	Username string
	Role     string
	Password string
}

type Session struct {
	user   User
	active bool
}

// returns whether the given file exists or not
func isFile(path string) bool {
	_, err := os.Stat(path)

	if err == nil {
		return true
	}

	return false
}

func getUserPasswordList(file string) (map[string]User, error) {
	if !isFile(file) {
		return nil, errors.New("password file does not exist")
	}

	f, err := os.Open(file)

	if err != nil {
		return nil, errors.New("could not open password file")
	}

	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1

	rows, err := reader.ReadAll()

	if err != nil {
		return nil, errors.New("could not read password file")
	}

	var userPass = make(map[string]User)

	for _, each := range rows {
		var username, password, role string
		if len(each) == 3 {
			username, password, role = each[0], each[1], each[2]
		} else if len(each) == 2 {
			username, password, role = each[0], each[1], "Regular"
		} else {
			continue
		}

		userInfo := User{Username: username, Password: password, Role: role}
		userPass[username] = userInfo
	}

	return userPass, nil
}

func IsRegisteredUser(u string) bool {
	users, err := getUserPasswordList("passwd")

	if err != nil {
		return false
	}

	_, ok := users[u]

	return ok
}

func IsValidUserPass(u string, p []byte) bool {
	return IsRegisteredUser(u) && string(p) == getPassword(u)
}

func getPassword(u string) string {
	users, err := getUserPasswordList("passwd")

	if err != nil {
		return ""
	}

	user, ok := users[u]
	if !ok {
		return ""
	}

	return user.Password
}

func OpenSession(name, pass string, users map[string]User) (Session, error) {
	user, ok := users[name]
	if !ok {
		return Session{}, errors.New("user does not exist")
	}

	var session Session
	if users[name].Password == pass {
		session = Session{user: user, active: true}
	} else {
		session = Session{user: user, active: false}
	}

	return session, nil
}

func IsRegularUser(name string, users map[string]User) (bool, error) {
	user, ok := users[name]

	if !ok {
		return false, errors.New("user not found")
	}

	return user.Role == "Regular", nil
}

func IsAdminUser(name string, users map[string]User) (bool, error) {
	user, ok := users[name]

	if !ok {
		return false, errors.New("user not found")
	}

	return user.Role == "Admin", nil
}

func IsLoggedIn(name string, session Session) bool {
	return session.user.Username == name && session.active
}

func IsValidNewUsername(name string) (bool, error) {
	users, err := getUserPasswordList("passwd")

	if err != nil {
		return false, errors.New("could not get list of registered users")
	}

	_, ok := users[name]

	if ok {
		return false, errors.New("username already taken")
	}

	return true, nil
}

func RegisterUser(name, password string) error {
	users, err := getUserPasswordList("passwd")
	if err != nil {
		return errors.New("could not read user list")
	}

	user := User{Username: name, Password: password, Role: "Regular"}
	users[name] = user

	err = updateUserList(users)
	return err
}

func updateUserList(users map[string]User) error {
	err := os.Remove("passwd")

	if err != nil {
		return errors.New("could not remove password file")
	}

	f, err := os.OpenFile("passwd", os.O_WRONLY|os.O_CREATE, 0600)

	if err != nil {
		return errors.New("could not open password file")
	}
	defer f.Close()

	w := csv.NewWriter(f)

	records := make([][]string, 0)
	for _, info := range users {
		record := make([]string, 0)
		record = append(record, info.Username)
		record = append(record, info.Password)
		record = append(record, info.Role)
		records = append(records, record)
	}

	err = w.WriteAll(records)

	if err != nil {
		return errors.New("could not write password file")
	}

	return nil
}

func DeleteUser(user string) error {
	users, err := getUserPasswordList("passwd")

	if err != nil {
		return errors.New("could not open user list")
	}

	_, ok := users[user]

	if !ok {
		return errors.New("cannot erase user. does not exist")
	}

	delete(users, user)

	err = updateUserList(users)

	return err
}
