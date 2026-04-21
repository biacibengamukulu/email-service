package config

import (
	"os"
	"strings"

	"github.com/biangacila/email-service/pkg/constants"
)

type Config struct {
	ServiceName string

	HTTPPort string

	CassandraHosts    string
	CassandraKeyspace string

	KafkaBrokers string
	KafkaGroupID string

	DropboxRefreshToken string
	ResendToken         string

	SmsGatewayUsername string
	SmsGatewayPassword string
	SmsGatewayUrl      string
}

func Getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func sanitizeServiceName(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	return s
}

func Load(serviceName string) Config {
	svc := sanitizeServiceName(serviceName)

	defaultKeyspace := constants.CASSANDRA_KEYSPACE

	return Config{
		ServiceName: serviceName,
		HTTPPort:    Getenv("HTTP_PORT", constants.HTTP_PORT),

		CassandraHosts:    Getenv("CASSANDRA_HOSTS", constants.CASSANDRA_HOSTS),
		CassandraKeyspace: Getenv("CASSANDRA_KEYSPACE", defaultKeyspace),

		KafkaBrokers:        Getenv("KAFKA_BROKERS", constants.KAFKA_HOST),
		KafkaGroupID:        Getenv("KAFKA_GROUP_ID", svc+"-group"),
		DropboxRefreshToken: Getenv("DROPBOX_REFRESH_TOKEN", constants.DROPBOX_REFRESH_TOKEN),
		ResendToken:         Getenv("RESEND_TOKEN", constants.RESEND_TOKEN),
		SmsGatewayUsername:  Getenv("SMS_GATEWAY_USERNAME", constants.SMS_GATEWAY_USERNAME),
		SmsGatewayPassword:  Getenv("SMS_GATEWAY_PASSWORD", constants.SMS_GATEWAY_PASSWORD),
		SmsGatewayUrl:       Getenv("SMS_GATEWAY_URL", constants.SMS_GATEWAY_URL),
	}
}
