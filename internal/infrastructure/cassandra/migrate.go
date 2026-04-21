package cassandra

import (
	"fmt"
	"strings"
)

// EnsureKeyspace creates the keyspace if it doesn't exist.
// NOTE: Uses SimpleStrategy by default. For multi-DC production, you may switch to NetworkTopologyStrategy.
func EnsureKeyspace(hostsCSV, keyspace string, replicationFactor int) error {
	if replicationFactor <= 0 {
		replicationFactor = 1
	}

	sys, err := NewSystem(hostsCSV)
	if err != nil {
		return err
	}
	defer sys.Close()

	ks := strings.ToLower(keyspace)
	q := fmt.Sprintf(
		"CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class':'SimpleStrategy','replication_factor':%d};",
		ks,
		replicationFactor,
	)
	return sys.Query(q).Exec()
}

// EnsureTables creates tables (and indexes) if they don't exist.
func EnsureTables(hostsCSV, keyspace string, ddls []string) error {
	sess, err := New(hostsCSV, keyspace)
	if err != nil {
		return err
	}
	defer sess.Close()

	for _, ddl := range ddls {
		ddl = strings.TrimSpace(ddl)
		if ddl == "" {
			continue
		}
		if err := sess.Query(ddl).Exec(); err != nil {
			return err
		}
	}
	return nil
}

// BootstrapKeyspaceAndTables is a convenience for keyspace + table creation.
func BootstrapKeyspaceAndTables(hostsCSV, keyspace string, replicationFactor int, ddls []string) error {
	if err := EnsureKeyspace(hostsCSV, keyspace, replicationFactor); err != nil {
		return err
	}
	return EnsureTables(hostsCSV, keyspace, ddls)
}

//ssh safer "docker exec -it eff-node1 bash cqlsh CREATE KEYSPACE IF NOT EXISTS email_service WITH replication = {'class':'SimpleStrategy','replication_factor':1};"
