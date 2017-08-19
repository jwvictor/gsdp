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
	"golang.org/x/net/context"
	"log"
)

const (
	default_keyspace = "gsdprotocol"
	default_port     = "50051"
)

type LocalUser struct {
	identity *pb.Identity
	privKey  []byte
}

type GSDPClient struct {
	user       LocalUser
	identities IdentityStore
	connPool   *ConnectionPool
}

func (l LocalUser) PrivKey() []byte {
	return l.privKey
}

func (l LocalUser) Ident() []byte {
	return l.identity.Ident
}

func NewClient(lident *LocalUser, identities IdentityStore, connPool *ConnectionPool) GSDPClient {
	return GSDPClient{*lident, identities, connPool}
}

func MakeLocalUser(id *pb.Identity, pk []byte) LocalUser {
	return LocalUser{id, pk}
}

func (c *GSDPClient) getConnectionByDomain(domain string) (*OpenConnection, error) {
	oc, e := c.connPool.GetConnection(domain)
	if e != nil {
		return nil, e
	}
	return oc, nil
}

func (c *GSDPClient) getConnection(id *pb.Identity) (*OpenConnection, error) {
	conn, e := c.getConnectionByDomain(id.Domain)
	return conn, e
}

func (c *GSDPClient) GetMine(purge bool) ([]*pb.RawMessage, error) {
	oconn, err := c.getConnection(c.user.identity)
	if err != nil {
		return nil, err
	}
	conn := oconn.conn
	defer c.connPool.ReleaseConnection(oconn)
	client := pb.NewGSDPClient(conn)
	getReq := &pb.GetRequest{c.user.identity, 0, purge, 0, []byte{}}
	pending, err := client.GetMine(context.Background(), getReq)
	if err != nil {
		panic(err)
	}
	lst := make([]*pb.RawMessage, 0)
	for _, m := range pending.Messages {
		lst = append(lst, m)
	}
	return lst, nil
}

func (c *GSDPClient) RequestPermissionsFrom(from *pb.Identity, perms *pb.UserPermissions) error {
	return nil
}

func (c *GSDPClient) SendK(perms *pb.UserPermissions) error {
	oconn, err := c.getConnection(perms.Ident)
	defer c.connPool.ReleaseConnection(oconn)
	if err != nil {
		return err
	}
	conn := oconn.conn
	client := pb.NewGSDPClient(conn)
	approval := &pb.ApprovePermissions{c.user.identity, perms, []byte{}}
	client.K(context.Background(), approval)
	return nil
}

func (c *GSDPClient) Say(msg *pb.RawMessage) error {
	oconn, err := c.getConnection(msg.ToIdent[0]) // TODO: send to ALL!!!!
	defer c.connPool.ReleaseConnection(oconn)
	if err != nil {
		return err
	}
	conn := oconn.conn
	client := pb.NewGSDPClient(conn)
	client.Say(context.Background(), msg)
	return nil
}

func (c *GSDPClient) GetPendingPermissions() ([]*pb.UserPermissions, error) {
	return nil, nil
}

func (c *GSDPClient) Name(req *pb.NameInquiry) (*pb.NameResponse, error) {
	//h := req.RequestHandle
	d := req.RequestDomain
	oconn, err := c.getConnectionByDomain(d)
	defer c.connPool.ReleaseConnection(oconn)
	if err != nil {
		return nil, err
	}
	conn := oconn.conn
	client := pb.NewGSDPClient(conn)
	a, b := client.Name(context.Background(), req)
	log.Printf("%v %v", a, b)
	return a, b
}
