# Dev Notes

 1. How much should I copy the official OneLogin app?
    - Easier to migrate
    - Config files in JSON though???  Oh _and_ YAML.  Because.
 1. Integrate with KeyChain like [aws-vault](https://github.com/99designs/aws-vault)
    - Will need to [code sign](https://github.com/99designs/aws-vault#development)
 1. Shouldn't just spit out ENV vars and expect you to copy & paste.
    - Follow `aws-vault` example of executing programs (including a shell)
 1. Run an EC2-Metadata service like `aws-vault`?
 1. Should use [External Sourcing](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sourcing-external.html)
	- Need to write the necessary info as a JSON blob which allows you to easily
	define OneLogin as the means of accessing an AWS_PROFILE without having to 
	edit the ~/.aws/config file!   Note that we would need to impliment some level
	of caching for this to work, but that seems reasonable :)
 1. Another golang program [allcloud-io/clisso](https://github.com/allcloud-io/clisso)
 1. Another secret mgmt library for OSX/Linux: [tmc/keyring](https://github.com/tmc/keyring)

# Security

 1. The OAuth2 AccessToken is good for 10hrs and should be cached to avoid rate limiting
    This is perfectly safe as long as the creds aren't exposed and someone uses them
    to DoS us due to the 5000/req/hr/account.  (account, not user?)
 1. User Login Session tokens may require MFA and expire after 2 minutes
 1. The SAML assertion is only good for a service defined number of minutes?
    AWS SAML is for a few minutes.

# Challenges

 1. The onelogin-go-sdk is neutered and doesn't support MFA :-/
 1. Need to see how `--loop` feature is supported?  Login Session Tokens can't be used for long periods of time?
    Pretty sure this doesn't automate authentication!  Looks like it merely 
    automates running the tool again which is _very_ different (requires you to re-auth)

