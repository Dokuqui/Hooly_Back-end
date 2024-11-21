#!/bin/bash

# Define the path to the commit message file
COMMIT_FILE=".git/COMMIT_EDITMSG"

# Check if the file exists
if [[ -f "$COMMIT_FILE" ]]; then
  # Read the commit message
  COMMIT_MSG=$(cat "$COMMIT_FILE")
else
  echo "Error: Commit message file not found at $COMMIT_FILE"
  exit 1
fi

# Define your commit message validation pattern
PATTERN="^(feat|fix|chore|docs|style|refactor|perf|test|ci)\([a-z]+\): .{1,50}"

# Validate the commit message
if [[ ! "$COMMIT_MSG" =~ $PATTERN ]]; then
  echo "ERROR: Commit message does not follow the required format."
  echo "Example: feat(auth): add login feature"
  exit 1
fi

# If valid, print success message
echo "Commit message is valid."
exit 0
