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
	pb "github.com/jwvictor/gsdprotocol"
	"io/ioutil"
	"strings"
	"sync"
)

type IdentityStore interface {
	GetIdentityForHandleDomain(string, string) *pb.Identity
	AddIdentity(*pb.Identity) error
}

type PrivateKeyStore interface {
	GetKeyFor([]byte) *[]byte
}

type InMemoryIdentStore struct {
	Idents  []*pb.Identity
	SrcPath string
	lck     *sync.Mutex
}

type FilePrivateKeyStore struct {
	Idents []LocalUser
	lck    *sync.Mutex
}

type SingleKeyStore struct {
	Ident      []byte
	PrivateKey []byte
}

func MakeFilePrivateKeyStore(path string) *FilePrivateKeyStore {
	ids := make([]LocalUser, 0)
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		fn := f.Name()
		if strings.HasSuffix(fn, ".priv") {
			pfn := path + "/" + fn[:len(fn)-5]
			pkid, pk, err := LoadIdentity(pfn)
			if err == nil {
				ids = append(ids, LocalUser{pkid, pk})
			}
		}
	}
	m := &sync.Mutex{}
	return &FilePrivateKeyStore{ids, m}
}

func SameBytes(x []byte, y []byte) bool {
	for i, _ := range x {
		if i >= len(y) || x[i] != y[i] {
			return false
		}
	}
	return true
}

func (s *SingleKeyStore) GetKeyFor(bs []byte) *[]byte {
	if SameBytes(s.Ident, bs) {
		z := s.PrivateKey
		return &z
	} else {
		return nil
	}
}

func MakeInMemoryIdentStoreFromFiles(path string) *InMemoryIdentStore {
	ids := make([]*pb.Identity, 0)
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		fn := f.Name()
		if strings.HasSuffix(fn, ".ident") {
			pfn := path + "/" + fn[:len(fn)-6]
			newid, err := LoadPublicIdentity(pfn)
			if err == nil {
				ids = append(ids, newid)
			}
		}
	}
	m := &sync.Mutex{}
	return &InMemoryIdentStore{ids, path, m}
}

func (s *InMemoryIdentStore) AddIdentity(id *pb.Identity) error {
	s.lck.Lock()
	s.Idents = append(s.Idents, id)
	idPath := s.SrcPath + "/" + id.Handle + "__" + id.Domain
	SaveIdentity(id, nil, idPath)
	s.lck.Unlock()
	return nil
}

func (s *FilePrivateKeyStore) GetKeyFor(bs []byte) *[]byte {
	s.lck.Lock()
	for _, v := range s.Idents {
		if SameBytes(v.Ident(), bs) {
			s.lck.Unlock()
			q := v.PrivKey()
			return &q
		}
	}
	s.lck.Unlock()
	return nil
}

func (s *InMemoryIdentStore) GetIdentityForHandleDomain(handle string, domain string) *pb.Identity {
	s.lck.Lock()
	for _, v := range s.Idents {
		if v.Handle == handle && strings.ToLower(v.Domain) == strings.ToLower(domain) {
			s.lck.Unlock()
			return v
		}
	}
	s.lck.Unlock()
	return nil
}
