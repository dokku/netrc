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

@test "(version)" {
  run $NETRC_BIN version
  assert_success

  run $NETRC_BIN -v
  assert_success

  run $NETRC_BIN --version
  assert_success
}

@test "(get) no netrc" {
  run test -f "$HOME/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run $NETRC_BIN get invalid
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid machine 'invalid' specified"

  run test -f "$HOME/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(get) empty netrc" {
  run $NETRC_BIN get
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "This command requires 1 argument: <name>"

  run cp "fixtures/empty/.netrc" "$HOME/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run $NETRC_BIN get invalid
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid machine 'invalid' specified"
}

@test "(get) valid netrc" {
  run cp "fixtures/valid/.netrc" "$HOME/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run $NETRC_BIN get
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "This command requires 1 argument: <name>"

  run $NETRC_BIN get heroku.com
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "username:longpassword"

  run $NETRC_BIN get invalid
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid machine 'invalid' specified"
}

@test "(set) no netrc" {
  run test -f "$HOME/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run $NETRC_BIN set github.com username password
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run cat "$HOME/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$(cat fixtures/empty/github.netrc)"

  run test -f "$HOME/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(set) empty netrc" {
  run cp "fixtures/empty/.netrc" "$HOME/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run $NETRC_BIN set github.com
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "This command requires 3 and at most 4 arguments, 1 argument given:"

  run $NETRC_BIN set github.com username
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "This command requires 3 and at most 4 arguments, 2 arguments given:"

  run $NETRC_BIN set github.com username password
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run cat "$HOME/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$(cat fixtures/empty/github.netrc)"

  run $NETRC_BIN set github.com username password account
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run cat "$HOME/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$(cat fixtures/empty/github-account.netrc)"
}

@test "(set) valid netrc" {
  run cp "fixtures/valid/.netrc" "$HOME/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run $NETRC_BIN set github.com
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "This command requires 3 and at most 4 arguments, 1 argument given:"

  run $NETRC_BIN set github.com username
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "This command requires 3 and at most 4 arguments, 2 arguments given:"

  run $NETRC_BIN set github.com username password
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run cat "$HOME/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$(cat fixtures/valid/github.netrc)"

  run $NETRC_BIN set github.com username password account
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run cat "$HOME/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$(cat fixtures/valid/github-account.netrc)"
}

@test "(unset) no netrc" {
  run test -f "$HOME/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run $NETRC_BIN unset github.com
  echo "output: $output"
  echo "status: $status"
  assert_success

  run cat "$HOME/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$(cat fixtures/empty/.netrc)"

  run test -f "$HOME/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(unset) empty netrc" {
  run cp "fixtures/empty/.netrc" "$HOME/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run $NETRC_BIN unset github.com
  echo "output: $output"
  echo "status: $status"
  assert_success

  run $NETRC_BIN unset github.com username
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "This command requires 1 argument, 2 arguments given: <name>"

  run cat "$HOME/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$(cat fixtures/empty/.netrc)"
}

@test "(unset) valid netrc" {
  run cp "fixtures/valid/.netrc" "$HOME/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run $NETRC_BIN unset github.com username
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "This command requires 1 argument, 2 arguments given: <name>"

  run $NETRC_BIN unset github.com
  echo "output: $output"
  echo "status: $status"
  assert_success

  run cat "$HOME/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$(cat fixtures/valid/.netrc)"

  run $NETRC_BIN unset github.com
  echo "output: $output"
  echo "status: $status"
  assert_success

  run $NETRC_BIN unset heroku.com username
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "This command requires 1 argument, 2 arguments given: <name>"

  run $NETRC_BIN unset heroku.com
  echo "output: $output"
  echo "status: $status"
  assert_success

  run cat "$HOME/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$(cat fixtures/empty/.netrc)"

  run $NETRC_BIN unset heroku.com
  echo "output: $output"
  echo "status: $status"
  assert_success
}

# test functions
flunk() {
  {
    if [[ "$#" -eq 0 ]]; then
      cat -
    else
      echo "$*"
    fi
  }
  return 1
}

assert_equal() {
  if [[ "$1" != "$2" ]]; then
    {
      echo "expected: $1"
      echo "actual:   $2"
    } | flunk
  fi
}

# ShellCheck doesn't know about $status from Bats
# shellcheck disable=SC2154
# shellcheck disable=SC2120
assert_failure() {
  if [[ "$status" -eq 0 ]]; then
    flunk "expected failed exit status"
  elif [[ "$#" -gt 0 ]]; then
    assert_output "$1"
  fi
}

# ShellCheck doesn't know about $output from Bats
# shellcheck disable=SC2154
assert_output() {
  local expected
  if [[ $# -eq 0 ]]; then
    expected="$(cat -)"
  else
    expected="$1"
  fi
  assert_equal "$expected" "$output"
}

# ShellCheck doesn't know about $output from Bats
# shellcheck disable=SC2154
assert_output_contains() {
  local input="$output"
  local expected="$1"
  local count="${2:-1}"
  local found=0
  until [ "${input/$expected/}" = "$input" ]; do
    input="${input/$expected/}"
    found=$((found + 1))
  done
  assert_equal "$count" "$found"
}

# ShellCheck doesn't know about $output from Bats
# shellcheck disable=SC2154
assert_output_exists() {
  [[ -n "$output" ]] || flunk "expected output, found none"
}

# ShellCheck doesn't know about $output from Bats
# shellcheck disable=SC2154
assert_output_not_exists() {
  [[ -z "$output" ]] || flunk "expected no output, found some"
}

# ShellCheck doesn't know about $status from Bats
# shellcheck disable=SC2154
# shellcheck disable=SC2120
assert_success() {
  if [[ "$status" -ne 0 ]]; then
    flunk "command failed with exit status $status"
  elif [[ "$#" -gt 0 ]]; then
    assert_output "$1"
  fi
}
