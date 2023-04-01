HOSTNAME=github.com
NAMESPACE=LeMikaelF
NAME=meetingsascode
BINARY=terraform-provider-${NAME}
VERSION=0.1
OS_ARCH=linux_amd64

.PHONY: auth

default: install

auth:
	@echo $$(go run ./auth)

install:
	go build -o ${BINARY} ./provider
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
