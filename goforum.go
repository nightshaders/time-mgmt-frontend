package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
)

type User struct {
	username string
	role     string
	password string
}

type Session struct {
	user   User
	active bool
}

// exists returns whether the given file or directory exists or not
func doesFileExist(path string) bool {
	_, err := os.Stat(path)

	if err == nil {
		return true
	}

	return false
}

func readPasswordFile(file string) (map[string]User, error) {
	if !doesFileExist(file) {
		return nil, errors.New("password file does not exist")
	}

	csvfile, err := os.Open(file)

	if err != nil {
		return nil, errors.New("could not open password file")
	}

	defer csvfile.Close()

	reader := csv.NewReader(csvfile)

	reader.FieldsPerRecord = -1 // see the Reader struct information below

	rawCSVdata, err := reader.ReadAll()

	if err != nil {
		return nil, errors.New("could not read password file")
	}

	var userPass = make(map[string]User)

	for _, each := range rawCSVdata {
		var username, password, role string
		if len(each) == 3 {
			username, password, role = each[0], each[1], each[2]
		} else {
			username, password, role = each[0], each[1], "Regular"
		}

		userInfo := User{username: username, password: password, role: role}
		userPass[username] = userInfo
	}

	return userPass, nil
}

func isRegisteredUser(n string, users map[string]User) bool {
	_, ok := users[n]

	return ok
}

func isPasswordValid(b []byte, up string) bool {
	if string(b) == up {
		return true
	}

	return false
}

func createSession(name, pass string, users map[string]User) (Session, error) {
	user, ok := users[name]
	if !ok {
		return Session{}, errors.New("user does not exist")
	}

	var session Session
	if users[name].password == pass {
		session = Session{user: user, active: true}
	} else {
		session = Session{user: user, active: false}
	}

	return session, nil
}

func isRegularUser(name string, users map[string]User) (bool, error) {
	user, ok := users[name]

	if !ok {
		return false, errors.New("user not found")
	}

	return user.role == "Regular", nil
}

func isAdminUser(name string, users map[string]User) (bool, error) {
	user, ok := users[name]

	if !ok {
		return false, errors.New("user not found")
	}

	return user.role == "Admin", nil
}

func promptUser() int32 {
	var c int32
	fmt.Println("Menu")
	fmt.Println("===========")
	fmt.Println("1. Logout")
	fmt.Scanf("%d", &c)
	return c
}

func isLoggedIn(name string, session Session) bool {
	return session.user.username == name && session.active
}

func loginPrompt() int32 {
	var c int32
	fmt.Println("Menu")
	fmt.Println("===========")
	fmt.Println("1. Sign in")
	fmt.Println("2. Create a new account")
	fmt.Println("3. Quit")
	fmt.Scanf("%d", &c)
	return c
}

func initialChoice(choice int32) {
	promptUser()
}

func createUser(name string) (bool, error) {
	users, err := readPasswordFile("passwd")
	if err != nil {
		return false, errors.New("could not get list of registered users")
	}

	_, ok := users[name]

	if ok {
		return false, errors.New("username already taken")
	}

	return true, nil
}

func createUserPassword(name, password string) (error) {
	users, err := readPasswordFile("passwd")
	if err != nil {
		return errors.New("could not read user list")
	}

	user := User{username: name, password: password, role: "Regular"}
	users[name] = user

	err = writeUserFile(users)
	return err
}

func writeUserFile(users map[string]User) (error) {
	f, err := os.OpenFile("passwd", os.O_WRONLY, 0600)

	if err != nil {
		return errors.New("could not open password file")
	}

	defer f.Close()

	w := csv.NewWriter(f)

	records := make([][]string, 0)
	for _, info := range users {
		record := make([]string, 0)
		record = append(record, info.username)
		record = append(record, info.password)
		record = append(record, info.role)
		records = append(records, record)
	}

	err = w.WriteAll(records)

	if err != nil {
		return errors.New("could not write password file")
	}

	return nil
}

func main() {
	users, err := readPasswordFile("passwd")

	if err != nil {
		panic("Could not open password file")
	}

	vu := false
	vp := false
	for !vu || !vp {
		var u string
		fmt.Print("Username: ")
		fmt.Scanf("%s", &u)

		fmt.Print("Enter password: ")
		pass, err := terminal.ReadPassword(0)
		fmt.Println()

		if err != nil {
			panic("Could not obtain password")
		}

		vu = isRegisteredUser(u, users)

		user, ok := users[u]

		if !ok {
			vp = false
		} else {
			vp = isPasswordValid([]byte(pass), user.password)
		}
	}

	sel := promptUser()
	for sel != 1 {
		sel = promptUser()
	}
}
