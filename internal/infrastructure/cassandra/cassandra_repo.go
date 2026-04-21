package cassandra

import (
	"context"
	"log"

	"github.com/biangacila/email-service/internal/domain"
	"github.com/gocql/gocql"
)

type CassandraRepo struct {
	session *gocql.Session
}

func NewCassandraRepo(hosts []string) *CassandraRepo {
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = "email_service"
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}

	return &CassandraRepo{session: session}
}

func (r *CassandraRepo) Save(ctx context.Context, email domain.Email) error {
	return r.session.Query(`INSERT INTO emails (id, to, subject, status, retries) VALUES (?, ?, ?, ?, ?)`,
		email.ID, email.Receiver, email.Subject, email.Status, email.Retries).Exec()
}

func (r *CassandraRepo) UpdateStatus(ctx context.Context, id string, status domain.EmailStatus, retries int) error {
	return r.session.Query(`UPDATE emails SET status=?, retries=? WHERE id=?`,
		status, retries, id).Exec()
}
