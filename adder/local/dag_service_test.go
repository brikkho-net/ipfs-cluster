package local

import (
	"context"
	"errors"
	"mime/multipart"
	"sync"
	"testing"

	peer "github.com/libp2p/go-libp2p-peer"

	adder "github.com/ipfs/ipfs-cluster/adder"
	"github.com/ipfs/ipfs-cluster/api"
	"github.com/ipfs/ipfs-cluster/test"

	rpc "github.com/libp2p/go-libp2p-gorpc"
)

type testRPC struct {
	blocks sync.Map
	pins   sync.Map
}

func (rpcs *testRPC) IPFSBlockPut(ctx context.Context, in *api.NodeWithMeta, out *struct{}) error {
	rpcs.blocks.Store(in.Cid.String(), in)
	return nil
}

func (rpcs *testRPC) Pin(ctx context.Context, in *api.Pin, out *struct{}) error {
	rpcs.pins.Store(in.Cid.String(), in)
	return nil
}

func (rpcs *testRPC) BlockAllocate(ctx context.Context, in *api.Pin, out *[]peer.ID) error {
	if in.ReplicationFactorMin > 1 {
		return errors.New("we can only replicate to 1 peer")
	}
	// it does not matter since we use host == nil for RPC, so it uses the
	// local one in all cases.
	*out = []peer.ID{test.PeerID1}
	return nil
}

func TestAdd(t *testing.T) {
	t.Run("balanced", func(t *testing.T) {
		rpcObj := &testRPC{}
		server := rpc.NewServer(nil, "mock")
		err := server.RegisterName("Cluster", rpcObj)
		if err != nil {
			t.Fatal(err)
		}
		client := rpc.NewClientWithServer(nil, "mock", server)
		params := api.DefaultAddParams()
		params.Wrap = true

		dags := New(client, params.PinOptions)
		add := adder.New(dags, params, nil)

		sth := test.NewShardingTestHelper()
		defer sth.Clean(t)
		mr, closer := sth.GetTreeMultiReader(t)
		defer closer.Close()
		r := multipart.NewReader(mr, mr.Boundary())

		rootCid, err := add.FromMultipart(context.Background(), r)
		if err != nil {
			t.Fatal(err)
		}

		if rootCid.String() != test.ShardingDirBalancedRootCIDWrapped {
			t.Fatal("bad root cid: ", rootCid)
		}

		expected := test.ShardingDirCids[:]
		for _, c := range expected {
			_, ok := rpcObj.blocks.Load(c)
			if !ok {
				t.Error("no IPFSBlockPut for block", c)
			}
		}

		_, ok := rpcObj.pins.Load(test.ShardingDirBalancedRootCIDWrapped)
		if !ok {
			t.Error("the tree wasn't pinned")
		}
	})

	t.Run("trickle", func(t *testing.T) {
		rpcObj := &testRPC{}
		server := rpc.NewServer(nil, "mock")
		err := server.RegisterName("Cluster", rpcObj)
		if err != nil {
			t.Fatal(err)
		}
		client := rpc.NewClientWithServer(nil, "mock", server)
		params := api.DefaultAddParams()
		params.Layout = "trickle"

		dags := New(client, params.PinOptions)
		add := adder.New(dags, params, nil)

		sth := test.NewShardingTestHelper()
		defer sth.Clean(t)
		mr, closer := sth.GetTreeMultiReader(t)
		defer closer.Close()
		r := multipart.NewReader(mr, mr.Boundary())

		rootCid, err := add.FromMultipart(context.Background(), r)
		if err != nil {
			t.Fatal(err)
		}

		if rootCid.String() != test.ShardingDirTrickleRootCID {
			t.Fatal("bad root cid")
		}

		_, ok := rpcObj.pins.Load(test.ShardingDirTrickleRootCID)
		if !ok {
			t.Error("the tree wasn't pinned")
		}
	})
}
