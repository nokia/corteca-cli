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

test_global_config_add(){
    begin_test "Testcase: global corteca config --global add"
    corteca config --global add devices "foo: { addr: bar }"
    assert_test_equal "corteca config add" $? 0

    local result=$(corteca config --global get devices.foo )
    assert_test_equal "corteca config --global add devices.foo" "${result}" "addr: bar"
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
    local appname="test-app"
    local appversion="1.1.1"
    local appfqdn="test.application.org"
    local appauthor="author"
    local expected_layout="./corteca.yaml ./src/${appname}.file ./.corteca/test.template"
    # sort expected layout
    expected_layout=$(echo "${expected_layout}" | xargs -n1 | sort)
    mkdir -p $HOME/.config/corteca/templates
    cp -r /ft/test-template $HOME/.config/corteca/templates/
    cp -r /ft/_baseTestTemplate $HOME/.config/corteca/templates/

    corteca create test-template \
        --skipPrompts \
        --lang="${applang}" \
        --fqdn="${appfqdn}" \
        --config app.name="${appname}" \
        --config app.version="${appversion}" \
        --config app.author="${appauthor}"

    assert_test_equal "corteca create test-template" $? 0
    actual_layout=$(cd test-template && find . -type f | sort)
    assert_test_equal "produced folder layout" "$actual_layout" "${expected_layout}"
    # validate non-blank DUID
    local duid=$(corteca config get app.duid)
    assert_test_notequal "corteca config get app.duid" "${duid}" ""

    test_global_config_add
}

test_regen() {
    cd test-template
    begin_test "Testcase: corteca regen template files"

    corteca config set app.name test-template-regen

    #check project layout
    expected_layout="./.corteca/test.template ./corteca.yaml ./src/test-app.file ./test"
    actual_layout=$(find . -type f | sort | tr '\n' ' ' | sed 's/[[:space:]]*$//')
    assert_test_equal "updated folder layout after regenerating template files" "${actual_layout}" "${expected_layout}"

    #check if files have been updated
    grep -q "test-template-regen" ./test
    assert_test_equal "regenerated file updated with new content" $? 0

    cd ..

    test_global_config_add
}

test_build() {
    langs=("go" "c" "cpp")
    archs=("armv7l" "aarch64")

    for lang in "${langs[@]}"; do
        create_template_app "${lang}"
        cd "${lang}"
        for arch in "${archs[@]}"; do
            begin_test "Testcase: corteca build ${lang} application for ${arch}"
            ../dist/bin/corteca-linux-amd64-* -r ../data/ config set build.options.outputType rootfs
            ../dist/bin/corteca-linux-amd64-* -r ../data/ build ${arch}
            assert_test_equal "corteca build ${arch}" $? 0
        done
        cd ../
    done
}

start_cortecavm() {
    container_name="$1"
    arch="$2"
    port="$3"

    echo "Running cortecaVM-${arch}..."
    docker run -d --rm --name "${container_name}" -e QEMU_PORTS="${port}:8022" ni-bbd-container-apps-local.artifactory-espoo1.int.net.nokia.com/rc/cortecavm-${arch}:23.4.3
}

test_exec(){
    container_name_armv7="$1"
    container_name_armv8="$2"
    exec_command="$3"
    to_be_published="$4"
    flag_artifact="$5"

    langs=("go" "c" "cpp")
    vm_archs=("armv7" "armv8")

    declare -A arch_map=(
        ["armv7"]="8027:${container_name_armv7}:armv7l"
        ["armv8"]="8028:${container_name_armv8}:aarch64"
    )

    for lang in "${langs[@]}"; do
        cd "${lang}"
        cp ./corteca.yaml ./corteca.yaml.original
        for vm_arch in "${vm_archs[@]}"; do
        
            begin_test "Testcase: corteca ${exec_command} ${lang} application for ${vm_arch}"

            cp ./corteca.yaml.original ./corteca.yaml
            publish_port=$(shuf -i 8000-8999 -n 1)
            map="${arch_map[$vm_arch]}"
            port="${map%%:*}"
            temp="${map%:*}"
            container_name="${temp##*:}"
            tmp_arch="${map#*:}"
            architecture="${tmp_arch#*:}"
            ip=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "${container_name}")

            prepare_yaml "${publish_port}" "${vm_arch}" "${ip}" "${port}"
            add_sequence_to_yaml

            exists=$(is_sequence_exists "${exec_command}")

            wait_for_ssh_connectivity "root@${ip}" "${port}" 360 10

            if [ -n "${exists}" ]; then
                execute_command "${exec_command}" "${vm_arch}" "${to_be_published}" "${flag_artifact}"
            else
                assert_test_equal "corteca exec unknown command ${exec_command}" $? 0
            fi

            assert_test_equal "corteca exec ${exec_command}" $? 0
            
            if [ "${flag_artifact}" -eq 0 ]; then
                prepare_yaml_global "${publish_port}" "${vm_arch}" "${ip}" "${port}"
                test_artifact_flag "${architecture}" "${lang}" "${vm_arch}"
            fi
        done
        
        cd ../
        cp ./${lang}/corteca.yaml.original ./${lang}/corteca.yaml
    done
}

