# P2PWebProxy

P2P Web Proxy to the MNOs of the Digital Cash Network (CORS-enabled)

## How to Self-Host Web Proxy + Explorer

0. Clone and enter the repo

    ```sh
    git clone https://github.com/digitalcashdev/P2PWebProxy.git
    pushd ./P2PWebProxy/
    ```

1. Install Go **v1.22+**

    ```sh
    curl https://webi.sh/go | sh
    source ~/.config/envman/PATH.env
    ```

2. Build and Run the Proxy \

    ```sh
    # Option 1: to run locally
    go build -o ./dash-p2pwebproxy ./cmd/dash-p2pwebproxy/

    # Option 2: to run on a server or in a container
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
         go build -o ./dash-p2pwebproxy-linux-x86_64 ./cmd/dash-p2pwebproxy/
    ```

    ```sh
    ./dash-p2pwebproxy --help

    ./dash-p2pwebproxy --port 8080 --testnet
    ```

## How to Register as a System Daemon

0. Install `serviceman`, `pathman`, and `setcap-netbind`

    ```sh
    curl https://webi.sh/ | sh
    source ~/.config/envman/PATH.env

    webi serviceman pathman setcap-netbind
    ```

1. Place `dash-p2pwebproxy` in your `PATH`

    ```sh
    mkdir -p ~/bin/
    pathman add ~/bin/
    source ~/.config/envman/PATH.env

    mv ./dash-p2pwebproxy ~/bin/
    ```

2. Allow binding to privileged ports \
   (optional: non-root install on Linux)

    ```sh
    setcap-netbind 'dash-p2pwebproxy'
    ```

3. Register the service

    ```sh
    sudo env PATH="$PATH" \
        serviceman add --name "dash-p2pwebproxy" --system --path="$PATH" -- \
        dash-p2pwebproxy --port 8080 --test
    ```

    (note: this also works with ENVs, see [./example.env](/example.env))
