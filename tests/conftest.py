from falcon import testing
from pytest import exit, fixture
from nftables import Nftables

from nftables_api.app import app

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
