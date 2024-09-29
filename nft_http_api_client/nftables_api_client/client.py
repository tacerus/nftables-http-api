"""
A RESTful HTTP API client for nftables
Copyright 2024, Georg Pfuetzenreuter <mail@georg-pfuetzenreuter.net>

Licensed under the EUPL, Version 1.2 or - as soon they will be approved by the European Commission - subsequent versions of the EUPL (the "Licence").
You may not use this work except in compliance with the Licence.
An English copy of the Licence is shipped in a file called LICENSE along with this applications source code.
You may obtain copies of the Licence in any of the official languages at https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12.
"""

from os import getenv

import urllib3


class NftablesRemote:
  def __init__(self, endpoint, token=None):
    if token is None:
      token = getenv('NFT-API-TOKEN')

    if token is None:
      raise ValueError('Missing token, pass one as an argument or via NFT-API-TOKEN.')

    self.endpoint = endpoint

    self.auth_headers = {
      'X-NFT-API-Token': token,
    }


  def get(self, path):
    retmap = {
      'status': -1,
      'data': {},
    }

    response = urllib3.request(
      'GET',
      f'{self.endpoint}/{path}',
      headers=self.auth_headers,
    )

    retmap['status'] = response.status
    retmap['data'] = response.json()

    return retmap
