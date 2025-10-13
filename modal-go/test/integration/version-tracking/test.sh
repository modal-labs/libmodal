#!/bin/sh
set -e

cd "$(dirname "$0")"
go mod tidy
go build -o test-app .

output=$(./test-app)

expected="modal-go/v0.0.99"

if [ "$output" = "$expected" ]; then
	rm -f test-app
	exit 0
else
	echo "Version tracking failed:"
	echo "  Expected: $expected"
	echo "  Got: $output"
	rm -f test-app
	exit 1
fi
