from falcon import App
from .resources import nft_set

app = App()

rSet = nft_set.SetResource()

app.add_route('/set/{xfamily}/{xtable}/{xset}', rSet)
