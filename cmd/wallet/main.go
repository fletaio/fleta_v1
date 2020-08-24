package main

import (
	"encoding/hex"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/fletaio/fleta_v1/service/bank"

	"github.com/fletaio/fleta_v1/cmd/app"
	"github.com/fletaio/fleta_v1/cmd/closer"
	"github.com/fletaio/fleta_v1/cmd/config"
	"github.com/fletaio/fleta_v1/common"
	"github.com/fletaio/fleta_v1/common/key"
	"github.com/fletaio/fleta_v1/common/rlog"
	"github.com/fletaio/fleta_v1/core/backend"
	_ "github.com/fletaio/fleta_v1/core/backend/buntdb_driver"
	"github.com/fletaio/fleta_v1/core/chain"
	"github.com/fletaio/fleta_v1/core/pile"
	"github.com/fletaio/fleta_v1/core/types"
	"github.com/fletaio/fleta_v1/pof"
	"github.com/fletaio/fleta_v1/process/admin"
	"github.com/fletaio/fleta_v1/process/formulator"
	"github.com/fletaio/fleta_v1/process/gateway"
	"github.com/fletaio/fleta_v1/process/payment"
	"github.com/fletaio/fleta_v1/process/vault"
	"github.com/fletaio/fleta_v1/service/apiserver"
	"github.com/fletaio/fleta_v1/service/p2p"
)

// Config is a configuration for the cmd
type Config struct {
	SeedNodeMap  map[string]string
	NodeKeyHex   string
	ObserverKeys []string
	Port         int
	APIPort      int
	StoreRoot    string
	RLogHost     string
	RLogPath     string
	UseRLog      bool
}

func main() {
	var cfg Config
	if err := config.LoadFile("./config.toml", &cfg); err != nil {
		panic(err)
	}
	if len(cfg.StoreRoot) == 0 {
		cfg.StoreRoot = "./ndata"
	}
	if len(cfg.RLogHost) > 0 && cfg.UseRLog {
		if len(cfg.RLogPath) == 0 {
			cfg.RLogPath = "./ndata_rlog"
		}
		rlog.SetRLogHost(cfg.RLogHost)
		rlog.Enablelogger(cfg.RLogPath)
	}

	var ndkey key.Key
	if len(cfg.NodeKeyHex) > 0 {
		if bs, err := hex.DecodeString(cfg.NodeKeyHex); err != nil {
			panic(err)
		} else if Key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
			panic(err)
		} else {
			ndkey = Key
		}
	} else {
		if bs, err := ioutil.ReadFile("./ndkey.key"); err != nil {
			k, err := key.NewMemoryKey()
			if err != nil {
				panic(err)
			}

			fs, err := os.Create("./ndkey.key")
			if err != nil {
				panic(err)
			}
			fs.Write(k.Bytes())
			fs.Close()
			ndkey = k
		} else {
			if Key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
				panic(err)
			} else {
				ndkey = Key
			}
		}
	}

	ObserverKeys := []common.PublicHash{}
	for _, k := range cfg.ObserverKeys {
		pubhash, err := common.ParsePublicHash(k)
		if err != nil {
			panic(err)
		}
		ObserverKeys = append(ObserverKeys, pubhash)
	}
	SeedNodeMap := map[common.PublicHash]string{}
	for k, netAddr := range cfg.SeedNodeMap {
		pubhash, err := common.ParsePublicHash(k)
		if err != nil {
			panic(err)
		}
		SeedNodeMap[pubhash] = netAddr
	}

	cm := closer.NewManager()
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		cm.CloseAll()
	}()
	defer cm.CloseAll()

	MaxBlocksPerFormulator := uint32(10)
	ChainID := uint8(0x01)
	Symbol := "FLETA"
	Usage := "Mainnet"
	Version := uint16(0x0001)

	back, err := backend.Create("buntdb", cfg.StoreRoot+"/context")
	if err != nil {
		panic(err)
	}
	cdb, err := pile.Open(cfg.StoreRoot + "/chain")
	if err != nil {
		panic(err)
	}
	cdb.SetSyncMode(true)
	st, err := chain.NewStore(back, cdb, ChainID, Symbol, Usage, Version)
	if err != nil {
		panic(err)
	}
	cm.Add("store", st)

	if st.Height() > 0 {
		if _, err := cdb.GetData(st.Height(), 0); err != nil {
			panic(err)
		}
	}

	cs := pof.NewConsensus(MaxBlocksPerFormulator, ObserverKeys)
	app := app.NewFletaApp()
	cn := chain.NewChain(cs, app, st)
	cn.MustAddProcess(admin.NewAdmin(1))
	vp := vault.NewVault(2)
	cn.MustAddProcess(vp)
	fp := formulator.NewFormulator(3)
	cn.MustAddProcess(fp)
	cn.MustAddProcess(gateway.NewGateway(4))
	cn.MustAddProcess(payment.NewPayment(5))
	as := apiserver.NewAPIServer()
	cn.MustAddService(as)
	keyStore, err := backend.Create("buntdb", cfg.StoreRoot+"/keystore")
	if err != nil {
		panic(err)
	}
	bp := bank.NewBank(keyStore, cfg.StoreRoot+"/bank")
	cn.MustAddService(bp)
	if err := cn.Init(); err != nil {
		panic(err)
	}
	if err := bp.InitFromStore(st); err != nil {
		panic(err)
	}
	cm.RemoveAll()
	cm.Add("chain", cn)

	if err := st.IterBlockAfterContext(func(b *types.Block) error {
		if cm.IsClosed() {
			return chain.ErrStoreClosed
		}
		if err := cn.ConnectBlock(b, nil); err != nil {
			return err
		}
		log.Println(b.Header.Height, "Connect block From local", b.Header.Generator.String(), b.Header.Height)
		return nil
	}); err != nil {
		if err == chain.ErrStoreClosed {
			return
		}
		panic(err)
	}

	nd := p2p.NewNode(ndkey, SeedNodeMap, cn, cfg.StoreRoot+"/peer")
	if err := nd.Init(); err != nil {
		panic(err)
	}
	bp.SetNode(nd)
	cm.RemoveAll()
	cm.Add("node", nd)

	go nd.Run(":" + strconv.Itoa(cfg.Port))
	go as.Run(":" + strconv.Itoa(cfg.APIPort))

	cm.Wait()
}
