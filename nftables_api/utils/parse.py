import json


def parse_nft_error(err, raw):
  if isinstance(err, str):
    if '\n' in err:
      err = err.split('\n')
    elif err == '':
      err = []

  if raw:
    return err

  if isinstance(err, list) and len(err) > 0 and 'Error: ' in err[0]:
    return err[0].replace('Error: ', '')

  return err


def parse_nft_output(rc, out):
  if rc == 0 and out != '':
    status = True
    out_parsed = json.loads(out)

  elif rc == 0 and out == '':
    status = True
    out_parsed = None

  else:
    status = False
    out_parsed = {'nftables': []}

  return out_parsed, status


def parse_nft_response(rc, out, err, raw):
  return *parse_nft_output(rc, out), parse_nft_error(err, raw)
