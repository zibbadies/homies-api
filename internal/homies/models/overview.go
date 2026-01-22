package models

type Overview struct {
	User  User              `json:"user"`
	House House             `json:"house"`
	Lists map[string][]Item `json:"lists"`
}
