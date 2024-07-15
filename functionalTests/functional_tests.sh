#!/bin/bash
ENV_OS="$1"
PKG_NAME="$2"

echo -e "Performing functional tests on ${ENV_OS} for ${PKG_NAME}\n\n"

if [ $# -eq 0 ]; then
    echo "Usage: $0 <corteca-package>"
    exit 1
fi

source /ft/functional_tests_functions.sh

if [[ ${ENV_OS} == debian:bookworm-slim || ${ENV_OS} == debian:jessie-slim ]]; then
    debian_install "$PKG_NAME"
elif [[ ${ENV_OS} == centos:8 ]]; then
    centos_install "$PKG_NAME"
else
    echo "Error: Unable to determine the package management system."
    exit 1
fi

test_global_config_get

test_user_config_get

test_create_test_template_with_cmd

if [[ ${ENV_OS} == debian:bookworm-slim || ${ENV_OS} == debian:jessie-slim ]]; then
    debian_uninstall
elif [[ ${ENV_OS} == centos:8 ]]; then
    centos_uninstall
else
    echo "Error: Unable to determine the package management system."
    exit 1
fi
