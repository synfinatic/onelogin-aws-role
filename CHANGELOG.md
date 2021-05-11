# OneLogin AWS Role Changelog

## Unreleased

## v0.1.4 - 2021-05-11

- Add support for `--prompt-mfa` to force promting for which MFA you want to use
- Add support for `expire` command
- Various code clean up & refactoring
- Fix crash in `list` command when account aliases are not defined
- Add support for `ONELOGIN_AWS_DURATION` environment variable
- Improve `version` command
- Remove all traces of `duration` from config file as it was never supported
- Change log level argument to `-L`
- Add support for `--lines` argument 

## v0.1.3 - 2021-05-10

- Fix SecTrustedApplicationCreateFromPath deprecation warning

## v0.1.2 - 2021-04-27

- Change binary name to match git repo: onelogin-aws-role

## v0.1.1 - 2021-04-27

- add version command
- remove debugging output

## v0.1.0 - 2021-03-17

Initial release
