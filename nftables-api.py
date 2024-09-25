from waitress import serve
from nftables_api.app import app

if __name__ == '__main__':
    serve(app, host='*', port=9090)
