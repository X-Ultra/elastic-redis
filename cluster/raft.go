package cluster

import (
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	connectionPollPeriod = 10 * time.Second

	//snapshots history retained num.
	retainSnapshotCount  = 2
	applyTimeout         = 10 * time.Second
	openTimeout          = 120 * time.Second
	sqliteFile           = "raft.db"
	leaderWaitDelay      = 100 * time.Millisecond
	appliedWaitDelay     = 100 * time.Millisecond
	connectionPoolCount  = 5
	connectionTimeout    = 10 * time.Second

	// raftLogCacheSize is the maximum number of logs to cache in-memory.
	// This is used to reduce disk I/O for the recently committed entries.
	raftLogCacheSize     = 512
)

type Server struct {
	//server config, raft ,serf
	config *Config

	Ln Listener

	// The raft instance is used among Consul nodes within the DC to protect
	// operations that require strong consistency.
	// the state directly.
	raft          *raft.Raft
	raftStore     *raftboltdb.BoltStore
	raftTransport *raft.NetworkTransport
	raftInmem     *raft.InmemStore

	// raftNotifyCh is set up by setupRaft() and ensures that we get reliable leader
	// transition notifications from the Raft layer.
	raftNotifyCh <-chan bool

}

func NewServer(lnr Listener, config *Config) (*Server, error) {

	//Create server.
	s := &Server{
		Ln:     lnr,
		config: config,
	}

	if err := s.setupRaft(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) setupRaft() error {
	logger := log.New(os.Stdout, "[raft-server] ", log.LstdFlags)

	// If we have an unclean exit then attempt to close the Raft store.
	defer func() {
		if s.raft == nil && s.raftStore != nil {
			if err := s.raftStore.Close(); err != nil {
				logger.Printf("[ERR] consul: failed to close Raft store: %v", err)
			}
		}
	}()

	// Build an all in-memory setup for dev mode, otherwise prepare a full
	// disk-based setup.
	var logStore raft.LogStore
	var stable raft.StableStore
	var snap raft.SnapshotStore
	var fsm raft.FSM

	// init raft config
	raftConf := raft.DefaultConfig()


	//Dev Mode , in memory
	if s.config.DevMode {
		store := raft.NewInmemStore()
		s.raftInmem = store
		stable = store
		logStore = store
		snap = raft.NewInmemSnapshotStore()
		logger.Println("[INFO] RaftConfig.ProtocolVersion: ", raftConf.ProtocolVersion)
	} else {
		// Create the FSM
		fsm = &simpleFsm{}

		// Create the backend raft store for logs and stable storage.
		store, err := raftboltdb.NewBoltStore(filepath.Join(s.config.DataPath, sqliteFile))
		if err != nil {
			return err
		}
		s.raftStore = store
		stable = store

		// Create the raft log cache, wrap the store in a LogCache to improve performance.
		logCacheStore, err := raft.NewLogCache(raftLogCacheSize, store)
		if err != nil {
			return err
		}
		logStore = logCacheStore

		//Create the snapshot store
		raftSnapLogOutput, err := os.OpenFile(s.config.LogsPath+"/snapshots.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		snapshot, err := raft.NewFileSnapshotStore(s.config.DataPath, retainSnapshotCount, raftSnapLogOutput)
		if err != nil {
			return err
		}
		snap = snapshot

	}

	//Create the TCPTransport
	// Create a transport layer.
	tpLogOutput, tplerr := os.OpenFile(s.config.LogsPath+"/transport.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if tplerr != nil {
		return tplerr
	}
	trans := raft.NewNetworkTransport(
		NewTransportDelegate(s.Ln),
		connectionPoolCount,
		connectionTimeout,
		tpLogOutput)
	s.raftTransport=trans

	// Versions of the Raft protocol below 3 require the LocalID to match the network
	// address of the transport.
	raftConf.LocalID = raft.ServerID(trans.LocalAddr())


	// If we are in bootstrap or dev mode and the state is clean then we can
	// bootstrap now.
	if s.config.Bootstrap || s.config.DevMode {
		hasState, err := raft.HasExistingState(logStore, stable, snap)
		if err != nil {
			return err
		}
		if !hasState {
			configuration := raft.Configuration{
				Servers: []raft.Server{
					raft.Server{
						ID:      raftConf.LocalID,
						Address: trans.LocalAddr(),
					},
				},
			}
			if err := raft.BootstrapCluster(raftConf,
				logStore, stable, snap, trans, configuration); err != nil {
				return err
			}
		}
	}

	// Setup the Raft store.
	var err error

	// Set up a channel for reliable leader notifications.
	raftNotifyCh := make(chan bool, 1)
	s.raftNotifyCh = raftNotifyCh

	s.raft, err = raft.NewRaft(raftConf, fsm, logStore, stable, snap, trans)
	if err != nil {
		return err
	}

	return nil
}
