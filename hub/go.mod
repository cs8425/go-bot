module hub

go 1.16

replace lib/chacha20 => ../lib/chacha20

replace lib/fakehttp => ../lib/fakehttp

replace lib/smux => ../lib/smux

replace lib/godaemon => ../lib/godaemon

replace local/base => ../lib/base

replace local/streamcoder => ../lib/streamcoder

replace local/toolkit => ../lib/toolkit

replace local/log => ../lib/log

require (
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	lib/fakehttp v0.0.0-00010101000000-000000000000 // indirect
	local/base v0.0.0-00010101000000-000000000000 // indirect
	local/log v0.0.0-00010101000000-000000000000 // indirect
)
