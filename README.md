# OneLogin AWS (Assume) Role

Assume an AWS Role and get temporary credentials using [OneLogin](
https://www.onelogin.com).

Users will be able to choose from among multiple AWS roles in multiple AWS
accounts when they sign in using OneLogin in order to assume an AWS Role and
obtain temporary AWS acccess credentials.

This is really useful for SRE/Engineering organizations that run complex
environments with multiple AWS accounts, roles and many different people that
need periodic access as it saves manually generating and managing AWS credentials.

## Why?

Because the idea of running a [Java application](
https://github.com/onelogin/onelogin-aws-cli-assume-role)
every I want AWS creds or dealing
with dependency/package management in [Python](
https://github.com/onelogin/onelogin-python-aws-assume-role)
made me shudder in horror.  I wanted a _single_ binary and something that is
cross-platform.  Also, I wanted to provide a different user experience- no
mixing of JSON and YAML for config files for example.

Lastly, I wanted something that was a bit more secure than these options and
handled all the secrets more securely.


## AWS and OneLogin prerequisites

Read what [OneLogin has to say here](
https://github.com/onelogin/onelogin-aws-cli-assume-role#aws-and-onelogin-prerequisites).

## Installation

Just run `make` (GNU Make!) to build a binary and copy it to a location in your `$PATH`.

## Settings

OneLogin AWS Role has a single YAML configuration file:
`~/.onelogin-aws-role.yaml`

It contains the following sections:

### OneLogin Config

This section defines the OAuth2 configuration necessary to talk to OneLogin
API servers.  You will need to get this from your administrator.

```yaml
region: <OneLogin Region>
username: <OneLogin Username>
subdomain: <OneLogin Subdomain>
ip: <IP Address>
mfa: <device_id>
```

Where:

 * `region`  - One of `us` or `eu` depending on OneLogin server region to use (required)
 * `username` - Your OneLogin username/email address (required)
 * `subdomain` - Your organization's OneLoging subdomain (required)
 * `ip`  - Specify the IP to be used on the method to retrieve the SAMLResponse in
    order to bypass MFA if that IP was previously whitelisted. (optional)
 * `mfa` - Default device_id for MFA to skip prompting (optional)

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
              profile: <AWS Profile Name>
              region: <Default AWS Region>
              duration: <Number of Seconds>
              profile: <Profile Name>
```

Where:

 * `app_id`  - Is the Application ID provided by your administrator (required)
 * `name` - Name of the OneLogin Application (optional)
 * `alias` - Alias for OneLogin Application (optional)
 * `arn`   - AWS ARN to assume (required)
 * `profile`  - Friendly name of this role and section of AWS_PROFILE to write to `~/.aws/credentials` (required)
 * `region`  - Configure the default AWS region.  Default: `us-east-1` (optional)

Note that you can configure multiple roles for each account, multiple accounts for
each applications and multiple applications.

###  AWS Account Aliases

There is an optional section that can be created to give more
human readable names to the account list.

```yaml
aws_accounts:
    <account id>: <account alias/name>
```

## Usage

### Check your config

After you have edited your `~/.onelogin-aws-role.yaml` config file, you can verify it by
running `onelogin-aws-role` and you should see a list of AWS Accounts and Roles that
you have configured.

### Configure your OneLogin ClientId and Client Secret

All API calls to OneLogin [require a valid Oauth access token](
https://developers.onelogin.com/api-docs/2/oauth20-tokens/generate-tokens-2).  In
order to get one of these tokens, you must first authenticate to the OneLogin service
using the ClientId and Client Secret provided to you by your administrator.  Both
ClientId and Client Secrets are 64 character hex strings.

##### Set ClientId and Client Secret

`onelogin-aws-role oauth set`

##### Show ClientId and Client Secret

`onelogin-aws-role oauth show`

<!--
### Get STS Session Token for an IAM Role

`onelogin-aws-role role <profile name>`

This will ask you to authenticate to OneLogin and then retrieve the STS Session Token
for the specified IAM role and cache that in your Keychain.  If you have an existing
cached STS Session Token for this role, it will renew it.
-->

### Execute command with an IAM Role

`onelogin-aws-role exec <profile name> [command] [args...]`

This will use the cached STS Session Token or ask you to authenticate to OneLogin
and then execute the provided command & arguments using those credentials.  If no
command is specified, it will start a new interactive shell.

Note that all the necessary shell environment variables will be set:

 * `AWS_ACCESS_KEY_ID` -- AWS Credential for authentication
 * `AWS_SECRET_ACCESS_KEY` -- AWS Credential for authentication
 * `AWS_SESSION_TOKEN` -- AWS Credential for authentication
 * `AWS_DEFAULT_REGION` -- AWS region
 * `AWS_SESSION_EXPIRATION` -- Date & Time this session token will expire
 * `AWS_ROLE_ARN` -- Selected AWS Role ARN
 * `AWS_ENABLED_PROFILE` -- note that this is different from `AWS_PROFILE` as we do not
	want to [confuse clients that may try to load](
        https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-profiles.html)
        `~/.aws/config` and `~/.aws/credentials`.

<!--
### Cache All STS Session Tokens for a OneLogin Application

`onelogin-aws-role app <appid>`

This will authenticate you to OneLogin and retrieve and cache all of the STS
Session Tokens for all the IAM roles associated with this OneLogin Application.  Further
calls to `onelogin-aws-role exec <profile> ...` which are contained in that OneLogin 
Application will not require re-authentication until the STS Session Tokens expire.
-->

## Other Files

onelogin-aws-role will create the following file(s):

 * `~/.onelogin-aws-role.cache`
	Contains SAML Assertions (good for ~3min) and the OneLogin bearer token (good for ~10hrs)

## Environment Variables

The following environment variables are honored to specify defaults:

 * `ONELOGIN_AWS_DURATION` -- Default number of minutes to request the STS Session to be good for
 * `AWS_DEFAULT_REGION` -- Default AWS Region to make API calls to

## License

This program is available under the terms of the [GPLv3 License](https://opensource.org/licenses/gpl-3.0)

## Thanks

This README and project is heavily based on and includes content from
[OneLogin-Python-AWS-Assume-Role](
https://github.com/onelogin/onelogin-python-aws-assume-role).
