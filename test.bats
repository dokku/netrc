#!/usr/bin/env bats

export SYSTEM_NAME="$(uname -s | tr '[:upper:]' '[:lower:]')"
export NETRC_BIN="build/$SYSTEM_NAME/netrc"

setup() {
  touch "$HOME/.netrc"
  mv "$HOME/.netrc" "$HOME/.netrc-bak"
  make prebuild $NETRC_BIN
}

teardown() {
  mv "$HOME/.netrc-bak" "$HOME/.netrc"
}

@test "version" {
  run $NETRC_BIN version
  [[ "$status" -eq 0 ]]

  run $NETRC_BIN -v
  [[ "$status" -eq 0 ]]

  run $NETRC_BIN --version
  [[ "$status" -eq 0 ]]
}
