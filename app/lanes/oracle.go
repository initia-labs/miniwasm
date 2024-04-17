package lanes

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	opchildtypes "github.com/initia-labs/OPinit/x/opchild/types"

	initiaapplanes "github.com/initia-labs/initia/app/lanes"
	"github.com/skip-mev/block-sdk/v2/block"
	blockbase "github.com/skip-mev/block-sdk/v2/block/base"
)

// OracleLaneMatchHandler returns the default match handler for the oracle lane. The
// default implementation matches transactions that are oracle related. In particular,
// any transaction that is only a MsgUpdateOracle.
func OracleLaneMatchHandler() blockbase.MatchHandler {
	return func(ctx sdk.Context, tx sdk.Tx) bool {
		msgs := tx.GetMsgs()
		if len(msgs) != 1 {
			return false
		}

		if _, ok := msgs[0].(*opchildtypes.MsgUpdateOracle); !ok {
			return false
		}
		return true
	}
}

const (
	// FreeLaneName defines the name of the free lane.
	OracleLaneName = "oracle"
)

// NewFreeLane returns a new free lane.
func NewOracleLane(
	cfg blockbase.LaneConfig,
) block.Lane {
	lane := &blockbase.BaseLane{}
	proposalHandler := initiaapplanes.NewDefaultProposalHandler(lane)

	_lane, err := blockbase.NewBaseLane(
		cfg,
		OracleLaneName,
		blockbase.WithMatchHandler(OracleLaneMatchHandler()),
		blockbase.WithMempool(initiaapplanes.NewMempool(blockbase.NewDefaultTxPriority(), cfg.SignerExtractor, cfg.MaxTxs)),
		blockbase.WithPrepareLaneHandler(proposalHandler.PrepareLaneHandler()),
		blockbase.WithProcessLaneHandler(proposalHandler.ProcessLaneHandler()),
	)
	if err != nil {
		panic(err)
	}

	*lane = *_lane
	return lane
}
