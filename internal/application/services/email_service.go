package services

import (
	"context"

	"github.com/biangacila/email-service/internal/domain"
	"github.com/biangacila/email-service/internal/infrastructure/libs"
)

type EmailSender interface {
	Send(ctx context.Context, email domain.Email) error
}

type EmailRepository interface {
	Save(ctx context.Context, email domain.Email) error
	UpdateStatus(ctx context.Context, id string, status domain.EmailStatus, retries int) error
}

type EmailService struct {
	sender EmailSender
	repo   EmailRepository
}

func NewEmailService(sender EmailSender, repo EmailRepository) *EmailService {
	return &EmailService{sender: sender, repo: repo}
}

func (s *EmailService) ProcessEmail(ctx context.Context, email domain.Email) {
	err := s.sender.Send(ctx, email)

	if err != nil {
		email.Retries++
		if email.Retries >= 3 {
			s.repo.UpdateStatus(ctx, email.ID, domain.Failed, email.Retries)
			return
		}

		// retry
		libs.Retry(func() error {
			return s.sender.Send(ctx, email)
		})

		s.repo.UpdateStatus(ctx, email.ID, domain.Pending, email.Retries)
		return
	}

	s.repo.UpdateStatus(ctx, email.ID, domain.Sent, email.Retries)
}
