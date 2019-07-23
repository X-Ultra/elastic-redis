package main

import (
	"flag"
	"fmt"
	"github.com/x-ultra/elastic-redis/cluster"
	"github.com/x-ultra/elastic-redis/libs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
)

const name = `elastic-redis`
const desc = `elastic-redis is a lightweight, distributed relational database, which uses SQLite as its
storage engine. It provides an easy-to-use, fault-tolerant store for relational data.`

var (
	version = "0.1"
)

var dataPath string
var logsPath string
var devMode bool
var httpAddr string
var httpAdv string
var httpIdleTimeout string
var httpTxTimeout string
var nodeEncrypt bool
var nodeID string
var raftAddr string
var raftAdv string
var joinAddr string
var noVerify bool
var noNodeVerify bool
var raftSnapThreshold uint64
var raftHeartbeatTimeout string
var raftApplyTimeout string
var raftOpenTimeout string
var showVersion bool
var cpuProfile string
var memProfile string

func init() {
	flag.StringVar(&dataPath, "data-path", "data", "data path")
	flag.StringVar(&logsPath, "logs-path", "logs", "logs path")
	flag.BoolVar(&devMode, "dev-mode", false, "use dev mode")
	flag.StringVar(&nodeID, "node-id", "", "Unique name for node. If not set, set to hostname")
	flag.StringVar(&httpAddr, "http-addr", "localhost:5001", "HTTP server bind address. For HTTPS, set X.509 cert and key")
	flag.StringVar(&httpAdv, "http-adv-addr", "", "Advertised HTTP address. If not set, same as HTTP server")
	flag.StringVar(&httpIdleTimeout, "http-conn-idle-timeout", "60s", "HTTP connection idle timeout. Use 0s for no timeout")
	flag.StringVar(&httpTxTimeout, "http-conn-tx-timeout", "10s", "HTTP transaction timeout. Use 0s for no timeout")
	flag.BoolVar(&noVerify, "http-no-verify", false, "Skip verification of remote HTTPS cert when joining cluster")
	flag.BoolVar(&nodeEncrypt, "node-encrypt", false, "Enable node-to-node encryption")
	flag.BoolVar(&noNodeVerify, "node-no-verify", false, "Skip verification of a remote node cert")
	flag.StringVar(&raftAddr, "raft-addr", "localhost:5000", "Raft communication bind address")
	flag.StringVar(&raftAdv, "raft-adv-addr", "", "Advertised Raft communication address. If not set, same as Raft bind")
	flag.StringVar(&joinAddr, "join", "", "Comma-delimited list of nodes, through which a cluster can be joined (proto://host:port)")
	flag.BoolVar(&showVersion, "version", false, "Show version information and exit")
	flag.StringVar(&raftHeartbeatTimeout, "raft-timeout", "1s", "Raft heartbeat timeout")
	flag.StringVar(&raftApplyTimeout, "raft-apply-timeout", "10s", "Raft apply timeout")
	flag.StringVar(&raftOpenTimeout, "raft-open-timeout", "120s", "Time for initial Raft logs to be applied. Use 0s duration to skip wait")
	flag.Uint64Var(&raftSnapThreshold, "raft-snap", 8192, "Number of outstanding log entries that trigger snapshot")
	flag.StringVar(&cpuProfile, "cpu-profile", "", "Path to file for CPU profiling information")
	flag.StringVar(&memProfile, "mem-profile", "", "Path to file for memory profiling information")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\n%s\n\n", desc)
		fmt.Fprintf(os.Stderr, "Usage: %s [arguments] <data directory>\n", name)
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	// Configure logging and pump out initial message.
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stderr)
	log.SetPrefix(fmt.Sprintf("[%s] ", name))
	log.Printf("%s starting, version %s", name, version)
	log.Printf("%s, target architecture is %s, operating system target is %s", runtime.Version(), runtime.GOARCH, runtime.GOOS)

	if showVersion {
		fmt.Printf("name:%s, version:%s, os:%s, arch:%s, go version:%s\n", name, version, runtime.GOOS, runtime.GOARCH, runtime.Version())
		os.Exit(0)
	}

	os.Exit(realMain())
}

func realMain() int {

	conf := &cluster.Config{
		DataPath: dataPath,
		LogsPath: logsPath,
		DevMode:  devMode,
	}
	conf.Init()

	// Create inter node network layer.
	var tp *libs.Transport
	tp = libs.NewTransport()
	if err := tp.Open(raftAddr); err != nil {
		log.Fatalf("failed to open internode network layer: %s", err.Error())
	}

	//Start raft cluster
	conf.Bootstrap = true
	server, err := cluster.NewServer(tp, conf)
	if err != nil {
		log.Fatalln(err)
	}

	// Start the HTTP API server.
	if err := startHTTPService(server); err != nil {
		log.Fatalf("failed to start HTTP server: %s", err.Error())
	}

	// Block until signalled.
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Println("rqlite server stopped")

	return 1

}

func startHTTPService(s *cluster.Server) error {
	httpServer := http.Server{
		Addr: httpAddr,
		Handler: &cluster.RootHandler{s},
	}

	go func() {
		err := httpServer.ListenAndServe()
		if err != nil {
			fmt.Println("HTTP service Serve() returned:", err.Error())
		}
	}()
	fmt.Println("http api service listening on", l.Addr() )

	return nil
}


