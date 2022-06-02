# pingdom2jira

`pingdom2jira` is a service draft that provides a web hook to be called from pingdom.
On a web hook trigger the service creates a jira ticket (details can be configured via
URL query parameters).

Since the web hook triggered by pingdom does not provide any details on the use root
cause (which tends to be useful for debugging) `pingdom2jira` performs two more steps:

* First, based on the ID of the check in the web hook body, the pingdom API is queried to fetch details of the actual check that failed.
* With this details, the check executed by pingdom is reconstructed and _reproduced_ by `pingdom2jira` in order to get information on the root cause.

This is not very sophisticated, however that seems to be the only way to get the desired
information.

## Configuration
`pingdom2jira` takes a few arguments:

```
go run *.go -h
Usage of /tmp/go-build1466037281/b001/exe/config:
  -a string
    	ip:port to listen on (runs as lambda if empty) (default ":8080")
  -c string
    	name of the config file (alternatively the config file be specified via environment variable 'CONFIG_FILE')) (default "./config.yaml")
  -m string
    	mode, can me either 'local', 'azurefunc' or 'awslambda' (default "local")
  -v	print version information
```

Note that the configuration file (`-c` or environment variable `CONFIG_FILE`) can be provided using different locations:

* Use a simple file path to indicate that a _local file_ is used.
* Use a path of the pattern `s3://[bucket_name]/[object_path]` to indicate that file is stored on s3 is used. All [official authentication methods](https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/) are allowed.
* Use  a path of the pattern `blob://[storage_account]/[container_name]/[object_path]` to indicate that file is stored on Azure Blob storage. The `AZURE_STORAGE_ACCOUNT_KEY` environment variable must be set in order to grant access.

The configuration file holds should look something like this:

```
---
path_prefix: api
pingdom:
  token: 19_aoestuhpcuRCEgu3...
jira:
  url: https://example.atlassian.net/
  username: john.doe@example.com
  password: 434oua...

```

## Usage

On pingdom configure an integration (aka web hook) with an URL of the following pattern:

```
http://[BASE_URL]/hooks/?jira_assignee_email=[ASSIGNEE]&jira_type=[JIRA_TYPE]&jira_project=[PROJECT_SHORT_NAME]'
```
