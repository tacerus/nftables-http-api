# RESTful HTTP API for nftables - Client

## Usage

The token can be passed either as an argument to `NftablesRemote()` or in the environment variable `NFT-API-TOKEN`.

```
# initialize
from nftables_api_client.client import NftablesRemote
nft = NftablesRemote('http://localhost:9090', 'mytoken')

# do things
nft.get('/set/inet/filter/foo')
```
