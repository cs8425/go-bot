package base

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"encoding/hex"
	"strings"
	"sort"
	"fmt"

	kit "../toolkit"
	"../smux"
)

const (
	maxInPool = 16384
	maxInfoMem = 4096
)

type Pool struct {
	lock    sync.RWMutex
	nextId  int32
	m2P     map[int32]*Peer // for ref Pear
	m2ID    map[int32]string // for UUID	// TODO: change type as []byte
	m2IPP   map[int32]string // for IP:port
}

func NewPool() (*Pool) {
	return &Pool {
		m2P: make(map[int32]*Peer),
		m2ID: make(map[int32]string),
		m2IPP: make(map[int32]string),
	}
}

func (p *Pool) AddPear(peer *Peer) (int32, bool) {
	p.lock.Lock()
	if len(p.m2P) > maxInPool {
		p.lock.Unlock()
		return int32(-1), false
	}
	id := p.nextId

	peer.id = id
	p.m2P[id] = peer
	p.m2ID[id] = string(peer.UUID)
	p.m2IPP[id] = peer.Conn.RemoteAddr().String()

	p.nextId++
	if p.nextId < 0 {
		p.nextId = 0
	}

	p.lock.Unlock()
	return id, true
}

func (p *Pool) DelPear(id int32) {
	p.lock.Lock()

	delete(p.m2P, id)
	delete(p.m2ID, id)
	delete(p.m2IPP, id)

	p.lock.Unlock()
}

func (p *Pool) getUUIDList(UUID []byte) (list []int32) {
	cmpID := string(UUID)
	list = make([]int32, 0)
	p.lock.RLock()
	for id, uid := range p.m2ID {
		if uid == cmpID {
			_, ok := p.m2P[id]
			if !ok {
				continue
			}
			list = append(list, id)
		}
	}
	p.lock.RUnlock()
	return list
}

func (p *Pool) CheckOld(UUID []byte, addr string) ([]*Peer) {
	cmpAddr := strings.Split(addr, ":")
	if len(cmpAddr) < 2{
		return nil
	}

	list := p.getUUIDList(UUID)
	out := make([]*Peer, 0)

	for _, id := range list {
		v, ok := p.m2IPP[id]
		if !ok {
			continue
		}
		t := strings.Split(v, ":")
		if t[0] == cmpAddr[0] {
			oldpeer, ok := p.GetByID(id)
			if ok {
				out = append(out, oldpeer)
			}
		}
	}
	return out
}

func (p *Pool) GetByID(id int32) (*Peer, bool) {
	p.lock.RLock()
	data, ok := p.m2P[id]
	p.lock.RUnlock()
	return data, ok
}

func (p *Pool) GetByUTag(UTag string) (data *Peer, ok bool) {
	arg := strings.Split(UTag, "/")
	if len(arg) < 2{
		return nil , false
	}
	p.lock.RLock()
	defer p.lock.RUnlock()

	ok = false
	maybeID := make([]int32, 0)
	for idx, addr := range p.m2IPP {
		if addr == arg[1] {
			maybeID = append(maybeID, idx)
		}
	}

	decoded, _ := hex.DecodeString(arg[0])
	cmpID := string(decoded)
	for _, id := range maybeID {
		UUID, _ := p.m2ID[id]
		if UUID == cmpID {
			data, ok = p.m2P[id]
			break
		}
	}

	return data, ok
}

func (p *Pool) getList() ([]*PeerInfo) {
	peers := make([]*PeerInfo, 0, len(p.m2P))
	for _, v := range p.m2P {
		addr := v.Conn.RemoteAddr().String()
		p := &PeerInfo{
			UUID: v.UUID,
			Addr: addr,
			UTag: fmt.Sprintf("%v/%v", kit.Hex(v.UUID), addr),
			UpTime: v.UpTime,
			RTT: v.Mux.GetRTT(),
		}
		peers = append(peers, p)
	}
	return peers
}

