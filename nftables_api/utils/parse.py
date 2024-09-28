"""
A RESTful HTTP API for nftables
Copyright 2024, Georg Pfuetzenreuter <mail@georg-pfuetzenreuter.net>

Licensed under the EUPL, Version 1.2 or - as soon they will be approved by the European Commission - subsequent versions of the EUPL (the "Licence").
You may not use this work except in compliance with the Licence.
An English copy of the Licence is shipped in a file called LICENSE along with this applications source code.
You may obtain copies of the Licence in any of the official languages at https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12.
"""

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
