package migrations

var DDL = []string{
	`CREATE TABLE IF NOT EXISTS email_service.emails (
	id text PRIMARY KEY,
	receiver list<text>,
	subject text,
	status text,
	retries int
	);`,
}
