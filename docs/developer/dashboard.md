Dashboard Developer Doc
-----

## How to develop

### API Server

* Require

Please make sure the `~/.kubec/config` already present.

* Start Api Server

```
$ go run cmd/server/main.go
2020/02/23 09:55:30 Use OutCluster Config Mode
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:   export GIN_MODE=release
 - using code:  gin.SetMode(gin.ReleaseMode)
...
[GIN-debug] Listening and serving HTTP on :8000
```

This will create a web server on `:8080`

### Web UI

Install denpendencies

> require: node v9.11.1+

```
npm install -g cnpm
cnpm install
npm start
```

Open Browser `http://localhost:3000/`