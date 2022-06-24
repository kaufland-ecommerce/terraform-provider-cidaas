TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=local
NAMESPACE=team-account
NAME=cidaas
BINARY=terraform-provider-${NAME}
VERSION=0.1.0
OS_ARCH=darwin_amd64

default: install

build:
	go build -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

initgpg:
	cat $(GPG_PRIVATE_KEY) | gpg --import --batch --no-tty
	echo "hello world" > temp.txt
	gpg --detach-sig --yes -v --output=/dev/null --pinentry-mode loopback --passphrase "$(PASSPHRASE)" temp.txt
	rm temp.txt

prepare: initgpg
	curl -sfL https://goreleaser.com/static/run | bash -s -- release --rm-dist --skip-publish

publish:
	echo "PUBLISHING"
