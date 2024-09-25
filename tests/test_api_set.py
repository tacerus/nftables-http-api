from pytest import mark
from falcon import HTTP_CREATED, HTTP_OK
from json import dumps, loads

vs = [4, 6]

@mark.parametrize('v', vs)
def test_get_set(client, nft_ruleset_populated_sets, v):
  want_out = {
    4: ["192.168.0.0/24", "127.0.0.1"],
    6: ["fd80::/64", "fe80::1"],
  }
  response = client.simulate_get(f'/set/inet/filter/testset{v}')
  have_out = loads(response.content)
  assert sorted(have_out) == sorted(want_out[v])
  assert response.status == HTTP_OK

@mark.parametrize('v', vs)
@mark.parametrize('plvariant', ['address', 'network'])
@mark.parametrize('plformat', ['string', 'list'])
def test_append_to_set(client, nft_ruleset_populated_sets, v, plvariant, plformat):
  nft = nft_ruleset_populated_sets

  # all the matrixes could be moved to parameters
  want_out = {
    4: ["192.168.0.0/24", "127.0.0.1", "192.168.5.0/26"],
    6: ["fd80::/64", "fe80::1", "fd10:f00::/128"],
  }

  if plformat == 'string':
    if plvariant == 'address':
      to_add = {
        4: '192.168.5.1',
        6: 'fd10:f00::',
      }
    elif plvariant == 'network':
      to_add = {
        4: '192.168.5.0/26',
        6: 'fd10:f00::/48',
      }
    added = to_add[v]
  elif plformat == 'list':
    if plvariant == 'address':
      to_add = {
        4: ['192.168.5.1'],
        6: ['fd10:f00::'],
      }
    elif plvariant == 'network':
      to_add = {
        4: ['192.168.5.0/26'],
        6: ['fd10:f00::/48'],
      }
    added = to_add[v][0]

  response = client.simulate_post(
    f'/set/inet/filter/testset{v}',
    body=dumps({
      'addresses': to_add[v],
    }),
    headers={
      'content-type': 'application/json',
    },
  )
  have_out = loads(response.content)

  assert have_out == {'status': True}
  assert response.status == HTTP_CREATED
  assert added in nft.cmd(f'list set inet filter testset{v}')[1]
