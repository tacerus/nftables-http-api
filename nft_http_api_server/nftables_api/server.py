"""
A RESTful HTTP API for nftables
Copyright 2024, Georg Pfuetzenreuter <mail@georg-pfuetzenreuter.net>

Licensed under the EUPL, Version 1.2 or - as soon they will be approved by the European Commission - subsequent versions of the EUPL (the "Licence").
You may not use this work except in compliance with the Licence.
An English copy of the Licence is shipped in a file called LICENSE along with this applications source code.
You may obtain copies of the Licence in any of the official languages at https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12.
"""

from waitress import serve

from nftables_api.app import app


def server():
  if __name__ == '__main__':
    serve(app, host='*', port=9090)
  else:
    raise RuntimeError('serve() is intended to be used as an entrypoint.')
