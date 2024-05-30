# dsnotify
Trivial golang stream event notifier for discord.

## Setup
Two required config files must be provided.

1. Auth config (default `./auth.yaml`), which includes the token or oauth2 key for the application.
```
token: <authentication token>
```
2. DSNotify config (default `./dsnotify.yaml`), which contains the guilds on which to listen, channel/role to notify, and enablement/debug settings. Multiple servers may be provided.
```
<First Guild ID>:
  channel: <Channel ID to notify>
  role: <Role to mention>
  debug: False
  enabled: True
<Second Guild ID>:
  channel: <Channel ID to notify>
  role: <Role to mention>
  debug: False
  enabled: False
```

## Run
```
$ ls
auth.yaml  dsnotify.yaml  main.go

$ go run main.go -h
Usage of /tmp/go-build1647414645/b001/exe/main:
  -auth string
        Auth config filename (default "./auth.yaml")
  -config string
        Server config filename (default "./dsnotify.yaml")
  -ready string
        Ready message (default "Discord Stream Notify")

$ go run main.go
2024/05/29 23:26:54 [INFO] DSNotify is now playing Discord Stream Notify.  Press CTRL-C to exit.
```
