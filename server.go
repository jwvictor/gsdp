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
	"log"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"sync"
)

const (
	port = ":50051"
)

type GSDPServer struct {
	localUsers    map[string]LocalUser
	ongoingBlocks map[string][]*pb.RawMessage
	knownUsers      IdentityStore
	updateBuffer    map[string][]*pb.RawMessage
	clients         map[string]*GSDPClient
	privateKeyStore PrivateKeyStore
	bufMutex        *sync.Mutex
	connectionPool  *ConnectionPool
}

func (s *GSDPServer) Sup(ctx context.Context, in *pb.RequestPermissions) (*pb.UserPermissions, error) {
	// Immediately accept and return good permissions.
	s.knownUsers.AddIdentity(in.RequestedPermissions.Ident)
	newperms := &pb.UserPermissions{in.ToIdent, "promiscuous", 100, true, true, true, true, true, true}
	toid := s.knownUsers.GetIdentityForHandleDomain(in.ToIdent.Handle, in.ToIdent.Domain)
	if toid != nil {
		client, err2 := s.getClient(toid.Ident)
		if err2 != nil {
			return nil, errors.New("Cannot contact server regarding permissions")
		}
		// Immediately respond ok
		client.SendK(in.RequestedPermissions)
	} else {
		return nil, errors.New("Unknown user")
	}
	return newperms, nil
}

func (s *GSDPServer) tryGetPermissions(me *pb.Identity, handle string, domain string) *pb.Identity {
	return nil
}

func (s *GSDPServer) getClient(ident []byte) (*GSDPClient, error) {
	if _, ok := s.clients[IdentToString(ident)]; !ok {
		if lu, okk := s.localUsers[IdentToString(ident)]; okk {
			cc := NewClient(&lu, s.knownUsers, s.connectionPool)
			c := &cc
			s.clients[IdentToString(ident)] = c
		}
	}
	return s.clients[IdentToString(ident)], nil
}

func (s *GSDPServer) GetMine(ctx context.Context, in *pb.GetRequest) (*pb.PendingData, error) {
	id := IdentToString(in.FromIdent.Ident)
	if _, ok := s.localUsers[id]; !ok {
		fmt.Printf("Never heard of %s -- adding user.\n", id)
		pk := s.privateKeyStore.GetKeyFor(in.FromIdent.Ident)
		if pk != nil {
			s.localUsers[id] = LocalUser{in.FromIdent, *pk}
		} else {
			fmt.Printf("No such local users, assuming remote client. This needs to be authenticated.\n") // TODO: AUTH
		}
	}
	msgs := make([]*pb.RawMessage, 0)
	s.bufMutex.Lock()
	if bmsgs, ok := s.updateBuffer[id]; ok {
		msgs = bmsgs
		if in.Purge {
			s.updateBuffer[id] = make([]*pb.RawMessage, 0)
		}
	}
	s.bufMutex.Unlock()
	return &pb.PendingData{in.FromIdent, in.SinceUtc, msgs}, nil
}

func (s *GSDPServer) LeaveBlock(ctx context.Context, in *pb.BlockLeaveRequest) (*pb.BlockStatusChangeResponse, error) {
	return &pb.BlockStatusChangeResponse{true, ""}, nil
}

func (s *GSDPServer) StartBlock(ctx context.Context, in *pb.BlockStartRequest) (*pb.BlockStatusChangeResponse, error) {
	return &pb.BlockStatusChangeResponse{true, ""}, nil
}

func (s *GSDPServer) K(ctx context.Context, in *pb.ApprovePermissions) (*pb.UserPermissions, error) {
	bs := []byte{1, 2, 3}
	fmt.Printf("Got approval! %v\n", in)
	return &pb.UserPermissions{&pb.Identity{bs, "@test", " ", "", bs, ""}, "custom", 4, true, false, true, true, true, true}, nil
}

func (s *GSDPServer) Btw(ctx context.Context, in *pb.ApprovePermissions) (*pb.MessageAck, error) {
	return &pb.MessageAck{true, "big prob"}, nil
}

func (s *GSDPServer) deliverTo(id *pb.Identity, in *pb.RawMessage) (*pb.MessageAck, error) {
	idk := IdentToString(id.Ident)
	s.bufMutex.Lock()
	buf, ok := s.updateBuffer[idk]
	if ok {
		s.updateBuffer[idk] = append(buf, in)
	} else {
		s.updateBuffer[idk] = []*pb.RawMessage{in}
	}
	s.bufMutex.Unlock()
	return &pb.MessageAck{false, ""}, nil
}

func (s *GSDPServer) Say(ctx context.Context, in *pb.RawMessage) (*pb.MessageAck, error) {
	for _, r := range in.ToIdent {
		toid := s.knownUsers.GetIdentityForHandleDomain(r.Handle, r.Domain)
		ok := (toid != nil)
		if !ok {
			// TODO: need to get such identity from name server and
			// add it to the local store
			//perms := s.tryGetPermissions(in.FromIdent, r.Handle, r.Domain)
			return &pb.MessageAck{true, "could not resolve user"}, nil
		}
		s.deliverTo(toid, in)
	}
	return &pb.MessageAck{false, "OK"}, nil
}

func (s *GSDPServer) Name(ctx context.Context, in *pb.NameInquiry) (*pb.NameResponse, error) {
	fmt.Printf("Looking up %s in %v \n", in.RequestHandle, s.localUsers)
	if u, ok := s.localUsers[in.RequestHandle]; ok {
		return &pb.NameResponse{false, u.identity, ""}, nil
	} else if theid := s.knownUsers.GetIdentityForHandleDomain(in.RequestHandle, in.RequestDomain); theid != nil {
		return &pb.NameResponse{false, theid, ""}, nil
	} else {
		return &pb.NameResponse{true, nil, ""}, nil
	}
}

func (s *GSDPServer) Initialize(pks PrivateKeyStore, idStore IdentityStore, connPool *ConnectionPool) error {
	s.localUsers = make(map[string]LocalUser)
	s.knownUsers = idStore
	s.updateBuffer = make(map[string][]*pb.RawMessage)
	s.bufMutex = &sync.Mutex{}
	s.connectionPool = connPool
	s.privateKeyStore = pks
	return nil
}

func Serve(port string, pks PrivateKeyStore, idStore IdentityStore, cp *ConnectionPool) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	gs := GSDPServer{}
	gs.Initialize(pks, idStore, cp)
	pb.RegisterGSDPServer(s, &gs)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

