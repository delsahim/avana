package utils

const (
	ReadRequestError string ="Failed to read the request body"
	ExistingDataError string = "Record already exists in the database"
	HashingError string = "Unable to hash the password"
	CreateRecordError string =  "Error trying to create Record"
	DatabaseCallError string = "Unable to read the data from database"
	CredentialsError string = "Wrong credentials provided"
	TokenError string = "Error generating token"
	UpdateRecordError string = "Error updating the record"
	ExpiresVerificationError string = "Unable to change the password"
	ValidateTokenError string = "Incorrect authentication token"
	IncorrecPermission string = "This user does not have required permission"
	PriceError string = "Cannot add Price to a free event"
	TicketTimeError string = "Ticket sales must end before the event date"
	DeleteRecordError string = "Error trying to delete record"
	TicketExpiredError string = "Ticket has expired"
	TicketAmountError string = "Ticket amount exceeds limit"
	EventHeldError string = "This event has not held"
)