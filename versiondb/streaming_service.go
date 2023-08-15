package versiondb

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/public-awesome/stargaze/v11/versiondb/tsrocksdb"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

var _ baseapp.StreamingService = &StreamingService{}

// StreamingService is a concrete implementation of StreamingService that accumulate the state changes in current block,
// writes the ordered changeset out to version storage.
type StreamingService struct {
	listeners          []*types.MemoryListener // the listeners that will be initialized with BaseApp
	VersionStore       VersionStore
	CurrentBlockNumber int64 // the current block number
}

// NewFileStreamingService is the streaming.ServiceConstructor function for
// creating a FileStreamingService.
func NewVersionDbStreamingService(
	homePath string,
	keys []storetypes.StoreKey,
	marshaller codec.BinaryCodec,
) (*StreamingService, error) {
	dataDir := filepath.Join(homePath, "data", "versiondb")
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		return nil, err
	}
	versionDB, err := tsrocksdb.NewStore(dataDir)
	if err != nil {
		return nil, err
	}

	// default to exposing all
	exposeStoreKeys := make([]storetypes.StoreKey, 0, len(keys))
	for _, storeKey := range keys {
		exposeStoreKeys = append(exposeStoreKeys, storeKey)
	}

	service := NewStreamingService(versionDB, exposeStoreKeys)

	return service, nil
}

// NewStreamingService creates a new StreamingService for the provided writeDir, (optional) filePrefix, and storeKeys
func NewStreamingService(versionStore VersionStore, storeKeys []types.StoreKey) *StreamingService {
	// sort by the storeKeys first
	sort.SliceStable(storeKeys, func(i, j int) bool {
		return strings.Compare(storeKeys[i].Name(), storeKeys[j].Name()) < 0
	})

	listeners := make([]*types.MemoryListener, len(storeKeys))
	for i, key := range storeKeys {
		listeners[i] = types.NewMemoryListener(key)
	}
	return &StreamingService{listeners, versionStore, 0}
}

// Listeners satisfies the baseapp.StreamingService interface
func (fss *StreamingService) Listeners() map[types.StoreKey][]types.WriteListener {
	listeners := make(map[types.StoreKey][]types.WriteListener, len(fss.listeners))
	for _, listener := range fss.listeners {
		listeners[listener.StoreKey()] = []types.WriteListener{listener}
	}
	return listeners
}

// ListenBeginBlock satisfies the baseapp.ABCIListener interface
// It sets the currentBlockNumber.
func (fss *StreamingService) ListenBeginBlock(ctx context.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error {
	fss.CurrentBlockNumber = req.GetHeader().Height
	return nil
}

// ListenDeliverTx satisfies the baseapp.ABCIListener interface
func (fss *StreamingService) ListenDeliverTx(ctx context.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error {
	return nil
}

// ListenEndBlock satisfies the baseapp.ABCIListener interface
// It merge the state caches of all the listeners together, and write out to the versionStore.
func (fss *StreamingService) ListenEndBlock(ctx context.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error {
	return nil
}

func (fss *StreamingService) ListenCommit(ctx context.Context, res abci.ResponseCommit) error {
	// concat the state caches
	var changeSet []types.StoreKVPair
	for _, listener := range fss.listeners {
		changeSet = append(changeSet, listener.PopStateCache()...)
	}

	return fss.VersionStore.PutAtVersion(fss.CurrentBlockNumber, changeSet)
}

// Stream satisfies the baseapp.StreamingService interface
func (fss *StreamingService) Stream(wg *sync.WaitGroup) error {
	return nil
}

// Close satisfies the io.Closer interface, which satisfies the baseapp.StreamingService interface
func (fss *StreamingService) Close() error {
	return nil
}