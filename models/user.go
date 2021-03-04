package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
)

var (
 	ErrUserInvalidPassword = errors.New("invalid password")
 	ErrUserNotFound = errors.New("user not found")
)

type User struct {
	ID       int64   `db:"id" json:"recid"`
	UnitID   int64   `db:"auth_unit_id" json:"unitid"`
	GroupsID []int64 `json:"groupsid"`
	Name     string  `db:"name" json:"name"`
	Password string  `db:"password" json:"password"`
	Blocked  bool    `db:"blocked" json:"blocked"`
}

func UserGetByID(id int64) (*User, error) {
	user := new(User)
	row := Db.QueryRowx("SELECT `id`, `auth_unit_id`, `name`, LOCATE('!', `password`) AS blocked FROM `auth_user` WHERE `id`=?", id)
	if row.Err() == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	err := row.StructScan(user)
	if err != nil {
		return nil, err
	}
	err = user.getGroups()
	if err != nil {
		return nil, err
	}
	return user, nil
}

func UserGetByName(name string) (*User, error) {
	user := new(User)
	row := Db.QueryRowx("SELECT `id`, `auth_unit_id`, `name`, LOCATE('!', `password`) AS blocked FROM `auth_user` WHERE `name`=?", name)
	if row.Err() == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	err := row.StructScan(user)
	if err != nil {
		return nil, err
	}
	err = user.getGroups()
	if err != nil {
		return nil, err
	}
	return user, nil
}

func UserGetByNameAndPassword(name, password string) (*User, error) {
	user := new(User)
	row := Db.QueryRowx("SELECT `id`, `auth_unit_id`, `name`, `password`, LOCATE('!', `password`) AS blocked FROM `auth_user` WHERE `name`=?", name)
	if row.Err() != nil {
		return nil, row.Err()
	}
	err := row.StructScan(user)
	if  err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	dbHash := user.Password
	user.Password = password
	passwordHash, err := user.passwordHash()
	if err != nil {
		return nil, err
	}
	if passwordHash != dbHash {
		return user, ErrUserInvalidPassword
	}
	err = user.getGroups()
	if err != nil {
		return nil, err
	}
	return user, nil
}
func (u *User) getGroups() error {
	q, err := Db.Query("SELECT `group_id` FROM `auth_user_group` WHERE `auth_user_id`=?", u.ID)
	if err != nil {
		return err
	}
	defer q.Close()
	for q.Next() {
		var i int64
		err = q.Scan(&i)
		if err != nil {
			return err
		}
		u.GroupsID = append(u.GroupsID, i)
	}
	return nil
}

func (u User) IsAdmin() bool {
	// admins has group 0
	return u.UnitID == 0
}

func (u *User) Add() error {
	var err error

	row := Db.QueryRow("SELECT 1 FROM `auth_user` WHERE `name`=?", u.Name)
	if row.Err() != nil {
		return err
	}
	err = row.Scan()
	if err != sql.ErrNoRows {
		return errors.New("user name already exist")
	}

	if u.Password == "" {
		return errors.New("password can not be empty")
	}
	if u.Name == "" {
		return errors.New("name can not be empty")
	}

	passwordHash, err := u.passwordHash()
	if err != nil {
		return err
	}

	s, err := Db.Exec("INSERT INTO `auth_user`(`auth_unit_id`, `name`, `password`) VALUES (?, ?, ?)", u.UnitID, u.Name, passwordHash)
	if err != nil {
		return err
	}
	u.ID, err = s.LastInsertId()
	if err != nil {
		return err
	}

	return u.Update()
}

func (u User) Update() error {
	var err error
	// Update password
	if u.Password != "" {
		passwordHash, err := u.passwordHash()
		if err != nil {
			return err
		}
		_, err = Db.Exec("UPDATE `auth_user` SET `password`=? WHERE `id`=?", passwordHash, u.ID)
		if err != nil {
			return err
		}
	}

	// Update block status and unit
	if u.Blocked {
		_, err = Db.Exec("UPDATE `auth_user` SET `auth_unit_id`=?, `password`=CONCAT('!', REPLACE(`password`, '!', '')) WHERE `id`=?", u.UnitID, u.ID)
	} else {
		_, err = Db.Exec("UPDATE `auth_user` SET `auth_unit_id`=?, `password`=REPLACE(`password`, '!', '') WHERE `id`=?", u.UnitID, u.ID)
	}
	if err != nil {
		return err
	}

	// Update user groups
	if u.GroupsID != nil {
		_, err := Db.Exec("DELETE FROM `auth_user_group` WHERE `auth_user_id`=?", u.ID)
		if err != nil {
			return err
		}
		tx, err := Db.Prepare("INSERT INTO `auth_user_group` (`auth_user_id`, `group_id`) VALUES (?, ?)")
		if err != nil {
			return err
		}
		defer tx.Close()
		for _, g := range u.GroupsID {
			_, err = tx.Exec(u.ID, g)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (u *User) passwordHash() (string, error) {
	hash := sha256.New()
	_, err := hash.Write([]byte(u.Password))
	u.Password = ""
	if err != nil {
		return "", err
	}
	md := hash.Sum(nil)
	passwordHash := hex.EncodeToString(md)
	if u.Blocked {
		passwordHash = "!" + passwordHash
	}
	return passwordHash, nil
}
