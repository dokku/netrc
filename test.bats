#!/usr/bin/env bats

export SYSTEM_NAME="$(uname -s | tr '[:upper:]' '[:lower:]')"
export NETRC_BIN="build/$SYSTEM_NAME/netrc-amd64"

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

@test "(get) back-compat default output" {
  run cp "fixtures/valid/.netrc" "$HOME/.netrc"
  assert_success

  run $NETRC_BIN get heroku.com
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "username:longpassword"
}

@test "(get) single field" {
  run cp "fixtures/valid/.netrc" "$HOME/.netrc"
  assert_success

  run $NETRC_BIN get heroku.com --field password
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "longpassword"

  run $NETRC_BIN get heroku.com --field login
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "username"
}

@test "(get) multiple fields text format" {
  run cp "fixtures/valid/github-account.netrc" "$HOME/.netrc"
  assert_success

  run $NETRC_BIN get github.com --field login --field password --field account
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "login=username
password=password
account=account"
}

@test "(get) field ordering preserved in text" {
  run cp "fixtures/valid/.netrc" "$HOME/.netrc"
  assert_success

  run $NETRC_BIN get heroku.com --field password --field login
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "password=longpassword
login=username"
}

@test "(get) json format default fields" {
  run cp "fixtures/valid/.netrc" "$HOME/.netrc"
  assert_success

  run $NETRC_BIN get heroku.com --format json
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "\"login\": \"username\""
  assert_output_contains "\"password\": \"longpassword\""
  assert_output_contains "\"account\"" 0
}

@test "(get) json format with explicit fields" {
  run cp "fixtures/valid/github-account.netrc" "$HOME/.netrc"
  assert_success

  run $NETRC_BIN get github.com --field login --field account --format json
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "\"login\": \"username\""
  assert_output_contains "\"account\": \"account\""
  assert_output_contains "\"password\"" 0
}

@test "(get) json single field is still object" {
  run cp "fixtures/valid/.netrc" "$HOME/.netrc"
  assert_success

  run $NETRC_BIN get heroku.com --field password --format json
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "{"
  assert_output_contains "\"password\": \"longpassword\""
}

@test "(get) shell format default fields" {
  run cp "fixtures/valid/.netrc" "$HOME/.netrc"
  assert_success

  run $NETRC_BIN get heroku.com --format shell
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "login='username'
password='longpassword'"

  unset login password
  eval "$output"
  assert_equal "username" "$login"
  assert_equal "longpassword" "$password"
}

@test "(get) shell format escapes single quotes" {
  run $NETRC_BIN set evil.example user "pa'ss"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run $NETRC_BIN get evil.example --field password --format shell
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "password='pa'\\''ss'"

  unset password
  eval "$output"
  assert_equal "pa'ss" "$password"
}

@test "(get) missing field returns empty" {
  run cp "fixtures/valid/.netrc" "$HOME/.netrc"
  assert_success

  run $NETRC_BIN get heroku.com --field account --format json
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "\"account\": \"\""

  run $NETRC_BIN get heroku.com --field account --format shell
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "account=''"
}

@test "(get) invalid format rejected" {
  run cp "fixtures/valid/.netrc" "$HOME/.netrc"
  assert_success

  run $NETRC_BIN get heroku.com --format yaml
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid format 'yaml' specified"
}

