package mtnsa

type Sms interface {
	Send(phone, message string)
}
type SmsClient struct {
	username string
	password string
}

func NewSmsClient(username string, password string) *SmsClient {
	return &SmsClient{username: username, password: password}
}
