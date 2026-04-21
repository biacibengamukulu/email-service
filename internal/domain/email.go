package domain

type EmailStatus string

const (
	Pending EmailStatus = "PENDING"
	Sent    EmailStatus = "SENT"
	Failed  EmailStatus = "FAILED"
)

type Email struct {
	ID       string
	From     string
	Receiver []string
	Subject  string
	HTML     string
	Status   EmailStatus
	Retries  int
}
