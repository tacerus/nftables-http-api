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


  def process_request(self, req, resp):
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


  def process_resource(self, req, resp, resource, params):
    """
    Fully validate whether a token is authorized to perform the request
    """
    token = req.get_header('X-NFT-API-TOKEN')
    resource_name = resource._name()

    for config_token, config_paths in config.get('tokens', {}).items():
      if not self._match(token, config_token):
        continue

      for config_path, methods in config_paths.items():
        if not isinstance(methods, list):
          raise RuntimeError(f'Invalid method configured for path {config_path}')

        # a leading slash causes an empty first list entry in the split
        if config_path.startswith('/'):
          config_path = config_path[1:]

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
          if req.method in methods:
            return
          else:
            raise HTTPUnauthorized(
              title='Unauthorized method for path',
            )

    raise HTTPUnauthorized(
      title='Unauthorized',
    )
