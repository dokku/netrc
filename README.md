# netrc

Utility that allows users to manage netrc files.

## Downloading

See Github Releases for prebuilt Linux and MacOS binaries, as well as Debian packages.

OS packages are also available via [packagecloud](https://packagecloud.io/dokku/dokku).

## Usage

> Warning: Macro setting is currently unhandled

```text
Usage: netrc [--version] [--help] <command> [<args>]

Available commands are:
    get      Get an entry from the .netrc file
    list     List machines in the .netrc file
    set      Set an entry in the .netrc file
    unset    Unset an entry from the .netrc file
    version  Return the version of the binary
```

### get

```text
netrc get <name> [--field login|password|account ...] [--format=text|json|shell] [--netrc-file PATH]
```

By default `get` prints `login:password` for back-compatibility:

```console
$ netrc get github.com
username:longpassword
```

Pass `--field` (repeatable) to select specific fields. With a single field, only the value is printed:

```console
$ netrc get github.com --field password
longpassword
```

Multiple `--field` flags emit one `key=value` per line in the order supplied:

```console
$ netrc get github.com --field login --field password
login=username
password=longpassword
```

Use `--format json` for machine-readable output:

```console
$ netrc get github.com --field login --field password --format json
{
  "login": "username",
  "password": "longpassword"
}
```

Use `--format shell` to emit `eval`-safe assignments (single-quote escaped):

```console
$ eval "$(netrc get github.com --format shell)"
$ echo "$login $password"
username longpassword
```
