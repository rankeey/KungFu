package plan

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

// PeerList is an ordered list of PeerIDs
type PeerList []PeerID

func (pl PeerList) String() string {
	var parts []string
	for _, p := range pl {
		parts = append(parts, p.String())
	}
	return strings.Join(parts, ",")
}

func (pl PeerList) DebugString() string {
	return fmt.Sprintf("[%d]{%s}", len(pl), pl)
}

func (pl PeerList) Bytes() []byte {
	b := &bytes.Buffer{}
	for _, p := range pl {
		binary.Write(b, binary.LittleEndian, &p)
	}
	return b.Bytes()
}

func (pl PeerList) Clone() PeerList {
	ql := make(PeerList, len(pl))
	copy(ql, pl)
	return ql
}

func (pl PeerList) Rank(q PeerID) (int, bool) {
	for i, p := range pl {
		if p == q {
			return i, true
		}
	}
	return -1, false
}

func (pl PeerList) LocalRank(q PeerID) (int, bool) {
	var i int
	for _, p := range pl {
		if p == q {
			return i, true
		}
		if p.ColocatedWith(q) {
			i++
		}
	}
	return -1, false
}

func (pl PeerList) Select(ranks []int) PeerList {
	var ql PeerList
	for _, rank := range ranks {
		ql = append(ql, pl[rank])
	}
	return ql
}

func (pl PeerList) Contains(p PeerID) bool {
	_, ok := pl.Rank(p)
	return ok
}

func (pl PeerList) Set() map[PeerID]struct{} {
	s := make(map[PeerID]struct{})
	for _, p := range pl {
		s[p] = struct{}{}
	}
	return s
}

func (pl PeerList) sub(ql PeerList) PeerList {
	s := ql.Set()
	var a PeerList
	for _, p := range pl {
		if _, ok := s[p]; !ok {
			a = append(a, p)
		}
	}
	return a
}

func (pl PeerList) Intersection(ql PeerList) PeerList {
	s := ql.Set()
	var a PeerList
	for _, p := range pl {
		if _, ok := s[p]; ok {
			a = append(a, p)
		}
	}
	return a
}

func (pl PeerList) Disjoint(ql PeerList) bool {
	return len(pl.Intersection(ql)) == 0
}

func (pl PeerList) Diff(ql PeerList) (PeerList, PeerList) {
	return pl.sub(ql), ql.sub(pl)
}

func (pl PeerList) Eq(ql PeerList) bool {
	if len(pl) != len(ql) {
		return false
	}
	for i, p := range pl {
		if p != ql[i] {
			return false
		}
	}
	return true
}

func (pl PeerList) On(host uint32) PeerList {
	var ql PeerList
	for _, p := range pl {
		if p.IPv4 == host {
			ql = append(ql, p)
		}
	}
	return ql
}

func ParsePeerList(val string) (PeerList, error) {
	parts := strings.Split(val, ",")
	var pl PeerList
	for _, p := range parts {
		id, err := ParsePeerID(p)
		if err != nil {
			return nil, err
		}
		pl = append(pl, *id)
	}
	return pl, nil
}
