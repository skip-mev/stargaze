package versiondb

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	types1 "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	interfaceRegistry            = codecTypes.NewInterfaceRegistry()
	testMarshaller               = codec.NewProtoCodec(interfaceRegistry)
	testStreamingService         *StreamingService
	testListener1, testListener2 types.WriteListener
	emptyContext                 = sdk.NewContext(nil, types1.Header{}, false, nil)
	emptyContextWrap             = sdk.WrapSDKContext(emptyContext)

	// test abci message types
	mockHash          = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	testBeginBlockReq = abci.RequestBeginBlock{
		Header: types1.Header{
			Height: 1,
		},
		ByzantineValidators: []abci.Evidence{},
		Hash:                mockHash,
		LastCommitInfo: abci.LastCommitInfo{
			Round: 1,
			Votes: []abci.VoteInfo{},
		},
	}
	testBeginBlockRes = abci.ResponseBeginBlock{
		Events: []abci.Event{
			{
				Type: "testEventType1",
			},
			{
				Type: "testEventType2",
			},
		},
	}
	testEndBlockReq = abci.RequestEndBlock{
		Height: 1,
	}
	testEndBlockRes = abci.ResponseEndBlock{
		Events:                []abci.Event{},
		ConsensusParamUpdates: &abci.ConsensusParams{},
		ValidatorUpdates:      []abci.ValidatorUpdate{},
	}
	testCommitRes = abci.ResponseCommit{
		Data:         []byte{1},
		RetainHeight: 0,
	}
	mockTxBytes1      = []byte{9, 8, 7, 6, 5, 4, 3, 2, 1}
	testDeliverTxReq1 = abci.RequestDeliverTx{
		Tx: mockTxBytes1,
	}
	mockTxBytes2      = []byte{8, 7, 6, 5, 4, 3, 2}
	testDeliverTxReq2 = abci.RequestDeliverTx{
		Tx: mockTxBytes2,
	}
	mockTxResponseData1 = []byte{1, 3, 5, 7, 9}
	testDeliverTxRes1   = abci.ResponseDeliverTx{
		Events:    []abci.Event{},
		Code:      1,
		Codespace: "mockCodeSpace",
		Data:      mockTxResponseData1,
		GasUsed:   2,
		GasWanted: 3,
		Info:      "mockInfo",
		Log:       "mockLog",
	}
	mockTxResponseData2 = []byte{1, 3, 5, 7, 9}
	testDeliverTxRes2   = abci.ResponseDeliverTx{
		Events:    []abci.Event{},
		Code:      1,
		Codespace: "mockCodeSpace",
		Data:      mockTxResponseData2,
		GasUsed:   2,
		GasWanted: 3,
		Info:      "mockInfo",
		Log:       "mockLog",
	}

	// mock store keys
	mockStoreKey1 = sdk.NewKVStoreKey("mockStore1")
	mockStoreKey2 = sdk.NewKVStoreKey("mockStore2")

	// file stuff
	testPrefix = "testPrefix"
	dbDir      string

	// mock state changes
	mockKey1   = []byte{1, 2, 3}
	mockValue1 = []byte{3, 2, 1}
	mockKey2   = []byte{2, 3, 4}
	mockValue2 = []byte{4, 3, 2}
	mockKey3   = []byte{3, 4, 5}
	mockValue3 = []byte{5, 4, 3}
)

func TestVersionDbStreamingService(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping TestFileStreamingService in CI environment")
	}

	testKeys := []types.StoreKey{mockStoreKey1, mockStoreKey2}
	var err error

	dbDir := t.TempDir()
	testStreamingService, err = NewVersionDbStreamingService(dbDir, testKeys, testMarshaller)

	require.Nil(t, err)
	require.IsType(t, &StreamingService{}, testStreamingService)

	listeners := testStreamingService.Listeners()
	require.Equal(t, len(testKeys), len(listeners))

	testListener1 = listeners[mockStoreKey1][0]
	testListener2 = listeners[mockStoreKey2][0]

	wg := new(sync.WaitGroup)

	testStreamingService.Stream(wg)
	testListenBlock(t)
	testStreamingService.Close()
	wg.Wait()
}