func (p *Pool) getListStr(peers []*Peer) ([]string) {
	out := make([]string, 0, len(peers))
	now := time.Now()
	for _, peer := range peers {
		UUID := peer.UUID
		Addr := peer.Conn.RemoteAddr().String()
		t := peer.UpTime.Format(time.RFC3339)
		t2 := now.Sub(peer.UpTime).String()
		tag := fmt.Sprintf("%v/%v [%v](%v)", kit.Hex([]byte(UUID)), Addr, t, t2)
		out = append(out, tag)
	}
	return out
}

func (p *Pool) GetListByID() ([]*PeerInfo) {
	p.lock.RLock()

	peers := p.getList()
	sort.Sort(ByID(peers))

	p.lock.RUnlock()
	return peers
}

func (p *Pool) GetListByAddr() ([]*PeerInfo) {
	p.lock.RLock()

	peers := p.getList()
	sort.Sort(ByAddr(peers))

	p.lock.RUnlock()
	return peers
}

func (p *Pool) GetListByTime() ([]*PeerInfo) {
	p.lock.RLock()

	peers := p.getList()
	sort.Sort(ByTime(peers))

	p.lock.RUnlock()
	return peers
}

func (p *Pool) GetListByRTT() ([]*PeerInfo) {
	p.lock.RLock()

	peers := p.getList()
	sort.Sort(ByRTT(peers))

	p.lock.RUnlock()
	return peers
}

func (p *Pool) Clear() {
	p.lock.Lock()
	p.m2P = make(map[int32]*Peer)
	p.m2ID = make(map[int32]string)
	p.m2IPP = make(map[int32]string)
	p.lock.Unlock()
}

func (p *Pool) WriteListTo(conn net.Conn) (int) {
	p.lock.RLock()
	peers := p.getList()
	p.lock.RUnlock()

	v := PeerList(peers)
	return v.WriteTo(conn)
}

// hub only
type Peer struct {
	id int32
	UUID []byte
	Conn net.Conn
	Mux *smux.Session
	UpTime time.Time
	Info *Info
}

// over hub and admin
type PeerInfo struct {
	UUID []byte
	Addr string
	UTag string
	UpTime time.Time
	RTT time.Duration
}

func (p *PeerInfo) String() (string) {
	now := time.Now()
	t := p.UpTime.Format(time.RFC3339)
	t2 := now.Sub(p.UpTime).String()
	return fmt.Sprintf("%v [%v](%v) %v", p.UTag, t, t2, p.RTT)
}

type PeerList []*PeerInfo
func (p PeerList) WriteTo(conn net.Conn) (int) {
	list := []*PeerInfo(p)
	n := len(list)
	kit.WriteVLen(conn, int64(n))
	for _, v := range list {
		kit.WriteVTagByte(conn, []byte(v.UTag))
		kit.WriteVLen(conn, int64(v.RTT))
		uptime, _ := v.UpTime.MarshalBinary()
		kit.WriteTagByte(conn, uptime)
	}
	return n
}
func (p *PeerList) ReadFrom(conn net.Conn) (int, error) {
	n, err := kit.ReadVLen(conn)
	if err != nil {
		return 0, err
	}

	list := make([]*PeerInfo, 0, n)
	for i := 0; i < int(n); i++ {
		utag, err := kit.ReadVTagByte(conn)
		if err != nil {
			break
		}

		rtt, err := kit.ReadVLen(conn)
		if err != nil {
			break
		}

		tbuf, err := kit.ReadTagByte(conn)
		if err != nil {
			break
		}
		uptime := time.Time{}
		err = uptime.UnmarshalBinary(tbuf)
		if err != nil {
			continue
		}

		v := &PeerInfo{
			UTag: string(utag),
			RTT: time.Duration(rtt),
			UpTime: uptime,
		}

		list = append(list, v)
	}

	*p = PeerList(list)

	return int(n), err
}

