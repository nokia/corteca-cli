#!/bin/bash
PKG_NAME="$1"

if [ $# -eq 0 ]; then
    echo "Usage: $0 <corteca-package>"
    exit 1
fi

source /ft/functional_tests_functions.sh

useradd -m -s /bin/bash tester
su - tester 

if command -v dpkg &> /dev/null; then
    debian_install "$PKG_NAME"
elif command -v rpm &> /dev/null; then
    centos_install "$PKG_NAME"
else
    echo "Error: Unable to determine the package management system."
    exit 1
fi

test_global_config_show

test_user_config_show

test_create_test_template_with_cmd

if command -v dpkg &> /dev/null; then
    debian_uninstall
elif command -v rpm &> /dev/null; then
    centos_uninstall
else
    echo "Error: Unable to determine the package management system."
    exit 1
fi
