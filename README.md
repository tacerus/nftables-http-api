# RESTful HTTP API for nftables

Early work in progress.

## Configuration

A configuration file must be passed using the environment variable `NFT-API-CONFIG`. The file contains a mapping of bcrypt hashed tokens to API paths and methods. Requests to the API are only authorized if the token passed using the header `X-NFT-API-Token` and the requested path and method match an entry in the configuration.

### Sample configuration:

```
nft-api:
  tokens:
    $2y$05$ZifkrfFg2XZU2ds7Lrcl9usJVyxHro9Ezjo84OMpsBSau4pEu42eS:
      /set/inet/filter/foo4:
        - GET
        - POST
    $2y$05$ZifkrfFg2XZU2ds7Lrcl9usJVyxHro9Ezjo84OMpsBSau4pEu42eS:
      /set/inet/filter/*:
        - GET
```

Any elements can contain a wildcard in the form of a `*` character.

Generate token hashes using any bcrypt hashing tool, such as `htpasswd` from the `apache-utils` suite:

```
$ htpasswd -Bn x
```

The username (here `x`) is not used, ignore the relevant part before and including the colon in the output.

## Deployment

The application can be served using any Python WSGI server. Easiest is to use `waitress` using the script found in the repository root:

```
./nftables-api.py
```

It will bind on `*:9090`.

## Usage

Currently `GET` and `POST` methods are implemented, allowing to query or append data respectively.

All requests must contain a header `X-NFT-API-Token` containing a token authorized for the requested path and method. All `POST` requests must contain a JSON payload.

By default, the response body will be output deemed most "useful" - in case of `GET` requests, that is only the data found for the given query, in case of `POST` requests the result of the change. Optionally, a query parameter `raw` can be passed to retrieve a mapping containing all status information, including the raw `libnftables-json(5)` output, instead.

Below are examples using `curl`.

### nftables sets

#### Get a set

```
$ curl -H 'X-NFT-API-Token: foo' localhost:9090/set/inet/filter/foo4
["192.168.0.0/24", "192.168.1.1", "192.168.3.0/24", "192.168.4.0/24", "192.168.5.0/26"]⏎
```

#### Append to a set

```
$ curl -H 'X-NFT-API-Token: foo' --json '{"addresses": "fe80::/64"}' localhost:9090/set/inet/filter/foo
{"status": true}⏎
```
