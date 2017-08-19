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
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	pb "github.com/jwvictor/gsdprotocol"
	"testing"
	"time"
)

func makeAnIdentity() (*pb.Identity, []byte, error) {
	rsrc := rand.Reader
	name := "testname"
	domain := "testname"
	profileUrl := "facebook.com/test"
	privk, err := rsa.GenerateKey(rsrc, 2048)
	if err != nil {
		return nil, nil, err
	}
	pubk := privk.Public()
	pubkrsa := pubk.(*rsa.PublicKey)
	pubkbs := PubKeyToBytes(pubkrsa)
	privkbs := PrivKeyToBytes(privk)
	hashd := BytesToIdentHash(pubkbs)
	id := &pb.Identity{hashd[:], name, name, domain, pubkbs, profileUrl}
	return id, privkbs, nil
}

func TestMakeIdentity(t *testing.T) {
	id, privk, uerr := makeAnIdentity()
	if uerr != nil {
		t.Error(fmt.Sprintf("Got an error creating user: %v", uerr))
	}
	if len(privk) == 0 {
		t.Error("Bad private key")
	}
	if len(id.Ident) == 0 {
		t.Error("Bad ident hash")
	}
}

func TestEncryptMessage(t *testing.T) {
	id, privk, uerr := makeAnIdentity()
	if uerr != nil {
		t.Error(fmt.Sprintf("Got an error creating user: %v", uerr))
	}
	txtBytes := []byte("hi there")
	nothin := []byte{}
	recips := []*pb.Identity{id}
	rawm := &pb.RawMessage{id, recips, nothin, pb.MessageType_PLAIN, txtBytes, nothin, nothin, time.Now().Unix(), nothin, nothin}
	err := DoRawMessageEncryption(txtBytes, id, rawm)
	if err != nil {
		t.Error(fmt.Sprintf("Got an error from encryption: %v", err))
	}
	pt, err := DoRawMessageDecryption(rawm, privk)
	if err != nil {
		t.Error(fmt.Sprintf("Got an error from decryption: %v", err))
	}
	decStr := string(pt)
	fmt.Printf("Decrypted %x into %s\n", pt, decStr)
	if decStr != "hi there" {
		t.Error("Decryption failed to return correct value.")
	}
}
