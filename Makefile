_app_name=no-phi-ai
_app_pkg_dir=pkg
_build_dir=build
_build_mode=pie
_build_packages="./${_app_pkg_dir}/..."
_coverage_out_file=coverage.out

_cmd_go_build=go build -buildmode="${_build_mode}" -o ../${_build_dir}/${_app_name}
_cmd_go_cover=go tool cover -func=${_coverage_out_file}
_cmd_go_fmt=go fmt ${_build_packages}
_cmd_go_tidy=go mod tidy
_cmd_go_test=go test -buildmode="${_build_mode}" -cover -coverprofile ${_coverage_out_file} -v -timeout=30s
_cmd_test=$$(go list ${_build_packages} | grep -v 'vendor')

_msg="${_app_name} : make"
_msg_error="ERROR : ${_msg}"
_msg_success="SUCCESS : ${_msg}"

default: build

.PHONY: build clean deploy format package remove test tidy

build: build_prep build_only

build_only:
	cd ${_app_pkg_dir} && ${_cmd_go_build} \
		&& echo "${_msg_success} build_only" \
		|| (echo "${_msg_error} build_only" && exit 70)

build_prep: format tidy test

clean:
	rm -rf ./${_build_dir}/${_app_name} ./vendor/ Gopkg.lock ${_coverage_out_file}

format:
	${_cmd_go_fmt} \
		&& echo "${_msg_success} format" \
		|| (echo "${_msg_error} format" && exit 40)

test:
	echo $(_cmd_test) | xargs ${_cmd_go_test} \
		&& echo "${_msg_success} test" \
		|| (echo "${_msg_error} test" && exit 60)

test_build: tidy build_only test

test_cover: test
	${_cmd_go_cover}

tidy:
	${_cmd_go_tidy} \
		&& echo "${_msg_success} tidy" \
		|| (echo "${_msg_error} tidy" && exit 50)

