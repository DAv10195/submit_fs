package fileserver

const (
	DefPort 					= ":5432"
	DefAdminUser                   = "admin"
	DefAdminPass                   = "admin123"
)


// submit server configuration
type Config struct {
	Port						int
}
