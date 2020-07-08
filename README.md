# tccbot-backend
![Go](https://github.com/tagirmukail/tccbot-backend/workflows/Go/badge.svg?branch=master)

This bot is trading cryptocurrency on bitmex exchange.

#### Run

```bash
make build
./tccbot-backend -level 5
```

If you need to output the log to a file
```bash
mkdir logs
./tccbot-backend -logdir logs
```

If you need to run this bot on testnet, install testnet key and secret in [config-example.yaml#testnet](config-yaml/config-example.yaml#L41)
```bash
./tccbot-backend -test
```

Run bot with your configuration
```bash
./tccbot-backend -config {your config file path}
```

#### Configuration
See [config-example.yaml](config-yaml/config-example.yaml)
