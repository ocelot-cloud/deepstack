#!/bin/bash

set -e

LAST_TAG=$(git describe --tags `git rev-list --tags --max-count=1`)
echo "The last tag was: $LAST_TAG"
read -p "Enter the new tag name: " NEW_TAG

git commit -am "refactoring" || true
git push
git tag $NEW_TAG
git push origin $NEW_TAG

echo "Tag $NEW_TAG created and pushed to origin."
