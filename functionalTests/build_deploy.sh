#!/bin/bash
CORTECAVM_ARMV7_NAME="$1"
CORTECAVM_ARMV8_NAME="$2"

echo -e "Performing functional test for corteca-linux-amd64 build and exec\n\n"

source ./functionalTests/functional_tests_functions.sh

start_cortecavm "${CORTECAVM_ARMV7_NAME}" armv7 8027
start_cortecavm "${CORTECAVM_ARMV8_NAME}" armv8 8028
test_build
# test_exec "${CORTECAVM_ARMV7_NAME}" "${CORTECAVM_ARMV8_NAME}" "${sequence_name}" "${to_be_published}" "${use_artifact_flag}"
test_exec "${CORTECAVM_ARMV7_NAME}" "${CORTECAVM_ARMV8_NAME}" "testSequence" 1 1

test_exec "${CORTECAVM_ARMV7_NAME}" "${CORTECAVM_ARMV8_NAME}" "deploy" 0 0
