import falcon
import json
from nftables import Nftables
from nftables_api.utils.parse import parse_nft_response
from nftables_api.utils.output import output_post

nft = Nftables()
nft.set_json_output(True)

class SetResource:
  def _name(self):
    return 'set'

  def on_get(self, request, response, xfamily, xtable, xset):
    raw = request.get_param_as_bool('raw', default=False)

    rc, out, err = nft.cmd(f'list set {xfamily} {xtable} {xset}')
    out_parsed, status, err_parsed = parse_nft_response(rc, out, err, raw)

    if raw:
      response.text = json.dumps({'rc': rc, **out_parsed, 'err': err_parsed, 'status': status})
      return

    out_parsed = out_parsed.get('nftables', [])
    elements = []

    if len(out_parsed) > 1 and isinstance(out_parsed[1], dict):
      elements_low = out_parsed[1].get('set', {}).get('elem', [])
      for element in elements_low:
        if isinstance(element, dict):
          prefix = element.get('prefix', {})
          address = prefix.get('addr')
          length = prefix.get('len')

          elements.append(f'{address}/{length}')

        elif isinstance(element, str):
          elements.append(element)

    else:
      status = False

    if status is True:
      response.text = json.dumps(elements)

    else:
      response.text = json.dumps({'status': status, 'error': err_parsed})


  def on_post(self, request, response, xfamily, xtable, xset):
    raw = request.get_param_as_bool('raw', default=False)
    data = request.get_media()

    if not isinstance(data, dict) or not ( 'address' in data or 'addresses' in data ):
      response.text = output_post(status=False, message='Invalid body.')
      return

    addresses = []

    for sp in ['address', 'addresses']:
      if sp in data:
        if isinstance(data[sp], str):
          addresses.append(data[sp])
        elif isinstance(data[sp], list):
          addresses.extend(data[sp])

    elements = []

    for address in addresses:
      if '/' in address:
        addrsplit = address.split('/')
        elements.append({
          'prefix': {
            'addr': addrsplit[0],
            'len': int(addrsplit[1]),
          }
        })

      else:
        elements.append(address)

    nft_payload = {
      'nftables': [
        {
          'add': {
            'element': {
              'elem': elements,
              'family': xfamily,
              'name': xset,
              'table': xtable,
            }
          }
        }
      ]
    }

    if not nft.json_validate(nft_payload):
      response.status = falcon.HTTP_BAD_REQUEST
      response.text = output_post(False, 'Payload did not validate.')
      return

    rc, out, err = nft.json_cmd(nft_payload)
    out_parsed, status, err_parsed = parse_nft_response(rc, out, err, raw)

    if status is True:
      response.status = falcon.HTTP_CREATED
    else:
      response.status = falcon.HTTP_BAD_REQUEST

    if raw:
      response.text = json.dumps({'rc': rc, 'err': err_parsed, 'status': status})

    else:
      response.text = output_post(status=status, message=err_parsed)
