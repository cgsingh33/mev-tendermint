package mempool

import (
	"github.com/tendermint/tendermint/p2p"
)

// TxInfo are parameters that get passed when attempting to add a tx to the
// mempool.
type TxInfo struct {
	// SenderID is the internal peer ID used in the mempool to identify the
	// sender, storing two bytes with each transaction instead of 20 bytes for
	// the types.NodeID.
	SenderID uint16

	// SenderP2PID is the actual p2p.ID of the sender, used e.g. for logging.
	SenderP2PID p2p.ID

	// ordering for sidecar tx
	BundleId int64

	// auction height desired for tx
	DesiredHeight int64

	// order desired within bundle (i.e. per BundleID)
	BundleOrder int64

	// total size of bundle
	BundleSize int64
}
