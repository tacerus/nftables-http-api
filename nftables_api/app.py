from falcon import App
from .resources import nft_set
from .config import config
from nftables_api.middlewares.authentication import AuthMiddleWare

app = App(
  middleware=[
    AuthMiddleWare(),
  ]
)

rSet = nft_set.SetResource()

app.add_route('/set/{xfamily}/{xtable}/{xset}', rSet)
