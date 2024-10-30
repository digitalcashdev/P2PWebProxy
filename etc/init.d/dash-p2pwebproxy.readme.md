```sh
mkdir -p ~/.config/dash-p2pwebproxy/
chmod 0700 ~/.config
chmod 0700 ~/.config/dash-p2pwebproxy

touch ~/.config/dash-p2pwebproxy/env
chmod 0600 ~/.config/dash-p2pwebproxy/env
```

```sh
mkdir -p ~/srv/dash-p2pwebproxy/
```

```sh
sudo chmod a+x /etc/init.d/dash-p2pwebproxy
sudo rc-update add dash-p2pwebproxy
sudo rc-service dash-p2pwebproxy start
```
