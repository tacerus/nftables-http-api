#!/bin/sh -ex

wd='/work'
podman run                \
	--cap-add CAP_NET_ADMIN \
	--pull=always           \
	--rm                    \
	-it                     \
	-v .:"$wd"              \
	registry.opensuse.org/home/crameleon/containers/containers/crameleon/pytest-nftables:latest \
	env NFT-API-CONFIG="$wd"/tests/config.yaml PYTHONPATH="$wd" pytest --pdb --pdbcls=IPython.terminal.debugger:Pdb -rA -s -v -x "$wd"/tests

