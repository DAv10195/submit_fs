package fileserver

const (
	DefPort 					= ":5432"
	adminUser                   = "admin"
	adminPass                   = "admin"
)

// submit server configuration
type Config struct {
	Port						int
}
