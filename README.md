# OneLogin Go AWS Assume Role

A GoLang implimentation of [OneLogin Python AWS Assume Role](
https://github.com/onelogin/onelogin-python-aws-assume-role).

Assume an AWS Role and get temporary credentials using Onelogin.

Users will be able to choose from among multiple AWS roles in multiple AWS 
accounts when they sign in using OneLogin in order to assume an AWS Role and 
obtain temporary AWS acccess credentials.

This is really useful for customers that run complex environments with multiple 
AWS accounts, roles and many different people that need periodic access as it 
saves manually generating and managing AWS credentials.

## Why?

Because the idea of running a Java application every I want AWS creds or dealing
with dependency/package management in Python made me shudder in horror.  I
wanted a _single_ binary and something that is cross-platform.

## AWS and OneLogin prerequisites

The "[Configuring SAML for Amazon Web Services (AWS) with Multiple Accounts and Roles](
https://onelogin.service-now.com/support?id=kb_article&sys_id=66a91d03db109700d5505eea4b9619a5)" guide explains how to:
 - Add the AWS Multi Account app to OneLogin
 - Configure OneLogin as an Identity Provider for each AWS account
 - Add or update AWS Roles to use OneLogin as the SAML provider
 - Add external roles to give OneLogin access to your AWS accounts
 - Complete your AWS Multi Account configuration in OneLogin

## Installation

### Hosting

#### Github

The project is hosted at github. You can download it from:
* Lastest release: https://github.com/synfinatic/onelogin-go-aws-assume-role/releases/latest
* Master repo: https://github.com/synfinatic/onelogin-go-aws-assume-role/tree/master

### Compiling

Just run `make`

### Settings

The onelogin-aws-assume-role utility uses the same settings file that
the [OneLogin Python SDK](
https://github.com/synfinatic/onelogin-python-sdk#getting-started) uses.

Is a json file named `onelogin.sdk.json` as follows:

```json
{
  "client_id": "",
  "client_secret": "",
  "region": "",
  "ip": ""
}
```

Where:

 * client_id  Onelogin OAuth2 client ID
 * client_secret  Onelogin OAuth2 client secret
 * region  Indicates where the instance is hosted. Possible values: 'us' or 'eu'.
 * ip  Indicates the IP to be used on the method to retrieve the SAMLResponse in
    order to bypass MFA if that IP was previously whitelisted.

For security reasons, IP only can be provided in `onelogin.sdk.json`.
On a shared machine where multiple users has access, That file should only be
readable by the root of the machine that also controls the
client_id / client_secret, and not by an end user, to prevent him manipulate the
IP value.

Place the file in the same path where the python script is invoked.


There is an optional file `onelogin.aws.json`, that can be used if you plan to
execute the script with some fixed values and avoid providing it on the command
line each time.

```json
{
  "app_id": "",
  "subdomain": "",
  "username": "",
  "profile": "",
  "duration": 3600,
  "aws_region": "",
  "aws_account_id": "",
  "aws_role_name": "",
  "profiles": {
    "profile-1": {
      "aws_account_id": "",
      "aws_role_name": "",
      "aws_region": ""
    },
    "profile-2": {
      "aws_account_id": ""
    }
  }
}
```

Where:

 * app_id Onelogin AWS integration app id
 * subdomain Needs to be set to the correct subdomain for your AWS integration
 * username The email address that is used to authenticate against Onelogin
 * profile The AWS profile to use in ~/.aws/credentials
 * duration Desired AWS Credential Duration in seconds. Default: 3600, Min: 900, Max: 43200
 * aws_region AWS region to use
 * aws_account_id AWS account id to be used
 * aws_role_name AWS role name to select
 * profiles Contains a list of profile->account id, and optionally role name
    mappings. If this attribute is populated `aws_account_id`, `aws_role_name`
    and `aws_region` will be set based on the `profile` provided when running
    the script.

**Note**: The values provided on the command line will take precedence over the
values defined on this file and, values defined at the _global_ scope in the
file, will take precedence over values defined at the `profiles` level. IN
addition, each attribute is treating individually, so be aware that this may
lead to somewhat strange behaviour when overriding a subset of parameters, when
others are defined at a _lower level_ and not overriden. For example, if you
had a `onelogin.aws.json` config file as follows:

```json
{
  ...
  "aws_region": "eu-east-1",
  "profiles": {
    "my-account": {
      "aws_account_id": "11111111",
      "aws_role_name": "Administrator"
    }
  }
}
````

And, you you subsequently ran the application with the command line arguments
`--profile my-account --aws-acccount-id 22222222` then the application would
ultimately attempt to log in with the role `Administrator` on account
`22222222`, with region set to `eu-east-1` and, if successful, save the
credentials to profile `my-account`.

In addition, there is another optional file that can be created to give more
human readable names to the account list, named `accounts.yaml`, which
should be placed in the same path where the python script is invoked:

```yaml
accounts:
  "987654321012": Prod account
  "123456789012": Dev Account
```

This isn't needed for the script to function but it provides a better user
experience.

### Docker installation method

* `git clone git@github.com:synfinatic/onelogin-go-aws-assume-role.git`
* `cd onelogin-go-aws-assume-role`
* Enter your credentials in the onelogin.sdk.json file as explained above
* Save the onelogin.sdk.json file in the root directory of the repo
* `docker build . -t awsaccess:latest`
* `docker run -it -v ~/.aws:/root/.aws awsaccess:latest onelogin-aws-assume-role
    --onelogin-username {user_email} --onelogin-subdomain {subdomain} 
    --onelogin-app-id {app_id} --aws-region {aws region} --profile default`

### How the process works

#### Step 1. Provide OneLogin data.

- Provide OneLogin's username/mail and password to authenticate the user
- Provide the OneLogin's App ID to identify the AWS app
- Provide the domain of your OneLogin's instance.

_Note: If you're bored typing your username (`--onelogin-username`),
App ID (`--onelogin-app-id`), subdomain (`--onelogin-subdomain`) or
AWS region (`--aws-region`)
every time, you can specify these parameters as command-line arguments or 
use the ENV vars (ONELOGIN_USERNAME, ONELOGIN_APP_ID, AWS_REGION) and
the tool won't ask for them any more._

_Note: Specifying your password directly with `--onelogin-password` is bad practice,
you should use that flag together with password managers, eg. with the OSX Keychain:
`--onelogin-password $(security find-generic-password -a $USER -s onelogin -w)`,
so your password won't be saved in you command line history.
Please note that your password **will** be visible in your process list,
if you use this flag (as the expanded command line arguments are part of the name of the process)._

With that data, a SAMLResponse is retrieved. And possible AWS Role are retrieved.

#### Step 2. Select AWS Role to be assumed.

- Provide the desired AWS Role to be assumed.
- Provide the AWS Region instance (required in order to execute the AWS API call).

#### Step 3. AWS Credentials retrieved.

A temporal AWS AccessKey and secretKey are retrieved in addition to a sessionToken.
Those data can be used to generate an AWS BasicSessionCredentials to be used in
any AWS API SDK.


## Quick Start

### Build the binary

```
make
```

### Installation

```
make install
```

### Usage

Assuming you have your AWS Multi Account app set up correctly and you’re using 
valid OneLogin API credentials stored on the `onelogin.sdk.json` placed at the 
root of the repository, using this tool is as simple as following the prompts.

```sh
> onelogin-aws-assume-role
```

Or alternately save them to your AWS credentials file to enable faster access 
from any terminal.

```sh
> onelogin-aws-assume-role --profile profilename
```

By default, the credentials only last for 1 hour, but you can 
[edit that restriction on AWS and set a max of 12h session duration](
https://aws.amazon.com/es/blogs/security/enable-federated-api-access-to-your-aws-resources-for-up-to-12-hours-using-iam-roles/).

Then set the `-z` or `duration` with the desired credentials session duration. 
The possible value to be used is limited by the AWS Role. Rea more at 
https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use.html#id_roles_use_view-role-max-session

You can also make it regenerate and update the credentials file by using the
`--loop` option to specify the number of iterations, and --time to specify the
minutes between iterations. If you specified a duration, be sure the value you
set for the duration session of the credential is bigger than the value you
set for the time, so your credentials will be renewed before expiration.

You can also make it interactive, with the `-x` or `--interactive`option, and
 at the end of the iteration, you will be asked if want to generate new
credentials for a new user or a new role.

The selection of the AWS account and Role can be also be done with the
`--aws-account-id` and `--aws-role-name` parameters. If both parameters are 
set then both will be matched against the list of available accounts and roles.
If only `--aws-account-id` is specified and you only have one available role
in that account, then that role will be chosen by default. If you have more
than one role in the given account then you will need to select the appropriate
one as per normal.

If you plan to execute the script several times over different Accounts/Roles
of the user and you want to cache the SAMLResponse, set the --cache-saml option

By default in order to select Account/Role, the list will be ordered by account
ids. Enable the --role_order option to list by role name instead.

For more info execute:

```sh
> onelogin-aws-assume-role --help
```

## Test your credentials with AWS CLI

AWS provide a CLI tool that makes remote access and management of resources 
super easy. If you don’t have it already then read more about it and install
it from here.

For convenience you can simply copy and paste the temporary AWS access
credentials generated above to set them as environment variables. This enables
you to instantly use AWS CLI commands as the environment variables will
take precedence over any credentials you may have in your *~/.aws* directory.

Assuming that:

 * you have the AWS CLI installed
 * you have set the OneLogin generated temporary AWS credentials as environment
    variables
* the role you selected has access to list EC2 instances

You should find success with the following AWS CLI command.

```
aws ec2 describe-instances
```

## Development

After checking out the repo, run `make test` to run tests.

To release a new version:
 1. Update the version number in `Makefile` and commit it.
 1. Create a new git tag 
 1. Run `make release` to build binaries
 1. Run `git push --all` to push commit & tags
 1. Create/edit the release on GitHub and include the binaries created above.

## Contributing

Bug reports and pull requests are welcome on GitHub at 
https://github.com/synfinatic/onelogin-go-aws-assume-role. 

## License

The gem is available as open source under the terms of the [GPLv3 License](https://opensource.org/licenses/gpl-3.0)

## Thanks

This README and project is heavily based on and includes content from
[OneLogin-Python-Go-AWS-Assume-Role](
https://github.com/onelogin/onelogin-python-aws-assume-role).