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
    rename   Rename an entry in the .netrc file
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

### set

```text
netrc set <name> <login> <password> [account] [--netrc-file PATH]
netrc set <name> [--login VALUE] [--password VALUE] [--account VALUE] [--netrc-file PATH]
```

Creates or updates an entry. Passing `account` is optional. With no `--netrc-file` flag, `$NETRC` is consulted, then `~/.netrc`; the file is created with `0600` permissions if it does not exist.

```text
netrc set github.com username longpassword
```

Pass `--stdin` to read the password from standard input instead of as a positional argument. This keeps the password out of shell history and `ps` output:

```text
echo "$PW" | netrc set github.com username --stdin
```

With `--stdin`, the password positional is omitted - the form is `netrc set <name> <login> --stdin [account]`. A trailing newline on stdin is stripped; empty stdin is rejected.

To update a single field on an existing entry without re-supplying the others, pass `--login`, `--password`, or `--account`. When any of those flags is supplied, only the machine name is taken positionally and unspecified fields are preserved:

```text
netrc set github.com --password newpassword
netrc set github.com --login newuser
netrc set github.com --account ""           # clears the account field
```

`--password` and `--stdin` are mutually exclusive. Creating a brand-new entry via flag mode requires both `--login` and `--password` (the underlying file format does not round-trip empty values for these fields).

### rename

```text
netrc rename <old-name> <new-name> [--force] [--netrc-file PATH]
```

Renames a machine, copying all of its fields (`login`, `password`, `account`) to the new name. This avoids the `unset` + `set` workaround, which forces the caller to know and re-supply every field.

```text
netrc rename old.example.com new.example.com
```

The command exits with an error if the source machine does not exist, or if the destination machine already exists. Pass `--force` to overwrite an existing destination; a warning is printed to stderr when this happens so the destructive overwrite is visible.

```text
netrc rename old.example.com new.example.com --force
```

When `<old-name>` and `<new-name>` are equal the command is a no-op and the file is left untouched.

### default block

`.netrc` supports a `default` block that acts as a fallback for any machine not explicitly listed. All commands accept the literal name `default` as a target, and `list` exposes it via `--include-default`.

```text
netrc get default
netrc set default --login user --password pw
netrc rename default new-name
netrc rename old-name default
netrc unset default
netrc list --include-default
```

`default` is reserved by the file format and cannot be used as a regular `machine` name; the parser would interpret it as the default block.
