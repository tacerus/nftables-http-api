from yaml import safe_load
from os import getenv

configpath = getenv('NFT-API-CONFIG')
if not configpath:
  raise RuntimeError('NFT-API-CONFIG is not set')

with open(configpath) as fh:
  configdata = safe_load(fh)

config = configdata.get('nft-api', {})

if not config:
  raise RuntimeError('Invalid configuration data')
