#!/bin/bash

# Tag manually first
# git tag -a v0.1.0 -m "First release"
# git push origin v0.1.0

# For dry run:
# goreleaser --snapshot --skip-publish --rm-dist

goreleaser release --rm-dist
