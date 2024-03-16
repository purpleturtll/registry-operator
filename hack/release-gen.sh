#!/usr/bin/env bash

set -ex

AUTHOR="${1}"

# Get the PR number
pr_number=$(gh pr list --author="${AUTHOR}" --state=open --json=number --jq='.[].number' --limit=1)

# Check if the PR number is not empty
if [ -n "${pr_number}" ]; then
    # Checkout to the PR
    gh pr checkout "${pr_number}"

    make generate manifests bundle

    if ! git diff --quiet ; then
      git add .
      git commit -sm "chore(release): generated manifests"
      git push
    else
      echo "Nothing to commit!"
    fi
else
    echo "No open PR found for ${AUTHOR}"
fi