test_artifact_flag(){
    architecture="$1"
    lang="$2"
    vm_arch="$3"

    begin_test "Testcase: corteca exec deploy --artifact ${lang} application for ${vm_arch}"
    
    cd ../

    ./dist/bin/corteca-linux-amd64-* -r ./data/ exec deploy ${vm_arch} --publish local --global --artifact ${architecture}:rootfs:./${lang}/dist/${lang}-app-1.1.1-${architecture}-rootfs.tar.gz
    assert_test_equal "corteca exec deploy --artifact" $? 0
    cd "${lang}"
}

execute_command(){
    exec_command="$1"
    vm_arch="$2"
    to_be_published="$3"

    if [ "${to_be_published}" -eq 0 ]; then
        ../dist/bin/corteca-linux-amd64-* -r ../data/ exec ${exec_command} ${vm_arch} --publish local
    else
        ../dist/bin/corteca-linux-amd64-* -r ../data/ exec ${exec_command} ${vm_arch}
    fi
    assert_test_equal "corteca exec ${exec_command}" $? 0
}

is_sequence_exists(){
    seq_name="$1"
    exists=$(../dist/bin/corteca-linux-amd64-* -r ../data/ config get sequences.${seq_name})

    echo ${exists}
}

prepare_yaml_global(){
    publish_port="$1"
    vm_arch="$2"
    ip="$3"
    port="$4"

    ../dist/bin/corteca-linux-amd64-* -r ../data/ config --global add publish "local: { addr: http://0.0.0.0:${publish_port}, method: listen, publicURL: http://172.17.0.1:${publish_port}}"

    ../dist/bin/corteca-linux-amd64-* -r ../data/ config --global add devices "${vm_arch}: { addr: ssh://root@${ip}:${port}}"

}

prepare_yaml(){
    publish_port="$1"
    arch="$2"
    ip="$3"
    port="$4"

    ../dist/bin/corteca-linux-amd64-* -r ../data/ config add publish "local: { addr: http://0.0.0.0:${publish_port}, method: listen, publicURL: http://172.17.0.1:${publish_port}}"

    ../dist/bin/corteca-linux-amd64-* -r ../data/ config add devices "${vm_arch}: { addr: ssh://root@${ip}:${port}}"
}

add_sequence_to_yaml(){
    echo "sequences:" >> ./corteca.yaml
    echo "    create: " >> ./corteca.yaml
    echo "        - cmd: touch test_seq.txt" >> ./corteca.yaml
    echo "          delay: 1000" >> ./corteca.yaml
    echo "    add: " >> ./corteca.yaml
    echo "        - cmd: \$(create)" >> ./corteca.yaml
    echo "        - cmd: echo \"Testing sequences\" >> test_seq.txt" >> ./corteca.yaml
    echo "          delay: 1000" >> ./corteca.yaml
    echo "        - cmd: ls -al | grep test_seq.txt" >> ./corteca.yaml
    echo "          delay: 1000" >> ./corteca.yaml
    echo "          retries: 2" >> ./corteca.yaml
    echo "    testSequence: " >> ./corteca.yaml
    echo "        - cmd: \$(add)" >> ./corteca.yaml
    echo "        - cmd: cat test_seq.txt | grep \"Testing sequences\"" >> ./corteca.yaml
    echo "          delay: 1000" >> ./corteca.yaml
    echo "          retries: 2" >> ./corteca.yaml
}

wait_for_ssh_connectivity() {
    local url=$1
    local port=$2
    local max_time=$3
    local interval=$4

    echo -e "Waiting for SSH connectivity on ${url}:${port}...\n\tmax time: ${max_time} second(s)\n\tpoll interval: ${interval} second(s)\n"
    start_time=$(date +%s)
    while ! ssh -q ${url} -p ${port} -o StrictHostKeyChecking=no -o ConnectTimeout=${interval} true; do
        current_time=$(date +%s)
        elapsed_time=$((current_time - start_time))
        if [ $elapsed_time -ge ${max_time} ]; then
            echo "Timeout after ${max_time} second(s)"
            exit 1
        fi
        sleep ${interval}
    done
    end_time=$(date +%s)
    total_time=$((end_time - start_time))
    echo -e "SSH connectivity established after ${total_time} second(s)"
}

create_template_app() {
    lang="$1"
    local appname="${lang}-app"
    local appversion="1.1.1"
    local appfqdn="${lang}.application.org"
    local appauthor="author"

    echo "Creating ${lang} application..."
    ./dist/bin/corteca-linux-amd64-* -r ./data/ create ${lang} \
        --skipPrompts \
        --lang="${lang}" \
        --fqdn="${appfqdn}" \
        --config app.name="${appname}" \
        --config app.version="${appversion}" \
        --config app.author="${appauthor}"
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
