#!/bin/sh -ex

u='user'
ud="/home/$u"
wd='/work'
podman run                                                         \
	--cap-add CAP_NET_ADMIN                                    \
	--pull=always                                              \
	--rm                                                       \
	--userns=keep-id                                           \
	-u="$u"                                                    \
	-it                                                        \
	-v "$PWD"/test/_cache/GOCACHE:"$ud"/.cache/go-build:rw     \
	-v "$PWD"/test/_cache/GOMODCACHE:"$ud"/go/pkg/mod:rw       \
	-v .:"$wd"                                                 \
	-w "$wd"                                                   \
	registry.opensuse.org/home/crameleon/containers/containers/go-nftables:latest \
	env NFT_HTTP_API_TEST_DESTRUCTIVE=yes go test -v ./...
