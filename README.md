# Submit File Server

## A file server to store the files uploaded to the new Submit system

### Usage:

```
submitfs

Usage:
  submitfs [command]

Available Commands:
  help        Help about any command
  start       start submitfs

Flags:
  -h, --help   help for submitfs

Use "submitfs [command] --help" for more information about a command.
```

```
start submitfs

Usage:
  submitfs start [flags]

Flags:
  -c, --config-file string         path to submit fs config file
      --file-server-path string    directory to store the files of the server, starting from home (default "var/cache/submit-file-server")
      --file-server-port int       port the file server should listen on (default 8081)
  -h, --help                       help for start
      --log-file string            log to file, specify the file location
      --log-file-and-stdout        write logs to stdout if log-file is specified?
      --log-file-max-age int       maximum age of the log file before it's rotated (default 3)
      --log-file-max-backups int   maximum number of log file rotations (default 3)
      --log-file-max-size int      maximum size of the log file before it's rotated (default 10)
      --log-level string           logging level [panic, fatal, error, warn, info, debug] (default "info")
      --password string            password (default "admin")
      --tls-cert-file string       path to a file containing a certificate to use for tls
      --tls-key-file string        path to a file containing a key to use for tls
      --user string                username (default "admin")

```