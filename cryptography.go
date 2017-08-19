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
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/golang/protobuf/proto"
	pb "github.com/jwvictor/gsdprotocol"
	"io"
	"os"
)

func PubKeyToBytes(pk *rsa.PublicKey) []byte {
	z, err := x509.MarshalPKIXPublicKey(pk)
	if err != nil {
		return nil
	}
	return z
}

func PrivKeyToBytes(pk *rsa.PrivateKey) []byte {
	z := x509.MarshalPKCS1PrivateKey(pk)
	if z == nil {
		return nil
	}
	return z
}

func BytesToPubKey(bs []byte) *rsa.PublicKey {
	k, err := x509.ParsePKIXPublicKey(bs)
	if err != nil {
		return nil
	}
	return k.(*rsa.PublicKey)
}

func BytesToPrivKey(bs []byte) *rsa.PrivateKey {
	k, err := x509.ParsePKCS1PrivateKey(bs)
	if err != nil {
		return nil
	}
	return k
}

func WriteBytes(fn string, bs []byte) error {
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	f.Write(bs)
	f.Sync()
	return nil
}

func SaveIdentity(id *pb.Identity, privKey []byte, path string) error {
	idBytes, err := proto.Marshal(id)
	if err == nil {
		s1 := path + ".ident"
		s2 := path + ".priv"
		WriteBytes(s1, idBytes)
		if privKey != nil {
			WriteBytes(s2, privKey)
		}
	} else {
		fmt.Printf("Error saving: %v\n", err)
		return err
	}
	return nil
}

func LoadBytes(f *os.File) []byte {
	res := []byte{}
	var buf [256]byte
	for n_read := 1; n_read > 0; {
		n, err := f.Read(buf[:])
		n_read = n
		if err != nil {
			n_read = 0 //panic(err)
		}
		if n_read > 0 {
			res = append(res, buf[:n_read]...)
		}
	}
	return res
}

func IdentToString(bs []byte) string {
	return base64.StdEncoding.EncodeToString(bs)
}

func DoRawMessageDecryption(msg *pb.RawMessage, userPrivk []byte) ([]byte, error) {
	rng := rand.Reader
	ciphertextSym := msg.SymKey
	ciphertext := msg.MessageContent
	usrKey := BytesToPrivKey(userPrivk)
	key := make([]byte, 32)
	if _, err := io.ReadFull(rng, key); err != nil {
		return nil, err
	}
	if err := rsa.DecryptPKCS1v15SessionKey(rng, usrKey, ciphertextSym, key); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plaintext, err := aead.Open(nil, msg.Nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func DoRawMessageEncryption(rawMsg []byte, recip *pb.Identity, msg *pb.RawMessage) error {
	pubk := BytesToPubKey(recip.PubKey)
	rng := rand.Reader
	key := make([]byte, 32)
	if _, err := io.ReadFull(rng, key); err != nil {
		panic(err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rng, nonce); err != nil {
		panic(err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}
	ciphertext := aesgcm.Seal(nil, nonce, rawMsg, nil)
	msg.MessageContent = ciphertext
	msg.Nonce = nonce
	encryptedSymKey, err := rsa.EncryptPKCS1v15(rng, pubk, key)
	if err != nil {
		panic(err)
	}
	msg.SymKey = encryptedSymKey
	return nil
}

func LoadPublicIdentity(path string) (*pb.Identity, error) {
	idPath := path + ".ident"
	fid, err1 := os.Open(idPath)
	if err1 != nil {
		return nil, err1
	}
	defer fid.Close()
	barr := LoadBytes(fid)
	ident := &pb.Identity{}
	perr := proto.Unmarshal(barr, ident)
	if perr != nil {
		return nil, perr
	}
	return ident, nil
}

func LoadIdentity(path string) (*pb.Identity, []byte, error) {
	pkPath := path + ".priv"
	fpk, err2 := os.Open(pkPath)
	if err2 != nil {
		return nil, nil, err2
	}
	defer fpk.Close()
	ident, ierr := LoadPublicIdentity(path)
	if ierr != nil {
		return nil, nil, ierr
	}
	privk := LoadBytes(fpk)
	return ident, privk, nil
}

func BytesToIdentHash(pubkbs []byte) []byte {
	hashd := sha256.Sum256(pubkbs)
	return hashd[:]
}

func NewIdentity(name string, handle string, domain string, profileUrl string) (*pb.Identity, []byte) {
	rsrc := rand.Reader
	privk, err := rsa.GenerateKey(rsrc, 2048)
	if err != nil {
		return nil, nil
	}
	pubk := privk.Public()
	pubkrsa := pubk.(*rsa.PublicKey)
	pubkbs := PubKeyToBytes(pubkrsa)
	privkbs := PrivKeyToBytes(privk)
	hashd := BytesToIdentHash(pubkbs)
	id := pb.Identity{hashd[:], handle, name, domain, pubkbs, profileUrl}
	return &id, privkbs
}
