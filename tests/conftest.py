"""
Tests for the RESTful HTTP API for nftables
Copyright 2024, Georg Pfuetzenreuter <mail@georg-pfuetzenreuter.net>

Licensed under the EUPL, Version 1.2 or - as soon they will be approved by the European Commission - subsequent versions of the EUPL (the "Licence").
You may not use this work except in compliance with the Licence.
An English copy of the Licence is shipped in a file called LICENSE along with this applications source code.
You may obtain copies of the Licence in any of the official languages at https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12.
"""

from falcon import testing
from nftables import Nftables
from nftables_api.app import app
from pytest import exit, fixture


def run_nft(nft, cmd):
  rc, out, err = nft.cmd(cmd)
  if rc != 0:
    print(out, err)
    exit()

@fixture
def client():
  return testing.TestClient(app)

@fixture
def nft():
  nft = Nftables()
  run_nft(nft, 'add table inet filter')
  yield nft
  nft.cmd('flush ruleset')

@fixture
def nft_ruleset_empty_sets(nft):
  for i in [4, 6]:
    run_nft(nft, f'add set inet filter testset{i} {{ type ipv{i}_addr ; flags interval ; }}')

  return nft

@fixture
def nft_ruleset_populated_sets(nft_ruleset_empty_sets):
  nft = nft_ruleset_empty_sets
  run_nft(nft, 'add element inet filter testset4 { 192.168.0.0/24, 127.0.0.1 }')
  run_nft(nft, 'add element inet filter testset6 { fd80::/64, fe80::1 }')
  return nft
