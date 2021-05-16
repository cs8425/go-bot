module base

go 1.16

replace lib/smux => ../smux

replace lib/godaemon => ../godaemon

replace lib/chacha20 => ../chacha20

replace local/streamcoder => ../streamcoder

replace local/toolkit => ../toolkit

require (
	lib/smux v0.0.0-00010101000000-000000000000 // indirect
	lib/godaemon v0.0.0-00010101000000-000000000000 // indirect
	lib/chacha20 v0.0.0-00010101000000-000000000000 // indirect
	local/streamcoder v0.0.0-00010101000000-000000000000 // indirect
	local/toolkit v0.0.0-00010101000000-000000000000 // indirect
)

