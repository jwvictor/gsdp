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

package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/jwvictor/gsdp"
	pb "github.com/jwvictor/gsdprotocol"
	"os"
	"strings"
	"time"
)

type GsdpIdentConfig struct {
	IdentityPath      string `toml:"ident"`
	PubIdentitiesPath string `toml:"idents_path"`
}

type GsdpClientConfig struct {
	Identity GsdpIdentConfig `toml:"identity"`
}

func printUsage() {
	fmt.Printf("Usage: %s <cmd> *args\n", os.Args[0])
}

func main() {
	idPathHelp := "Identity path (without .priv or .ident)"
	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	serveIdentPath := serveCmd.String("id", "", idPathHelp)
	serveIdsPath := serveCmd.String("pubidpath", "", "Public identity path (directory)")

	newIdCmd := flag.NewFlagSet("newid", flag.ExitOnError)
	newIdName := newIdCmd.String("name", "", "Name for newly generated user")
	newIdHandle := newIdCmd.String("handle", "", "Handle for newly generated user")
	newIdDomain := newIdCmd.String("domain", "", "Domain for newly generated user")
	newIdProfileUrl := newIdCmd.String("profile", "", "Profile URL for newly generated user")
	newIdIdsPath := newIdCmd.String("pubidpath", "", "Public identity path (directory)")

	sayCmd := flag.NewFlagSet("say", flag.ExitOnError)
	sayIdentPath := sayCmd.String("id", "", idPathHelp)
	sayTextInput := sayCmd.String("text", "", "Text input")
	sayToStr := sayCmd.String("to", "", "Recipient list (semicolon-delimited)")
	sayIdsPath := sayCmd.String("pubidpath", "", "Public identity path (directory)")

	lsCmd := flag.NewFlagSet("ls", flag.ExitOnError)
	lsIdentPath := lsCmd.String("id", "", idPathHelp)
	lsIdsPath := lsCmd.String("pubidpath", "", "Public identity path (directory)")

	popCmd := flag.NewFlagSet("pop", flag.ExitOnError)
	popIdsPath := popCmd.String("pubidpath", "", "Public identity path (directory)")
	popIdentPath := popCmd.String("id", "", idPathHelp)

	idsPath := ""
	idPath := &idsPath
	idPath = nil
	allIdentPath := &idsPath
	allIdentPath = nil
	r, _ := time.ParseDuration("60000ms")  // TODO: make configurable
	connectionPool := gsdp.NewConnectionPool(r)
	connectionPool.Start()

	// Handle for config file
	defaultCfgFn := os.Getenv("HOME") + "/.gsdp.toml"
	cfgFn := &defaultCfgFn
	config := GsdpClientConfig{}
	if _, err := toml.DecodeFile(*cfgFn, &config); err == nil {
		if len(config.Identity.PubIdentitiesPath) > 0 {
			idPath = &config.Identity.PubIdentitiesPath
		}
		if len(config.Identity.IdentityPath) > 0 {
			allIdentPath = &config.Identity.IdentityPath
		}
	} else {
		panic(err)
	}

	switch os.Args[1] {
	case "serve":
		serveCmd.Parse(os.Args[2:])
		if idPath == nil || (len(*idPath) == 0) {
			idPath = serveIdsPath
		}
		if allIdentPath == nil || (len(*allIdentPath) == 0) {
			allIdentPath = serveIdentPath
		}
	case "newid":
		newIdCmd.Parse(os.Args[2:])
		if idPath == nil || (len(*idPath) == 0) {
			idPath = newIdIdsPath
		}
	case "say":
		sayCmd.Parse(os.Args[2:])
		if idPath == nil || (len(*idPath) == 0) {
			idPath = sayIdsPath
		}
		if allIdentPath == nil || (len(*allIdentPath) == 0) {
			allIdentPath = sayIdentPath
		}
	case "ls":
		lsCmd.Parse(os.Args[2:])
		if idPath == nil || (len(*idPath) == 0) {
			idPath = lsIdsPath
		}
		if allIdentPath == nil || (len(*allIdentPath) == 0) {
			allIdentPath = lsIdentPath
		}
	case "test":
		lsCmd.Parse(os.Args[2:])
		if idPath == nil || (len(*idPath) == 0) {
			idPath = lsIdsPath
		}
		if allIdentPath == nil || (len(*allIdentPath) == 0) {
			allIdentPath = lsIdentPath
		}
	case "pop":
		popCmd.Parse(os.Args[2:])
		if idPath == nil || (len(*idPath) == 0) {
			idPath = popIdsPath
		}
		if allIdentPath == nil || (len(*allIdentPath) == 0) {
			allIdentPath = popIdentPath
		}
	default:
		printUsage()
		os.Exit(2)
	}
	envPath := os.Getenv("GSDPIDPATH")
	if len(envPath) > 0 {
		idPath = &envPath
	}

	allIdentities := gsdp.MakeInMemoryIdentStoreFromFiles(*idPath)
	privIds := gsdp.MakeFilePrivateKeyStore(*idPath)
	switch os.Args[1] {
	case "serve":
		fmt.Printf("Serving...\n")
		ids := allIdentities
		gsdp.Serve(":50051", privIds, ids, connectionPool)
	case "newid":
		newid, privkey := gsdp.NewIdentity(*newIdName, *newIdHandle, *newIdDomain, *newIdProfileUrl)
		fnb := *idPath + "/" + *newIdHandle + "__" + *newIdDomain
		fmt.Printf("Handle: %s\n", *newIdHandle)
		e := gsdp.SaveIdentity(newid, privkey, fnb)
		if e == nil {
			fmt.Printf("Identity saved to: %s\n", fnb)
		} else {
			panic(e)
		}
	case "say":
		path := allIdentPath
		if sayToStr == nil {
			panic(errors.New("Need a TO for a SAY"))
		}
		envPath := os.Getenv("GSDPID")
		if len(envPath) > 0 {
			path = &envPath
		}
		id, privk, err := gsdp.LoadIdentity(*path)
		if err != nil {
			panic(err)
		}
		uu := gsdp.MakeLocalUser(id, privk)
		client := gsdp.NewClient(&uu, allIdentities, connectionPool)
		nothin := []byte{}
		txtBytes := []byte(*sayTextInput)
		toPcs := strings.Split(*sayToStr, "\\")
		var lookupDomain string
		var recipDomain string
		if len(toPcs) < 2 {
			lookupDomain = "gsdp.co"
		} else {
			lookupDomain = toPcs[1]
			recipDomain = toPcs[1]
		}
		idLookup := allIdentities.GetIdentityForHandleDomain(toPcs[0], lookupDomain)
		if len(toPcs) < 2 || idLookup == nil {
			fmt.Printf("Looking up %s at %s", toPcs[0], lookupDomain)
			res, e := client.Name(&pb.NameInquiry{id, nil, false, toPcs[0], lookupDomain})
			if e != nil {
				panic(e)
			}
			if res.IsError == true {
				panic(errors.New("Cannot find user at nameserver."))
			} else {
				fmt.Printf("Got new user identity: %s (%s)\n", toPcs, res.ProfileUrl)
				recipDomain = res.Name.Domain
				allIdentities.AddIdentity(res.Name)
			}
		}
		recipId, err := gsdp.LoadPublicIdentity(*idPath + "/" + toPcs[0] + "__" + recipDomain)
		if err != nil {
			panic(err)
		}
		recips := []*pb.Identity{recipId}
		rawm := &pb.RawMessage{id, recips, nothin, pb.MessageType_PLAIN, txtBytes, nothin, nothin, time.Now().Unix(), nothin, nothin}
		err = gsdp.DoRawMessageEncryption(txtBytes, recipId, rawm)
		if err != nil {
			panic(err)
		}
		err = client.Say(rawm)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Sent message (%d bytes) \n", len(rawm.MessageContent))
	case "pop":
		path := allIdentPath
		envPath := os.Getenv("GSDPID")
		if len(envPath) > 0 {
			path = &envPath
		}
		id, privk, err := gsdp.LoadIdentity(*path)
		if err != nil {
			panic(err)
		}
		uu := gsdp.MakeLocalUser(id, privk)
		client := gsdp.NewClient(&uu, allIdentities, connectionPool)
		msgs, err := client.GetMine(true)
		if err != nil {
			panic(err)
		}
		for i, m := range msgs {
			pt, err := gsdp.DoRawMessageDecryption(m, privk)
			if err != nil {
				fmt.Printf("Err: %v\n", err)
			}
			fmt.Printf("Msg %d: %s - %s\n", i, (m.FromIdent.Handle + "\\" + m.FromIdent.Domain), string(pt))
		}
		fmt.Printf("\n\n")
	case "ls":
		path := allIdentPath
		envPath := os.Getenv("GSDPID")
		if len(envPath) > 0 {
			path = &envPath
		}
		id, privk, err := gsdp.LoadIdentity(*path)
		if err != nil {
			panic(err)
		}
		uu := gsdp.MakeLocalUser(id, privk)
		client := gsdp.NewClient(&uu, allIdentities, connectionPool)
		msgs, err := client.GetMine(false)
		if err != nil {
			panic(err)
		}
		matrix := make([][]string, 0)
		for _, m := range msgs {
			pt, err := gsdp.DoRawMessageDecryption(m, privk)
			if err != nil {
				fmt.Printf("Err: %v\n", err)
			}
			matrix = append(matrix, []string{m.FromIdent.Handle  + "\\" + m.FromIdent.Domain, string(pt)})
		}
		PrintGrid(matrix)
		fmt.Printf("\n\n")
	case "test":
		path := allIdentPath
		envPath := os.Getenv("GSDPID")
		if len(envPath) > 0 {
			path = &envPath
		}
		id, privk, err := gsdp.LoadIdentity(*path)
		if err != nil {
			panic(err)
		}
		uu := gsdp.MakeLocalUser(id, privk)
		client := gsdp.NewClient(&uu, allIdentities, connectionPool)
		msgs, err := client.GetMine(false)
		if err != nil {
			panic(err)
		}
		matrix := make([][]string, 0)
		for _, m := range msgs {
			pt, err := gsdp.DoRawMessageDecryption(m, privk)
			if err != nil {
				fmt.Printf("Err: %v\n", err)
			}
			matrix = append(matrix, []string{m.FromIdent.Handle  + "\\" + m.FromIdent.Domain, string(pt)})
		}
		PrintGrid(matrix)
		fmt.Printf("\n\n")
		for i := 0; i < 15; i++ {
			time.Sleep(time.Millisecond * 1000)
		}
		fmt.Printf("Next request:")
		client = gsdp.NewClient(&uu, allIdentities, connectionPool)
		msgs, err = client.GetMine(false)
		if err != nil {
			panic(err)
		}
		matrix = make([][]string, 0)
		for _, m := range msgs {
			pt, err := gsdp.DoRawMessageDecryption(m, privk)
			if err != nil {
				fmt.Printf("Err: %v\n", err)
			}
			matrix = append(matrix, []string{m.FromIdent.Handle  + "\\" + m.FromIdent.Domain, string(pt)})
		}
		PrintGrid(matrix)
		fmt.Printf("\n\n")
	}
}
