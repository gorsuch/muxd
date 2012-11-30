# muxd

An http abstraction on top of redis pubsub.

## Build

```bash
$ go get
```

## Deps

Redis

## Deployment

```bash
$ PORT=8080 REDIS_URL=redis://localhost:6379 ./muxd
```

## Usage

```bash
# listen for items coming into foobar
$ curl http://localhost:8080?channel=foobar
```

```bash
# write to foobar
$ curl -X POST -d data=neato http://localhost:8080?channel=foobar
```

Also see [mux](https://github.com/gorsuch/mux) a cli tool built to interact with muxd.