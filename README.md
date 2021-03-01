# OneLogin AWS (Assume) Role

Assume an AWS Role and get temporary credentials using [Onelogin](
https://www.onelogin.com).

Users will be able to choose from among multiple AWS roles in multiple AWS 
accounts when they sign in using OneLogin in order to assume an AWS Role and 
obtain temporary AWS acccess credentials.

This is really useful for SRE/Engineering organizations that run complex 
environments with multiple AWS accounts, roles and many different people that 
need periodic access as it saves manually generating and managing AWS credentials.

## ***WARNING***

This project doesn't work yet and the docs are bad.  Don't even bother telling me
you are having problems because the answer is "duh".

## Why?

Because the idea of running a [Java application](
https://github.com/onelogin/onelogin-aws-cli-assume-role)
every I want AWS creds or dealing
with dependency/package management in [Python](
https://github.com/onelogin/onelogin-python-aws-assume-role)
made me shudder in horror.  I wanted a _single_ binary and something that is
cross-platform.  Also, I wanted to provide a different user experience- no 
mixing of JSON and YAML for config files for example.


## AWS and OneLogin prerequisites

Read what [Onelogin has to say here](
https://github.com/onelogin/onelogin-aws-cli-assume-role#aws-and-onelogin-prerequisites).

## Installation

Just run `make` to build a binary and copy it to a location in your `$PATH`.

## Settings

OneLogin AWS Role has a single YAML configuration file: 
`~/.onelogin.yaml`

It contains the following sections:

### OneLogin Config

This section defines the OAuth2 configuration necessary to talk to OneLogin
API servers.  You will need to get this from your administrator.

```yaml
client_id: <OneLogin Client ID>
client_secret: <OneLogin Client Secret>
region: <OneLogin Region>
username: <OneLogin Username>
subdomain: <Onelogin Subdomain>
ip: <IP Address>
mfa: <device_id|device_type>
```

Where:

 * `client_id`  - OneLogin OAuth2 client ID (required)
 * `client_secret`  - OneLogin OAuth2 client secret (required)
 * `region`  - One of `us` or `eu` depending on OneLogin server region to use (required)
 * `username` - Your OneLogin username/email address (required)
 * `subdomain` - Your organization's OneLoging subdomain (required)
 * `ip`  - Specify the IP to be used on the method to retrieve the SAMLResponse in
    order to bypass MFA if that IP was previously whitelisted. (optional)
 * `mfa` - Default device_id or device_type for MFA to skip prompting

By default, the credentials only last for 1 hour, but you can 
[edit that restriction on AWS and set a max of 12h session duration](
https://aws.amazon.com/es/blogs/security/enable-federated-api-access-to-your-aws-resources-for-up-to-12-hours-using-iam-roles/).

### AWS Account Config

This section defines each of the AWS Account & Roles that may be used via
AssumeRole.

```yaml
apps:
    <app_id>:
        name: <Application Name>
        alias: <Application Alias>
        roles:
			- arn: <Role ARN>
              alias: <Role Alias>
              region: <Default AWS Region>
              duration: <Number of Seconds>
              profile: <Profile Name>
```

Where:

 * `app_id`  - Is the Application ID provided by your administrator (required)
 * `name` - Name of the OneLogin Application (optional)
 * `alias` - Alias for OneLogin Application (optional)
 * `arn`   - AWS ARN to assume (required)
 * `profile`  - Name of AWS_PROFILE to write to `~/.aws/config` (optional)
 * `region`  - Configure the default AWS region.  Default: `us-east-1` (optional)
 * `duration`  - How many seconds your AssumeRole credentials should last by default
    (optional, default is 3600, minimum 900 and max 43200)

Not that you can configure multiple roles for each account, multiple accounts for
each applications and multiple applications.  While it is strongly advised that 
you use unique profiles for each account+region combination, that is not required.

###  AWS Account Aliases

There is an optional section that can be created to give more
human readable names to the account list.

```yaml
aws_accounts:
    <account id>: <account alias/name>
```

Note if you do not have an alias set for a given AWS Account Id, it will try 
calling `iam:ListAccountAliases` to determine the alias of the account.


## License

This program is available under the terms of the [GPLv3 License](https://opensource.org/licenses/gpl-3.0)

## Thanks

This README and project is heavily based on and includes content from
[OneLogin-Python-Go-AWS-Assume-Role](
https://github.com/onelogin/onelogin-python-aws-assume-role).



FIX BELOW THIS LINE
===================


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
> onelogin-aws-role
```

Or alternately save them to your AWS credentials file to enable faster access 
from any terminal.

```sh
> onelogin-aws-role --profile profilename
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
