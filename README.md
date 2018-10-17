# simple bot
a simple backdoor/remote-admin-tool for practice. It's been written in GO to easy cross-compiling static binary for different architectures and OSes.

## Features
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
* [x] connection in simple tcp
* [ ] hide connection in http/https/ws
* [ ] run in P2P mode

## bot/client

### key

* RSA Pub * 1 for connect to hub

### ops

* basic:
  * `info`: pull info
  * `csh`: shell
  * `fastN`: socks
  * `reconn`: re-connect
  * `kill`: self-exit
* mod:
  * `ppend`: send `SIGTERM` (15) to parent process
  * `ppkill`: send `SIGKILL` (9) to parent process
  * `psig`: send signal to process
* extra: some are buggy
  * [x] `dodaemon`: daemonize
  * [?] `apoptosis`: remove self's binary without exit
  * [?] `rebirth`: put self's binary back and re-start
  * [?] `evolution`: self-update, pull binary and re-start itself
* fs: file operations
  * [ ] `fs`: top op code
    * `get`: read file to stream
    * `push`: save stream to file
    * `del`: delete file/directory
    * `call`: execute file
    * `mv`: rename/move
    * `mkdir`: make directory
* task: for tasks/jobs/schedules
  * [ ] `task`: top op code
  * [ ] `add`:
  * [ ] `del`:
  * [ ] `start`:
  * [ ] `stop`:
  * [ ] `ls`:

### build

only enable basic op: `go build -ldflags='-s -w' -tags="" bot.go`

enable `mod` and `extra` op: `go build -ldflags='-s -w' -tags="extra mod" bot.go`

enable all op: `go build -ldflags='-s -w' -tags="all" bot.go`



## hub server

A server run with public IP that can be connected.

### functions to add

* [ ] auto pull bot info
* [ ] push binary for update old version bot
* [ ] push tasks/woks for bot to run
* [ ] statistics for IP, uptime, bandwidth...etc.

### key

* RSA Priv * n for bot
* ECDSA Pub * n to check authorized admin

### build

```
go build hub.go share.go
```


## admin tool

A simple CLI tool to operate bots via hub.

### commands

* bot
	* ls [id | addr | time | rtt] : list all bot on hub by ...
	* kill
	* reconn
* local 
	* ls : list local side server
	* bind $bot_id $bind_addr $mode $argv... : bind server (eg. socks5) on local side
	* stop $bind_addr


### key

* RSA Pub * 1 for connect to hub, same as bot
* ECDSA Priv * 1 for hub to check authorized


### build

```
go build admin.go share.go
```

## `proxys.go`
[WIP] socks5 proxy server for auto switch between bots use round-robin.
With web API/UI for user operate, plan to replace admin tool.

### planning features

* [x] socks5 proxy server with auto switch between all bots in a hub
* [ ] select single/multiple bots for proxy auto switch
* [ ] multiple hub connection
* [x] list all bots on hub
* [ ] select single/multiple bots for operate


## other tools

* `genkeys.go` : RSA & ECDSA keys generator


