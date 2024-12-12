# cli-o-mat

A useful CLI tool for interacting with Omat-based infrastructure.

## Workstation Setup

This assumes you have HomeBrew set up and functioning properly.  Additionally, you will need a
functioning git install.  It is recommended that you use the one installable from HomeBrew, or
install your own.  The one provided by Apple is not recommended.

```bash
brew bundle --no-upgrade

eval "$(gimme $(<.go-version))"

make setup
```

## Working on the Code

```bash
eval "$(gimme $(<.go-version))" # Activate the pertinent Go version.

make # See what make targets are available

make compile # Compile things.

make fmt # Format code
make lint # Lint code
make fix # Fix some anti-patterns in code

make build # Run tests, generate code, compile code

make dist # Like `make build`, but build a distributable (fat) binary
```

## Usage

```bash
omat manual # Show the maual, including what configuration needs to be set up.

omat # List available commands.

omat launch <account> <template> <ssh key> # Launch a new instance from a template.

omat hosts <account> # See all running hosts.

omat help <subcommand> # See help for a specific sub-command.
```
