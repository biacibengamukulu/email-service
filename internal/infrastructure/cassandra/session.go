package cassandra

import (
	"strings"
	"time"

	"github.com/gocql/gocql"
)

type Session struct {
	*gocql.Session
}

func New(hostsCSV, keyspace string) (*Session, error) {
	cluster := gocql.NewCluster(strings.Split(hostsCSV, ",")...)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 5 * time.Second
	cluster.ConnectTimeout = 5 * time.Second
	s, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	return &Session{Session: s}, nil
}

// NewSystem connects to Cassandra using the "system" keyspace so we can create keyspaces/tables.
func NewSystem(hostsCSV string) (*Session, error) {
	return New(hostsCSV, "system")
}
