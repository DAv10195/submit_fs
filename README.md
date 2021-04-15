# submit-file-server


This service a standalone file system, designed to be used with the submit server.



```
Usage:
  submitfs [command]

Available Commands:
  help        Help about any command
  start       start submitfs

Flags:
  -h, --help   help for submitfs

Use "submitfs [command] --help" for more information about a command.

```

Downloading and uploading files:

After deploying the service, use HTTP GET to download files from the server.
Use the url to specify the path, for example:
/2020/cem1/ai-89202/nikita-kogan/assign1.zip

In order to upload files use HTTP POST and specify the path in the URL just like in the download.
The file itself should be put in the request body.
for example
/2020/cem1/ai-89202/nikita-kogan/assign1.zip
