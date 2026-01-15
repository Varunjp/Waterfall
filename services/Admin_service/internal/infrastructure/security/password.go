package security

import "golang.org/x/crypto/bcrypt"

func Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash),[]byte(password))
}

func Hash(password string)(string,error) {
	b,err := bcrypt.GenerateFromPassword([]byte(password),12)
	return string(b),err 
}