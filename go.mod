module github.com/DAv10195/submit_fs

go 1.13

require (
	github.com/DAv10195/submit_commons v0.0.0-20210414053531-8066a0155d69
	github.com/gorilla/mux v1.8.0
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/DAv10195/submit_commons => ..\submit_commons
