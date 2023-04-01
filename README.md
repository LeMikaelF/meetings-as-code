# Meetings as Code

This is a small Terraform provider that can be used to create, update, and delete Outlook calendar
events. I wrote it as an April's Fool joke, so you shouldn't use it for anything serious.

Usage:

```shell
# Register a Microsoft application and export the app's clientId as an env variable.
# See https://learn.microsoft.com/en-us/azure/active-directory/develop/quickstart-register-app
export MICROSOFT_APP_CLIENT_ID="something..."

make
export GRAPH_API_TOKEN=$(make auth) # follow the instructions in the console to authenticate

cd terraform
terraform init
terraform apply
```
