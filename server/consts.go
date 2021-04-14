package fileserver

import "time"

const (
	serverTimeout                   = 15 * time.Second
 	unauthorized = "failed to login to file server"
 	wrongCreds = "admin creds are incorrect. try again"
)


