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
    set      Set an entry in the .netrc file
    unset    Unset an entry from the .netrc file
    version  Return the version of the binary
```

## Releases

Anybody can propose a release. First bump the version in `Makefile` and make sure tests are passing. Then open a Pull Request from `master` into the `release` branch. Once a maintainer approves and merges, Github Actions will build a release and upload it to Github.
