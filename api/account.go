package api

import (
	"errors"
	"gonder/models"
	"unicode"
)

func account(req request) (js []byte, err error) {
	switch req.Cmd {
	case "changePassword":
		err = changeAccountPassword(req.auth.user.ID, req.Record.Password, req.Record.NewPassword, req.Record.ConfirmPassword)
	default:
		err = errors.New("command not found")
	}
	return js, err
}

func changeAccountPassword(userID int64, currentPassword, newPassword, confirmPassword string) error {
	user, err := models.UserGetByID(userID)
	if err != nil {
		return err
	}
	_, err = models.UserGetByNameAndPassword(user.Name, currentPassword)
	if err != nil {
		return err
	}
	if newPassword != confirmPassword {
		return errors.New("password confirmation does not match")
	}
	err = validPassword(newPassword)
	if err != nil {
		return err
	}
	user.Password = newPassword
	return user.Update()
}

func validPassword(p string) error {
	if len(p) < 8 {
		return errors.New("the password must be longer than 8 characters")
	}

	var (
		uppercasePresent   bool
		lowercasePresent   bool
		numberPresent      bool
		specialCharPresent bool
	)

	for _, r := range p {
		switch {
		case unicode.IsNumber(r):
			numberPresent = true
		case unicode.IsUpper(r):
			uppercasePresent = true
		case unicode.IsLower(r):
			lowercasePresent = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			specialCharPresent = true
		}
	}

	if !lowercasePresent {
		return errors.New("lowercase letter missing in password")
	}
	if !uppercasePresent {
		return errors.New("uppercase letter missing in password")
	}
	if !numberPresent {
		return errors.New("numeric character missing in password")
	}
	if !specialCharPresent {
		return errors.New("special character missing")
	}

	return nil
}
