#!/bin/sh

set -e

bosh-compile --file "${INPUT_FILE}" --packages "${INPUT_PACKAGES}"

