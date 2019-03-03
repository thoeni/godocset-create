## GoDocset Create

GoDocset create allows you to generate GoDoc Docset in the format required by
offline doc tools such as Dash.

Docs can be pulled down from Github both from public and private repositories
within an organisation.

In order to do the latter, an authorised access token is needed.

### TL;DR

#### Prerequisite:
Get a GitHub token from https://github.com/settings/tokens and **enable SSO** if
the organisation so requires.
---

Prepare the config file as described below, and run:
```
docker run -v `pwd`:/tmp thoeni/godocset-create
``` 

### Build:

Wanna build it yourself? Go ahead...

To build the Docker image, run this and replace the value with your authorised
token:
```
docker build -t godocset-create .
```

### Configure

#### Prerequisite:
Get a GitHub token from https://github.com/settings/tokens and **enable SSO** if
the organisation so requires.
---

Configuration is provided as `toml` file, and looks like this:
```toml
[Github]
# your github user id
user_id = "thoeni"
token = "1224abcde1234abcde1234abcde"
clone_target_dir = "/go"

[Docset]
# name is the name the docset will be generated with
name = "anameyoulike"
users = [
	"thoeni", # Github usernames for users you want to pull packages from
	"pkg"
]
organizations = [
	"myorg" # Github usernames for organisations you want to pull packages from.
]
# Filters for packages you want to export. Wildcard * means everything after that.
filters = [
	"github.com/thoeni/*"
	,"github.com/pkg/errors"
]
# output is where do you want to generate the output. Ideally this is where you will mount your volume to.
output = "/tmp"

[Options]
silent = true # Enable/Disable verbose output when parsing the documentation
```

### Run:

This will mount the current directory into the container /tmp, where the program
will look for the config and will output the result. 

```
 docker run -v `pwd`:/tmp godocset-create
```