func (p *PeerList) GetListByID() ([]*PeerInfo) {
	sort.Sort(ByID([]*PeerInfo(*p)))
	return []*PeerInfo(*p)
}

func (p *PeerList) GetListByAddr() ([]*PeerInfo) {
	sort.Sort(ByAddr([]*PeerInfo(*p)))
	return []*PeerInfo(*p)
}

func (p *PeerList) GetListByTime() ([]*PeerInfo) {
	sort.Sort(ByTime([]*PeerInfo(*p)))
	return []*PeerInfo(*p)
}

func (p *PeerList) GetListByRTT() ([]*PeerInfo) {
	sort.Sort(ByRTT([]*PeerInfo(*p)))
	return []*PeerInfo(*p)
}


type ByID []*PeerInfo
func (s ByID) Len() int {
	return len(s)
}
func (s ByID) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByID) Less(i, j int) bool {
	return s[i].UTag < s[j].UTag
}

type ByAddr []*PeerInfo
func (s ByAddr) Len() int {
    return len(s)
}
func (s ByAddr) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}
func (s ByAddr) Less(i, j int) bool {
    return s[i].Addr < s[j].Addr
}

type ByTime []*PeerInfo
func (s ByTime) Len() int {
    return len(s)
}
func (s ByTime) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}
func (s ByTime) Less(i, j int) bool {
    return s[i].UpTime.Before(s[j].UpTime) // oldest first
}

type ByRTT []*PeerInfo
func (s ByRTT) Len() int {
    return len(s)
}
func (s ByRTT) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}
func (s ByRTT) Less(i, j int) bool {
    return s[i].RTT < s[j].RTT
}

func NewPeer(p1 net.Conn, mux *smux.Session, UUID []byte) *Peer {
	return &Peer{
		UUID: UUID,
		Conn: p1,
		Mux: mux,
		UpTime: time.Now(),
		Info: NewInfo(),
	}
}

type Info struct {
	size int32
	data map[string]string
	lock sync.RWMutex
}

func NewInfo() *Info {
	return &Info{
		data: make(map[string]string),
	}
}

func (inf *Info) Get(key string) (string, bool) {
	inf.lock.RLock()
	data, ok := inf.data[key]
	inf.lock.RUnlock()
	return data, ok
}

func (inf *Info) Set(key string, value string) (bool) {
	inf.lock.Lock()
	newsize := int(inf.size) + len(key) + len(value)
	if newsize > maxInfoMem {
		inf.lock.Unlock()
		return false
	}
	inf.size = int32(newsize)
	inf.data[key] = value
	inf.lock.Unlock()
	return true
}

func (inf *Info) Del(key string) {
	inf.lock.Lock()
	data, ok := inf.data[key]
	if ok {
		inf.size -= int32(len(key) + len(data))
	}
	delete(inf.data, key)
	inf.lock.Unlock()
	return
}

func (inf *Info) Mem() int {
	return int(atomic.LoadInt32(&inf.size))
}

func (inf *Info) Clear() {
	inf.lock.Lock()
	inf.data = make(map[string]string)
	inf.lock.Unlock()
}

func (inf *Info) WriteTo(conn net.Conn) (int) {
	inf.lock.RLock()

	n := len(inf.data)
	kit.WriteVLen(conn, int64(n))
	for k, v := range inf.data {
		kit.WriteVTagByte(conn, []byte(k))	// TODO: key length limit!!
		kit.WriteVTagByte(conn, []byte(v))
	}

	inf.lock.RUnlock()
	return n
}

func (inf *Info) ReadFrom(conn net.Conn) (int, error) {

	n, err := kit.ReadVLen(conn)
	if err != nil {
		return 0, err
	}

	for i := 0; i < int(n); i++ {
		k, err := kit.ReadVTagByte(conn)	// TODO: key length limit!!
		if err != nil {
			break
		}
		v, err := kit.ReadVTagByte(conn)
		if err != nil {
			break
		}

		inf.Set(string(k), string(v))
	}

	return int(n), err
}

