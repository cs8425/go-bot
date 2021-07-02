package base

import (
	"net"
	// "lib/smux"
)

// Event ID
const (
	EV_admin_conn       = "adm_conn"
	EV_admin_conn_cls   = "adm_conn_cls"
	EV_admin_stream     = "adm_stream"     // new stream
	EV_admin_stream_cls = "adm_stream_cls" // stream close
)

// return net.Conn for bandwidth counting / throughput limit ...etc
// return error for auth check
type EventCallbackFunc func(ev_type string, mainConn net.Conn, streamConn net.Conn, argv ...interface{}) (warpConn net.Conn, err error)
