# simple bot
a simple backdoor for practice. It's been written in GO to easy cross-compiling static binary for different architectures and OSes.

## Functions
* [x] reverse connect
* [x] encrypted connection
* [x] built-in TCP multiplexing
* [x] socks5 Proxy (cmd 'CONNECT' only)
* [x] shell
* [ ] file operations
* [ ] self update
* [ ] downloader
* [ ] task, schedule
* [ ] Dos, DDoS
* [ ] miner

## bot

### key
* RSA Pub * 1 for hub

### build
```
go build -ldflags='-s -w' bot.go
```


## hub

### key
* RSA Priv * n for bot
* ECDSA Pub * n for admin

### build
```
go build hub.go share.go
```


## admin

### commands
* bot
	* ls [id | addr | time] : list all bot on hub by ...
	* kill
	* reconn
* local 
	* ls : list local side server
	* bind $bot_id $bind_addr $mode $argv... : bind server (eg. socks5) on local side
	* stop $bind_addr


### key
* RSA Pub * 1 for hub
* ECDSA Priv * 1 for hub

### build
```
go build admin.go share.go
```


## tool

`genkeys.go` : RSA & ECDSA keys generator