@test "(get) invalid field rejected" {
  run cp "fixtures/valid/.netrc" "$HOME/.netrc"
  assert_success

  run $NETRC_BIN get heroku.com --field bogus
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid field 'bogus' specified"
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

@test "(get) custom path via --netrc-file" {
  custom="$(mktemp)"
  cp "fixtures/valid/.netrc" "$custom"

  run test -f "$HOME/.netrc"
  assert_failure

  run $NETRC_BIN get heroku.com --netrc-file "$custom"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "username:longpassword"

  run test -f "$HOME/.netrc"
  assert_failure

  rm -f "$custom"
}

@test "(get) custom path via NETRC env var" {
  custom="$(mktemp)"
  cp "fixtures/valid/.netrc" "$custom"

  run test -f "$HOME/.netrc"
  assert_failure

  NETRC="$custom" run $NETRC_BIN get heroku.com
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "username:longpassword"

  run test -f "$HOME/.netrc"
  assert_failure

  rm -f "$custom"
}

@test "(set) custom path via --netrc-file" {
  custom="$(mktemp)"
  rm -f "$custom"

  run $NETRC_BIN set github.com username password --netrc-file "$custom"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run cat "$custom"
  assert_success
  assert_output "$(cat fixtures/empty/github.netrc)"

  run test -f "$HOME/.netrc"
  assert_failure

  rm -f "$custom"
}

@test "(--netrc-file flag overrides NETRC env var)" {
  flag_path="$(mktemp)"
  env_path="$(mktemp)"
  cp "fixtures/valid/.netrc" "$flag_path"
  cp "fixtures/empty/.netrc" "$env_path"

  NETRC="$env_path" run $NETRC_BIN get heroku.com --netrc-file "$flag_path"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "username:longpassword"

  rm -f "$flag_path" "$env_path"
}

@test "(list) no netrc" {
  run test -f "$HOME/.netrc"
  assert_failure

  run $NETRC_BIN list
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run test -f "$HOME/.netrc"
  assert_success
}

@test "(list) empty netrc" {
  run cp "fixtures/empty/.netrc" "$HOME/.netrc"
  assert_success

  run $NETRC_BIN list
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists
}

@test "(list) valid netrc" {
  run cp "fixtures/valid/.netrc" "$HOME/.netrc"
  assert_success

  run $NETRC_BIN list
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "heroku.com"
}

@test "(list) --with-fields" {
  run cp "fixtures/valid/.netrc" "$HOME/.netrc"
  assert_success

  run $NETRC_BIN list --with-fields
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "heroku.com"
  assert_output_contains "login=username"
  assert_output_contains "password=longpassword"
}

@test "(list) --with-fields omits empty account" {
  run cp "fixtures/valid/.netrc" "$HOME/.netrc"
  assert_success

  run $NETRC_BIN list --with-fields
  echo "output: $output"
  echo "status: $status"
  assert_success
  [[ "$output" != *"account="* ]] || flunk "expected no account= field for machine without account"
}

@test "(list) --format=json default" {
  run cp "fixtures/valid/.netrc" "$HOME/.netrc"
  assert_success

  run $NETRC_BIN list --format json
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "\"heroku.com\""
}

@test "(list) --format=json --with-fields" {
  run cp "fixtures/valid/.netrc" "$HOME/.netrc"
  assert_success

  run $NETRC_BIN list --format json --with-fields
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "\"name\": \"heroku.com\""
  assert_output_contains "\"login\": \"username\""
  assert_output_contains "\"password\": \"longpassword\""
  assert_output_contains "\"account\": \"\""
}

@test "(list) --format=invalid" {
  run cp "fixtures/valid/.netrc" "$HOME/.netrc"
  assert_success

  run $NETRC_BIN list --format yaml
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid format 'yaml' specified"
}

@test "(list) default block excluded by default" {
  run cp "fixtures/with-default/.netrc" "$HOME/.netrc"
  assert_success

  run $NETRC_BIN list
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "heroku.com"
}

@test "(list) --include-default includes default block" {
  run cp "fixtures/with-default/.netrc" "$HOME/.netrc"
  assert_success

  run $NETRC_BIN list --include-default
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "default"
  assert_output_contains "heroku.com"
}

@test "(list) custom path via --netrc-file" {
  custom="$(mktemp)"
  cp "fixtures/valid/.netrc" "$custom"

  run test -f "$HOME/.netrc"
  assert_failure

  run $NETRC_BIN list --netrc-file "$custom"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "heroku.com"

  run test -f "$HOME/.netrc"
  assert_failure

  rm -f "$custom"
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
