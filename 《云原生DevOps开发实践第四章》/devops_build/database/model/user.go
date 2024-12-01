package model

type User struct {
	Uuid       int    `mapstruct:"uuid,omitempty" db:"uuid"`
	Name       string `mapstruct:"name,omitempty" db:"name"`
	Account    string `mapstruct:"account,omitempty" db:"account"`
	Password   string `mapstruct:"password,omitempty" db:"password"`
	Department string `mapstruct:"department,omitempty" db:"department"`
	Tel        string `mapstruct:"tel,omitempty" db:"tel"`
	Admin      bool   `mapstruct:"admin,omitempty" db:"admin"`
}
