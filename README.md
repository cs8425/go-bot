# simple bot
a simple reverse bot in golang.


## bot

### key
RSA Pub * 1 for hub

### build
```
go build -ldflags='-s -w' bot.go
```


## hub

### key
RSA Priv * n for bot
ECDSA Pub * n for admin

### build
```
go build hub.go share.go
```


## admin

### key
RSA Pub * 1 for hub
ECDSA Priv * 1 for hub

### build
```
go build admin.go share.go
```


### tool

`genkeys.go` : RSA & ECDSA keys generator



