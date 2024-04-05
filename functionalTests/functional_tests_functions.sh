#!/bin/bash

assert_test_equal() {
    local test_name=$1
    local actual_value=$2
    local expected_value=$3
    if [ "${expected_value}" != "${actual_value}" ]; then
        echo -e "ERROR: '${test_name}' failed:"
        echo -e "\texpected value: '${expected_value}'"
        echo -e "\tactual value: '${actual_value}'"
        exit 1
    fi
}

assert_test_notequal() {
    local test_name=$1
    local actual_value=$2
    local nonexpected_value=$3
    if [ "${nonexpected_value}" == "${actual_value}" ]; then
        echo -e "ERROR: {$test_name} failed:"
        echo -e "\tfailure value: '${expected_value}'"
        exit 1
    fi
}

test_global_config_show() {
    corteca config show
    assert_test_equal "corteca config show" $? 0
    local result=$(corteca config get build.default)
    assert_test_equal "corteca config get build.default" "${result}" "armv8"
}

test_user_config_show() {
    mkdir -p $HOME/.config/corteca
    echo -e "build:\n    default: foobar\n" > $HOME/.config/corteca/corteca.yaml
    local result=$(corteca config get build.default)
    assert_test_equal "corteca config get build.default" "${result}" "foobar"
}

test_create_test_template_with_cmd () {
    local applang="test-template"
    local apptitle="test-application"
    local appname="testApp"
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
        --config \
app.lang="${applang}",\
app.title="${apptitle}",\
app.name="${appname}",\
app.version="${appversion}",\
app.fqdn="${appfqdn}",\
app.author="${appauthor}"

    assert_test_equal "corteca create test-template" $? 0
    actual_layout=$(cd test-template && find . -type f | sort)
    assert_test_equal "produced folder layout" "$actual_layout" "${expected_layout}"
    # validate non-blank DUID
    local duid=$(corteca config get app.duid)
    assert_test_notequal "corteca config get app.duid" "${duid}" ""
}

debian_install() {
    dpkg --ignore-depends=docker.io,docker-ce,podman-docker -i $1
    assert_test_equal "dpkg -i '$1'" $? 0
}

centos_install() {
    rpm -ivh --nodeps $1
    assert_test_equal "rpm -ivh --nodeps '$1'" $? 0
}

debian_uninstall() {
    dpkg --purge corteca-cli
    assert_test_equal "dpkg --purge" $? 0
}

centos_uninstall() {
    rpm -e corteca-cli
    assert_test_equal "rpm -e" $? 0
}

