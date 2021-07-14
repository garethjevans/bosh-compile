#!/bin/sh

set -e

bc compile --file "${INPUT_FILE}" --packages "${INPUT_PACKAGES}"

