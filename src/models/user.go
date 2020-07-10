package models

type User struct {
	UserId string
	Activated bool
	Role string
}

type UserRegistration struct {
	FirstName string `json:"firstname"`
	LastName string `json:"lastname"`
	Email string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	PrivacyLevel int `json:"privacylevel"`
	// TODO: implement nationality
}


