#!/bin/bash
set -x
CORTECAVM_ARMV7_NAME="$1"
CORTECAVM_ARMV8_NAME="$2"
ft_location=$(readlink -e $(dirname $0))
root_location=$(readlink -e "${ft_location}/..")

cd "${root_location}"

# step 1; run containerized tests
testParams=(
    "debian:bookworm-slim /dist/corteca*amd64.deb" \
    "debian:jessie-slim /dist/corteca*amd64.deb" \
    "centos:8 /dist/corteca*amd64.rpm" \
)

for param in "${testParams[@]}"; do
    img=$(echo "$param" | awk '{ print $1; }')
    pkg=$(echo "$param" | awk '{ print $2; }')
    docker run --rm \
        -v "./functionalTests:/ft/" \
        -v "./dist/packages:/dist/" \
        "${img}" \
        /bin/bash -c "/ft/functional_tests.sh ${img} ${pkg}"
done

# step 2; run host-based tests (to be containerized)
${ft_location}/build_deploy.sh ${CORTECAVM_ARMV7_NAME} ${CORTECAVM_ARMV8_NAME}
