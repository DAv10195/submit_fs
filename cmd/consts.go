package cmd

const (
	submitFileServer 		= "submitfs"
	fileServer              = "file-server"
	start					= "start"

	defaultConfigFileName	= "submit_file_server.yml"
	yaml					= "yaml"
	info					= "info"
	defMaxLogFileSize		= 10
	defMaxLogFileAge		= 3
	defMaxLogFileBackups	= 3
	deLogFileAndStdOut		= false

	flagConfigFile        = "config-file"
	flagFsDir             = "file-server-dir"
	flagFileServerPort    = "file-server-port"
	flagLogLevel          = "log-level"
	flagLogFile           = "log-file"
	flagLogFileAndStdout  = "log-file-and-stdout"
	flagLogFileMaxSize    = "log-file-max-size"
	flagLogFileMaxBackups = "log-file-max-backups"
	flagLogFileMaxAge     = "log-file-max-age"
	//flagFileServerFolder  = "submit-server-folder"
	flagAdminConfFile     = "admin-conf-file"
	flagFileServerPath    = "file-sever-path"

)
