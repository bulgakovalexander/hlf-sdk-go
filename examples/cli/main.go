package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"github.com/s7techlab/cckit/extensions/encryption"

	"log"

	"github.com/hyperledger/fabric/common/util"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/client"
	_ "github.com/s7techlab/hlf-sdk-go/crypto/ecdsa"
	_ "github.com/s7techlab/hlf-sdk-go/discovery/local"
	"github.com/s7techlab/hlf-sdk-go/identity"
	"go.uber.org/zap"
)

var ctx = context.Background()

var (
	mspId      = flag.String(`mspId`, ``, `MspId`)
	mspPath    = flag.String(`mspPath`, ``, `path to admin certificate`)
	configPath = flag.String(`configPath`, ``, `path to configuration file`)

	channel       = flag.String(`channel`, ``, `channel name`)
	cc            = flag.String(`cc`, ``, `chaincode name`)
	ccPath        = flag.String(`ccPath`, ``, `chaincode path`)
	ccVersion     = flag.String(`ccVersion`, ``, `chaincode version`)
	ccPolicy      = flag.String(`ccPolicy`, ``, `chaincode endorsement policy`)
	ccArgs        = flag.String(`ccArgs`, ``, `chaincode instantiation arguments`)
	ccTransient   = flag.String(`ccTransient`, ``, `chaincode transient arguments`)
	ccInstall     = flag.Bool(`ccInstall`, true, `install chaincode`)
	ccInstantiate = flag.Bool(`ccInstantiate`, true, `instantiate chaincode`)
)

func main() {
	id, err := identity.NewMSPIdentityFromPath(*mspId, *mspPath)

	if err != nil {
		log.Fatalln(`Failed to load identity:`, err)
	}

	l, _ := zap.NewDevelopment()

	core, err := client.NewCore(*mspId, id, client.WithConfigYaml(*configPath), client.WithLogger(l))
	if err != nil {
		log.Fatalln(`unable to initialize core:`, err)
	}

	if *ccInstall {
		if err = core.Chaincode(*cc).Install(ctx, *ccPath, *ccVersion); err != nil {
			log.Fatalln(err)
		}
	}

	if *ccInstantiate {
		transArgs := prepareTransArgs(*ccTransient)
		arg := *ccArgs
		chaincodeArgs := util.ToChaincodeArgs(arg)
		key := transArgs["ENCODE_KEY"]
		if key != nil {
			for i, arg := range chaincodeArgs {
				encryptBytes, err := encryption.EncryptBytes(key, arg)
				if err != nil {
					log.Fatalln(`unable to encrypt:`, arg, err)
				} else {
					chaincodeArgs[i] = encryptBytes
				}
			}
		}

		if err = core.Chaincode(*cc).Instantiate(
			ctx,
			*channel,
			*ccPath,
			*ccVersion,
			*ccPolicy,
			chaincodeArgs,
			transArgs,
		); err != nil {
			log.Fatalln(err)
		}
	}

	log.Println(`successfully initiated`)
}

func prepareTransArgs(args string) api.TransArgs {
	var t map[string]string
	var err error
	if err = json.Unmarshal([]byte(args), &t); err != nil {
		panic(err)
	}

	tt := api.TransArgs{}

	for k, v := range t {
		if tt[k], err = base64.StdEncoding.DecodeString(v); err != nil {
			panic(err)
		}
	}
	return tt
}

func init() {
	flag.Parse()
}