func testListenBlock(t *testing.T) {
	var (
		expectKVPairsStore1 [][]byte
		expectKVPairsStore2 [][]byte
	)

	// write state changes
	testListener1.OnWrite(mockStoreKey1, mockKey1, mockValue1, false)
	testListener2.OnWrite(mockStoreKey2, mockKey2, mockValue2, false)
	testListener1.OnWrite(mockStoreKey1, mockKey3, mockValue3, false)

	// expected KV pairs
	expectedKVPair1, err := testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey1.Name(),
		Key:      mockKey1,
		Value:    mockValue1,
		Delete:   false,
	})
	require.Nil(t, err)

	expectedKVPair2, err := testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey2.Name(),
		Key:      mockKey2,
		Value:    mockValue2,
		Delete:   false,
	})
	require.Nil(t, err)

	expectedKVPair3, err := testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey1.Name(),
		Key:      mockKey3,
		Value:    mockValue3,
		Delete:   false,
	})
	require.Nil(t, err)

	expectKVPairsStore1 = append(expectKVPairsStore1, expectedKVPair1, expectedKVPair3)
	expectKVPairsStore2 = append(expectKVPairsStore2, expectedKVPair2)

	// send the ABCI messages
	err = testStreamingService.ListenBeginBlock(emptyContextWrap, testBeginBlockReq, testBeginBlockRes)
	require.Nil(t, err)

	// write state changes
	testListener1.OnWrite(mockStoreKey1, mockKey1, mockValue1, false)
	testListener2.OnWrite(mockStoreKey2, mockKey2, mockValue2, false)
	testListener2.OnWrite(mockStoreKey2, mockKey3, mockValue3, false)

	// expected KV pairs
	expectedKVPair1, err = testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey1.Name(),
		Key:      mockKey1,
		Value:    mockValue1,
		Delete:   false,
	})
	require.Nil(t, err)

	expectedKVPair2, err = testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey2.Name(),
		Key:      mockKey2,
		Value:    mockValue2,
		Delete:   false,
	})
	require.Nil(t, err)

	expectedKVPair3, err = testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey2.Name(),
		Key:      mockKey3,
		Value:    mockValue3,
		Delete:   false,
	})
	require.Nil(t, err)

	expectKVPairsStore1 = append(expectKVPairsStore1, expectedKVPair1)
	expectKVPairsStore2 = append(expectKVPairsStore2, expectedKVPair2, expectedKVPair3)

	// send the ABCI messages
	err = testStreamingService.ListenDeliverTx(emptyContextWrap, testDeliverTxReq1, testDeliverTxRes1)
	require.Nil(t, err)

	// write state changes
	testListener2.OnWrite(mockStoreKey2, mockKey1, mockValue1, false)
	testListener1.OnWrite(mockStoreKey1, mockKey2, mockValue2, false)
	testListener2.OnWrite(mockStoreKey2, mockKey3, mockValue3, false)

	// expected KV pairs
	expectedKVPair1, err = testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey2.Name(),
		Key:      mockKey1,
		Value:    mockValue1,
		Delete:   false,
	})
	require.Nil(t, err)

	expectedKVPair2, err = testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey1.Name(),
		Key:      mockKey2,
		Value:    mockValue2,
		Delete:   false,
	})
	require.Nil(t, err)

	expectedKVPair3, err = testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey2.Name(),
		Key:      mockKey3,
		Value:    mockValue3,
		Delete:   false,
	})
	require.Nil(t, err)

	expectKVPairsStore1 = append(expectKVPairsStore1, expectedKVPair2)
	expectKVPairsStore2 = append(expectKVPairsStore2, expectedKVPair1, expectedKVPair3)

	// send the ABCI messages
	err = testStreamingService.ListenDeliverTx(emptyContextWrap, testDeliverTxReq2, testDeliverTxRes2)
	require.Nil(t, err)

	// write state changes
	testListener1.OnWrite(mockStoreKey1, mockKey1, mockValue1, false)
	testListener1.OnWrite(mockStoreKey1, mockKey2, mockValue2, false)
	testListener2.OnWrite(mockStoreKey2, mockKey3, mockValue3, false)

	// expected KV pairs
	expectedKVPair1, err = testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey1.Name(),
		Key:      mockKey1,
		Value:    mockValue1,
		Delete:   false,
	})
	require.Nil(t, err)

	expectedKVPair2, err = testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey1.Name(),
		Key:      mockKey2,
		Value:    mockValue2,
		Delete:   false,
	})
	require.Nil(t, err)

	expectedKVPair3, err = testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey2.Name(),
		Key:      mockKey3,
		Value:    mockValue3,
		Delete:   false,
	})
	require.Nil(t, err)

	expectKVPairsStore1 = append(expectKVPairsStore1, expectedKVPair1, expectedKVPair2)
	expectKVPairsStore2 = append(expectKVPairsStore2, expectedKVPair3)

	// send the ABCI messages
	err = testStreamingService.ListenEndBlock(emptyContextWrap, testEndBlockReq, testEndBlockRes)
	require.Nil(t, err)

	err = testStreamingService.ListenCommit(emptyContextWrap, testCommitRes)
	require.Nil(t, err)

	var version int64
	version = 1

	mockValue1Found, err := testStreamingService.VersionStore.GetAtVersion(mockStoreKey1.Name(), mockKey1, &version)
	require.Nil(t, err)
	require.Equal(t, mockValue1, mockValue1Found)

	mockValue2Found, err := testStreamingService.VersionStore.GetAtVersion(mockStoreKey1.Name(), mockKey2, &version)
	require.Nil(t, err)
	require.Equal(t, mockValue2, mockValue2Found)

	mockValue3Found, err := testStreamingService.VersionStore.GetAtVersion(mockStoreKey1.Name(), mockKey3, &version)
	require.Nil(t, err)
	require.Equal(t, mockValue3, mockValue3Found)

	mockValue1Found, err = testStreamingService.VersionStore.GetAtVersion(mockStoreKey2.Name(), mockKey1, &version)
	require.Nil(t, err)
	require.Equal(t, mockValue1, mockValue1Found)

	mockValue2Found, err = testStreamingService.VersionStore.GetAtVersion(mockStoreKey2.Name(), mockKey2, &version)
	require.Nil(t, err)
	require.Equal(t, mockValue2, mockValue2Found)

	mockValue3Found, err = testStreamingService.VersionStore.GetAtVersion(mockStoreKey2.Name(), mockKey3, &version)
	require.Nil(t, err)
	require.Equal(t, mockValue3, mockValue3Found)

}