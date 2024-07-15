#!/bin/bash

INFO_COLOR="\e[30m\e[44m"
SUCCESS_COLOR="\e[30m\e[42m"
FAIL_COLOR="\e[30m\e[41m"
RESET="\e[0m"

begin_test() {
    local test_name=$1
    echo -e "\n${INFO_COLOR}[TEST START]${RESET} Starting '$1'...\n"
}

assert_test_equal() {
    local test_name=$1
    local actual_value=$2
    local expected_value=$3
    if [ "${expected_value}" != "${actual_value}" ]; then
        echo -e "\n${FAIL_COLOR}[ERROR]${RESET} '${test_name}' failed:"
        echo -e "\texpected value: '${expected_value}'"
        echo -e "\tactual value: '${actual_value}'"
        exit 1
    else
        echo -e "\n${SUCCESS_COLOR}[SUCCESS]${RESET} '${test_name}' passed"
    fi
}

assert_test_notequal() {
    local test_name=$1
    local actual_value=$2
    local nonexpected_value=$3
    if [ "${nonexpected_value}" == "${actual_value}" ]; then
        echo -e "${SUCCESS_COLOR}[ERROR]${RESET} {$test_name} failed:"
        echo -e "\tfailure value: '${expected_value}'"
        exit 1
    fi
}

test_global_config_get() {
    begin_test "Testcase: global corteca config get"
    corteca config get 
    assert_test_equal "corteca config get" $? 0
    local result=$(corteca config get build.default)
    assert_test_equal "corteca config get build.default" "${result}" "aarch64"
}

test_user_config_get() {
    begin_test "Testcase: user config get"
    mkdir -p $HOME/.config/corteca
    echo -e "build:\n    default: foobar\n" > $HOME/.config/corteca/corteca.yaml
    local result=$(corteca config get build.default)
    assert_test_equal "corteca config get build.default" "${result}" "foobar"
}

test_create_test_template_with_cmd () {
    begin_test "Testcase: corteca create test_template"
    local applang="test-template"
    local apptitle="test-application"
    local appname="test-app"
    local appversion="1.1.1"
    local appfqdn="test.application.org"
    local appauthor="author"
    local expected_layout="./corteca.yaml ./src/folder_should.exist/should.exist ./src/should.exist ./src/${appname}.file"
    # sort expected layout
    expected_layout=$(echo "${expected_layout}" | xargs -n1 | sort)
    mkdir -p $HOME/.config/corteca/templates
    cp -r /ft/test-template $HOME/.config/corteca/templates/

    corteca create test-template \
        --skipPrompts \
        --config app.lang="${applang}" \
        --config app.title="${apptitle}" \
        --config app.name="${appname}" \
        --config app.version="${appversion}" \
        --config app.fqdn="${appfqdn}" \
        --config app.author="${appauthor}"

    assert_test_equal "corteca create test-template" $? 0
    actual_layout=$(cd test-template && find . -type f | sort)
    assert_test_equal "produced folder layout" "$actual_layout" "${expected_layout}"
    # validate non-blank DUID
    local duid=$(corteca config get app.duid)
    assert_test_notequal "corteca config get app.duid" "${duid}" ""
}

create_template_app() {
    lang="$1"
    local apptitle="${lang}-application"
    local appname="${lang}-app"
    local appversion="1.1.1"
    local appfqdn="${lang}.application.org"
    local appauthor="author"

    echo "Creating ${lang} application..."
    ./dist/bin/corteca-linux-amd64-* -r ./data/ create ${lang} \
        --skipPrompts \
        --config app.lang="${lang}" \
        --config app.title="${apptitle}" \
        --config app.name="${appname}" \
        --config app.version="${appversion}" \
        --config app.fqdn="${appfqdn}" \
        --config app.author="${appauthor}"

    chmod -R a+w ./ 2>/dev/null | echo ignoring errors
    sleep 5
}

debian_install() {
    begin_test "Testcase: install cortecacli on debian"
    dpkg --ignore-depends=docker.io,docker-ce,podman-docker -i $1
    assert_test_equal "dpkg -i '$1'" $? 0
}

centos_install() {
    begin_test "Testcase: install cortecacli on centos:8"
    rpm -ivh --nodeps $1
    assert_test_equal "rpm -ivh --nodeps '$1'" $? 0
}

debian_uninstall() {
    begin_test "Testcase: uninstall cortecacli on debian"
    dpkg --purge corteca-cli
    assert_test_equal "dpkg --purge" $? 0
}

centos_uninstall() {
    begin_test "Testcase: uninstall cortecacli on centos:8"
    rpm -e corteca-cli
    assert_test_equal "rpm -e" $? 0
}
