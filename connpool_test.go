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
	"fmt"
	"testing"
	"time"
)

func getMockPool() *ConnectionPool {
	td, _ := time.ParseDuration("10000ms")
	p := NewConnectionPool(td)
	oc := &OpenConnection{nil, "test.com", CONN_CLOSED, time.Now().Unix()}
	p.connPtrs["test.com"] = oc
	p.connExists["test.com"] = true
	ch := make(chan *OpenConnection, 1) // Buffered channel of size 1
	p.connRecvrs["test.com"] = ch
	ch <- oc // And send it into the buffer
	return p
}

func TestPoolGetsConnection(t *testing.T) {
	p := getMockPool()
	var myc *OpenConnection
	go func() {
		theoc, _ := p.GetConnection("test.com")
		myc = theoc
	}()
	// Get it back out
	for i := 0; i < 5 && myc == nil; i++ {
		time.Sleep(time.Millisecond * 1000)
	}
	if myc == nil {
		t.Error("Couldn't get connection out")
	}
	if myc.domain != "test.com" {
		t.Error("Wrong domain")
	}
}

func TestPoolReleaseLogic(t *testing.T) {
	p := getMockPool()
	var myc *OpenConnection
	go func() {
		theoc, _ := p.GetConnection("test.com")
		fmt.Printf("GetConnection returned.\n")
		myc = theoc
	}()
	// Get it back out
	for i := 0; i < 5 && myc == nil; i++ {
		time.Sleep(time.Millisecond * 1000)
	}
	if myc == nil {
		t.Error("Couldn't get connection out first time")
	}
	var myc2 *OpenConnection
	go func() {
		theoc, _ := p.GetConnection("test.com")
		fmt.Printf("GetConnection returned.\n")
		myc2 = theoc
	}()
	if myc2 != nil {
		t.Error("Got locked resource")
	}
	p.ReleaseConnection(myc)
	// Now the previous call should return and we should be able to get the resource lock
	for i := 0; i < 15 && myc2 == nil; i++ {
		time.Sleep(time.Millisecond * 1000)
	}
	if myc2 == nil {
		t.Error("Couldn't get connection out second time")
	}
}
