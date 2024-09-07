package users

type CreateUserSchema struct {
	Email string
	Password string
	FirstName string
	LastName string
}

type UserCredentials struct {
	Email string
	Password string
}

type OtpCredentials struct {
	Email string
	Otp uint
}

type UserEmail struct {
	Email string
}