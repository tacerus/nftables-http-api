"""
A RESTful HTTP API for nftables
Copyright 2024, Georg Pfuetzenreuter <mail@georg-pfuetzenreuter.net>

Licensed under the EUPL, Version 1.2 or - as soon they will be approved by the European Commission - subsequent versions of the EUPL (the "Licence").
You may not use this work except in compliance with the Licence.
An English copy of the Licence is shipped in a file called LICENSE along with this applications source code.
You may obtain copies of the Licence in any of the official languages at https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12.
"""

from bcrypt import checkpw
from falcon import HTTPUnauthorized

from nftables_api.config import config


class AuthMiddleWare:
  def _match(self, token_plain, token_hashed):
    """
    Check plain token against bcrypt hash
    """
    try:
      return checkpw(token_plain.encode(), token_hashed.encode())
    except ValueError:
      return False


  def _valid(self, token):
    """
    Check if token is contained in the configuration
    """
    for config_token, config_paths in config.get('tokens', {}).items():
      if self._match(token, config_token):
        return True
    return False


  def process_request(self, req, resp):  # noqa ARG002, resp is not used but needs to be passed by Falcon
    """
    Rudimentary token validation - check if it is worth walking further down the authorization chain
    """
    token = req.get_header('X-NFT-API-TOKEN')

    if token is None:
      raise HTTPUnauthorized(
        title='Authentication required',
      )

    if not self._valid(token):
      raise HTTPUnauthorized(
        title='Unauthorized',
      )


  def process_resource(self, req, resp, resource, params):  # noqa ARG002, resp is not used but needs to be passed by Falcon
    """
    Fully validate whether a token is authorized to perform the request
    """
    token = req.get_header('X-NFT-API-TOKEN')
    resource_name = resource._name()

    for config_token, config_paths in config.get('tokens', {}).items():
      if not self._match(token, config_token):
        continue

      for got_config_path, methods in config_paths.items():
        if not isinstance(methods, list):
          raise RuntimeError(f'Invalid method configured for path {got_config_path}')

        # a leading slash causes an empty first list entry in the split
        if got_config_path.startswith('/'):
          config_path = got_config_path[1:]
        else:
          config_path = got_config_path

        path_elements = config_path.split('/')

        if path_elements[0] != resource_name:
          continue

        path_elements = path_elements[1:]

        if resource_name == 'set':
          need_elements = ['xfamily', 'xtable', 'xset']

        for i, need_element in enumerate(need_elements):
          if path_elements[i] == '*':
            continue
          if path_elements[i] != params.get(need_element):
            break

        else:
          if req.method not in methods:
            raise HTTPUnauthorized(
              title='Unauthorized method for path',
            )

          return

    raise HTTPUnauthorized(
      title='Unauthorized',
    )
