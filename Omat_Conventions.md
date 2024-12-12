# Omat Conventions

## Canonical Slugs

The omat deployment works with "canonical account slugs", which uniquely identify an AWS account in a system.

Information for an account can be found under the  `/omat/account_registry/<canonical name>` SSM parameter. The following keys will be available:

* `account_id`: The numeric AWS account id
* `environment`: The SDLC environment of the account
* `name`: A human readable name for the account
* `prefix`: The prefix that all SSM parameters pertaining to the account are stored under

## Roles

Each account has some set of roles, accessible under `<prefix>/roles/*`

All accounts will have a `<prefix>/roles/admin` SSM parameter, whose value is the ARN for an IAM role for admin access to the account. For now, this is the role to use.
