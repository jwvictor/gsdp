/* Copyright (C) 2017 Jason Vitor

 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the Modified BSD License.

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.

 * You should have received a copy of the Modified BSD License
 * along with this program.  If not, see
 * <https://opensource.org/licenses/BSD-3-Clause>
 */

package gsdp

import (
	"errors"
	"log"
	"google.golang.org/grpc"
	"sync"
	"time"
)

const (
	CONN_CLOSED = iota
	CONN_USED   = iota
	CONN_OPEN   = iota
)

type ConnPoolStatus int

type OpenConnection struct {
	conn     *grpc.ClientConn
	domain   string
	state    ConnPoolStatus
	lastUsed int64
}

type ConnectionPool struct {
	connExists map[string]bool
	connPtrs   map[string]*OpenConnection
	connRecvrs map[string]chan *OpenConnection
	lock       *sync.Mutex
	reapFreq   time.Duration
}

func NewConnectionPool(reapF time.Duration) *ConnectionPool {
	ce := make(map[string]bool)
	cp := make(map[string]*OpenConnection)
	cr := make(map[string]chan *OpenConnection)
	m := &sync.Mutex{}
	p := &ConnectionPool{ce, cp, cr, m, reapF}
	return p
}

func (p *ConnectionPool) Start() {
	go p.ReapForPool()
}

func (c *ConnectionPool) makeConnectionForDomain(domain string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(domain+":"+default_port, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func makeNewOpenConnection(conn *grpc.ClientConn, domain string) *OpenConnection {
	return &OpenConnection{conn, domain, CONN_OPEN, time.Now().Unix()}
}

func (p *ConnectionPool) GetConnection(domain string) (*OpenConnection, error) {
	p.lock.Lock()
	if b, ok := p.connExists[domain]; !(ok && b) {
		// NOTE: in theory we could allow multiple connections by making this a counter
		p.connExists[domain] = true
		if cc, err := p.makeConnectionForDomain(domain); err != nil {
			return nil, err
		} else {
			log.Printf("Making new connection for domain %s\n", domain)
			oc := makeNewOpenConnection(cc, domain)
			p.connPtrs[domain] = oc
			oc.state = CONN_USED
			ch := make(chan *OpenConnection, 1) // Buffered channel of size 1
			p.connRecvrs[domain] = ch
			ch <- oc
		}
	}
	ch := p.connRecvrs[domain]
	p.lock.Unlock() // Only lock on whether connections need to be created
	log.Printf("Unlock for connection for %s, waiting...\n", domain)
	x := <-ch // Blocking wait
	x.lastUsed = time.Now().Unix()
	return x, nil
}

func (p *ConnectionPool) ReleaseConnection(conn *OpenConnection) error {
	if ch, ok := p.connRecvrs[conn.domain]; ok {
		ch <- conn // Return to pool
		return nil
	} else {
		return errors.New("No such domain in pool -- did you return this to the wrong pool?")
	}
}

func (p *ConnectionPool) ReapNotThreadSafe() error {
	for dom, oc := range p.connPtrs {
		if (time.Now().Unix() - oc.lastUsed) >= 10 {
			log.Printf("More than 10 seconds elapsed for connection to %s...\n", dom)
			p.ReapConnectionNotThreadSafe(dom)
		}
	}
	return nil
}

func (p *ConnectionPool) ReapConnectionNotThreadSafe(domain string) error {
	ch := p.connRecvrs[domain]
	c := <-ch
	p.connExists[domain] = false
	close(ch)
	delete(p.connRecvrs, domain)
	delete(p.connPtrs, domain)
	c.conn.Close()
	c.state = CONN_CLOSED
	log.Printf("CLOSED connection to domain %s\n", c.domain)
	return nil
}

func (p *ConnectionPool) ReapForPool() error {
	ticker := time.NewTicker(p.reapFreq)
	for t := range ticker.C {
		log.Printf("Reaping connections at %v\n", t)
		p.lock.Lock()
		p.ReapNotThreadSafe()
		p.lock.Unlock()
	}
	return nil
}
