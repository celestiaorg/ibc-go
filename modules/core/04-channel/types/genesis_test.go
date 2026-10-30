package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channel "github.com/cosmos/ibc-go/v8/modules/core/04-channel"
	"github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v8/testing/simapp"
)

const (
	testPort1         = "firstport"
	testPort2         = "secondport"
	testConnectionIDA = "connectionidatob"

	testChannel1 = "channel-0"
	testChannel2 = "channel-1"

	testChannelOrder   = types.ORDERED
	testChannelVersion = "1.0"
)

func TestValidateGenesis(t *testing.T) {
	counterparty1 := types.NewCounterparty(testPort1, testChannel1)
	counterparty2 := types.NewCounterparty(testPort2, testChannel2)
	testCases := []struct {
		name     string
		genState types.GenesisState
		expPass  bool
	}{
		{
			name:     "default",
			genState: types.DefaultGenesisState(),
			expPass:  true,
		},
		{
			name: "valid genesis",
			genState: types.NewGenesisState(
				[]types.IdentifiedChannel{
					types.NewIdentifiedChannel(
						testPort1, testChannel1, types.NewChannel(
							types.INIT, testChannelOrder, counterparty2, []string{testConnectionIDA}, testChannelVersion,
						),
					),
					types.NewIdentifiedChannel(
						testPort2, testChannel2, types.NewChannel(
							types.INIT, testChannelOrder, counterparty1, []string{testConnectionIDA}, testChannelVersion,
						),
					),
				},
				[]types.PacketState{
					types.NewPacketState(testPort2, testChannel2, 1, []byte("ack")),
				},
				[]types.PacketState{
					types.NewPacketState(testPort2, testChannel2, 1, []byte("")),
				},
				[]types.PacketState{
					types.NewPacketState(testPort1, testChannel1, 1, []byte("commit_hash")),
				},
				[]types.PacketSequence{
					types.NewPacketSequence(testPort1, testChannel1, 1),
				},
				[]types.PacketSequence{
					types.NewPacketSequence(testPort2, testChannel2, 1),
				},
				[]types.PacketSequence{
					types.NewPacketSequence(testPort2, testChannel2, 1),
				},
				2,
			),
			expPass: true,
		},
		{
			name: "invalid channel",
			genState: types.GenesisState{
				Channels: []types.IdentifiedChannel{
					types.NewIdentifiedChannel(
						testPort1, "(testChannel1)", types.NewChannel(
							types.INIT, testChannelOrder, counterparty2, []string{testConnectionIDA}, testChannelVersion,
						),
					),
				},
			},
			expPass: false,
		},
		{
			name: "invalid ack",
			genState: types.GenesisState{
				Acknowledgements: []types.PacketState{
					types.NewPacketState(testPort2, testChannel2, 1, nil),
				},
			},
			expPass: false,
		},
		{
			name: "invalid commitment",
			genState: types.GenesisState{
				Commitments: []types.PacketState{
					types.NewPacketState(testPort1, testChannel1, 1, nil),
				},
			},
			expPass: false,
		},
		{
			name: "invalid send seq",
			genState: types.GenesisState{
				SendSequences: []types.PacketSequence{
					types.NewPacketSequence(testPort1, testChannel1, 0),
				},
			},
			expPass: false,
		},
		{
			name: "invalid recv seq",
			genState: types.GenesisState{
				RecvSequences: []types.PacketSequence{
					types.NewPacketSequence(testPort1, "(testChannel1)", 1),
				},
			},
			expPass: false,
		},
		{
			name: "invalid recv seq 2",
			genState: types.GenesisState{
				RecvSequences: []types.PacketSequence{
					types.NewPacketSequence("(testPort1)", testChannel1, 1),
				},
			},
			expPass: false,
		},
		{
			name: "valid recv seq 0",
			genState: types.GenesisState{
				RecvSequences: []types.PacketSequence{
					types.NewPacketSequence(testPort1, testChannel1, 0),
				},
			},
			expPass: true,
		},
		{
			name: "invalid ack seq 0",
			genState: types.GenesisState{
				AckSequences: []types.PacketSequence{
					types.NewPacketSequence(testPort1, testChannel1, 0),
				},
			},
			expPass: false,
		},
		{
			name: "valid channel with an unchecked second hop",
			genState: func() types.GenesisState {
				gs := upgradeGenState()
				gs.Channels[0].ConnectionHops = []string{testConnectionIDA, ""}
				return gs
			}(),
			expPass: true,
		},
		{
			name: "invalid first hop",
			genState: func() types.GenesisState {
				gs := upgradeGenState()
				gs.Channels[0].ConnectionHops = []string{"", testConnectionIDA}
				return gs
			}(),
			expPass: false,
		},
		{
			name: "invalid ack seq",
			genState: types.GenesisState{
				AckSequences: []types.PacketSequence{
					types.NewPacketSequence(testPort1, "(testChannel1)", 1),
				},
			},
			expPass: false,
		},
		{
			name:     "valid upgrade records",
			genState: upgradeGenState(),
			expPass:  true,
		},
		{
			name: "invalid upgrade",
			genState: types.GenesisState{
				Upgrades: []types.IdentifiedUpgrade{types.NewIdentifiedUpgrade(testPort1, "(testChannel1)", types.Upgrade{})},
			},
			expPass: false,
		},
		{
			name: "invalid counterparty upgrade",
			genState: types.GenesisState{
				CounterpartyUpgrades: []types.IdentifiedUpgrade{types.NewIdentifiedUpgrade("(testPort1)", testChannel1, types.Upgrade{})},
			},
			expPass: false,
		},
		{
			name: "invalid error receipt",
			genState: types.GenesisState{
				ErrorReceipts: []types.IdentifiedErrorReceipt{types.NewIdentifiedErrorReceipt(testPort1, "(testChannel1)", types.ErrorReceipt{Sequence: 1})},
			},
			expPass: false,
		},
		{
			name: "invalid pruning sequence start",
			genState: types.GenesisState{
				PruningSequenceStarts: []types.PacketSequence{types.NewPacketSequence(testPort1, testChannel1, 0)},
			},
			expPass: false,
		},
		{
			name: "valid recv start sequence 0",
			genState: func() types.GenesisState {
				gs := upgradeGenState()
				gs.RecvStartSequences[0].Sequence = 0
				return gs
			}(),
			expPass: true,
		},
		{
			name: "invalid recv start sequence",
			genState: types.GenesisState{
				RecvStartSequences: []types.PacketSequence{types.NewPacketSequence(testPort1, "(testChannel1)", 1)},
			},
			expPass: false,
		},
		{
			name: "invalid channel identifier",
			genState: types.NewGenesisState(
				[]types.IdentifiedChannel{
					types.NewIdentifiedChannel(
						testPort1, "chan-0", types.NewChannel(
							types.INIT, testChannelOrder, counterparty2, []string{testConnectionIDA}, testChannelVersion,
						),
					),
					types.NewIdentifiedChannel(
						testPort2, testChannel2, types.NewChannel(
							types.INIT, testChannelOrder, counterparty1, []string{testConnectionIDA}, testChannelVersion,
						),
					),
				},
				[]types.PacketState{
					types.NewPacketState(testPort2, testChannel2, 1, []byte("ack")),
				},
				[]types.PacketState{
					types.NewPacketState(testPort2, testChannel2, 1, []byte("")),
				},
				[]types.PacketState{
					types.NewPacketState(testPort1, testChannel1, 1, []byte("commit_hash")),
				},
				[]types.PacketSequence{
					types.NewPacketSequence(testPort1, testChannel1, 1),
				},
				[]types.PacketSequence{
					types.NewPacketSequence(testPort2, testChannel2, 1),
				},
				[]types.PacketSequence{
					types.NewPacketSequence(testPort2, testChannel2, 1),
				},
				0,
			),
			expPass: false,
		},
		{
			name: "next channel sequence is less than maximum channel identifier sequence used",
			genState: types.NewGenesisState(
				[]types.IdentifiedChannel{
					types.NewIdentifiedChannel(
						testPort1, "channel-10", types.NewChannel(
							types.INIT, testChannelOrder, counterparty2, []string{testConnectionIDA}, testChannelVersion,
						),
					),
					types.NewIdentifiedChannel(
						testPort2, testChannel2, types.NewChannel(
							types.INIT, testChannelOrder, counterparty1, []string{testConnectionIDA}, testChannelVersion,
						),
					),
				},
				[]types.PacketState{
					types.NewPacketState(testPort2, testChannel2, 1, []byte("ack")),
				},
				[]types.PacketState{
					types.NewPacketState(testPort2, testChannel2, 1, []byte("")),
				},
				[]types.PacketState{
					types.NewPacketState(testPort1, testChannel1, 1, []byte("commit_hash")),
				},
				[]types.PacketSequence{
					types.NewPacketSequence(testPort1, testChannel1, 1),
				},
				[]types.PacketSequence{
					types.NewPacketSequence(testPort2, testChannel2, 1),
				},
				[]types.PacketSequence{
					types.NewPacketSequence(testPort2, testChannel2, 1),
				},
				0,
			),
			expPass: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		err := tc.genState.Validate()
		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}

// upgradeGenState returns a valid genesis with one FLUSHCOMPLETE channel and every upgrade record.
func upgradeGenState() types.GenesisState {
	upgrade := types.NewUpgrade(
		types.NewUpgradeFields(types.UNORDERED, []string{testConnectionIDA}, testChannelVersion),
		types.NewTimeout(clienttypes.NewHeight(1, 100), 0),
		1,
	)
	ch := types.NewChannel(types.FLUSHCOMPLETE, types.UNORDERED, types.NewCounterparty(testPort2, testChannel2), []string{testConnectionIDA}, testChannelVersion)
	ch.UpgradeSequence = 2

	gs := types.DefaultGenesisState()
	gs.Channels = []types.IdentifiedChannel{types.NewIdentifiedChannel(testPort1, testChannel1, ch)}
	gs.SendSequences = []types.PacketSequence{types.NewPacketSequence(testPort1, testChannel1, 1)}
	gs.Upgrades = []types.IdentifiedUpgrade{types.NewIdentifiedUpgrade(testPort1, testChannel1, upgrade)}
	gs.CounterpartyUpgrades = []types.IdentifiedUpgrade{types.NewIdentifiedUpgrade(testPort1, testChannel1, upgrade)}
	gs.ErrorReceipts = []types.IdentifiedErrorReceipt{types.NewIdentifiedErrorReceipt(testPort1, testChannel1, types.ErrorReceipt{Sequence: 1, Message: "upgrade failed"})}
	gs.PruningSequenceStarts = []types.PacketSequence{types.NewPacketSequence(testPort1, testChannel1, 1)}
	gs.RecvStartSequences = []types.PacketSequence{types.NewPacketSequence(testPort1, testChannel1, 1)}
	return gs
}

// keepOneRecord removes every upgrade record kind except the one at index kind.
func keepOneRecord(gs *types.GenesisState, kind int) {
	if kind != 0 {
		gs.Upgrades = nil
	}
	if kind != 1 {
		gs.CounterpartyUpgrades = nil
	}
	if kind != 2 {
		gs.ErrorReceipts = nil
	}
	if kind != 3 {
		gs.PruningSequenceStarts = nil
	}
	if kind != 4 {
		gs.RecvStartSequences = nil
	}
}

func TestValidateUpgradeState(t *testing.T) {
	proposal := types.NewUpgrade(types.NewUpgradeFields(types.UNORDERED, []string{testConnectionIDA}, testChannelVersion), types.Timeout{}, 0)
	type testCase struct {
		name     string
		malleate func(gs *types.GenesisState)
		expErr   string
	}
	testCases := []testCase{
		{"upgrade without a channel", func(gs *types.GenesisState) { gs.Upgrades[0].ChannelId = testChannel2 }, "upgrade firstport/channel-1: channel not found"},
		{"counterparty upgrade without a channel", func(gs *types.GenesisState) { gs.CounterpartyUpgrades[0].ChannelId = testChannel2 }, "counterparty upgrade firstport/channel-1: channel not found"},
		{"error receipt without a channel", func(gs *types.GenesisState) { gs.ErrorReceipts[0].ChannelId = testChannel2 }, "error receipt firstport/channel-1: channel not found"},
		{"recv start without a channel", func(gs *types.GenesisState) { gs.RecvStartSequences[0].ChannelId = testChannel2 }, "recv start sequence firstport/channel-1: channel not found"},
		{"pruning start without a channel", func(gs *types.GenesisState) { gs.PruningSequenceStarts[0].ChannelId = testChannel2 }, "pruning sequence start firstport/channel-1: channel not found"},
		{"duplicate upgrade", func(gs *types.GenesisState) { gs.Upgrades = append(gs.Upgrades, gs.Upgrades[0]) }, "upgrade firstport/channel-0: duplicate entry"},
		{"duplicate counterparty upgrade", func(gs *types.GenesisState) {
			gs.CounterpartyUpgrades = append(gs.CounterpartyUpgrades, gs.CounterpartyUpgrades[0])
		}, "counterparty upgrade firstport/channel-0: duplicate entry"},
		{"duplicate error receipt", func(gs *types.GenesisState) { gs.ErrorReceipts = append(gs.ErrorReceipts, gs.ErrorReceipts[0]) }, "error receipt firstport/channel-0: duplicate entry"},
		{"duplicate recv start", func(gs *types.GenesisState) {
			gs.RecvStartSequences = append(gs.RecvStartSequences, gs.RecvStartSequences[0])
		}, "recv start sequence firstport/channel-0: duplicate entry"},
		{"duplicate pruning start", func(gs *types.GenesisState) {
			gs.PruningSequenceStarts = append(gs.PruningSequenceStarts, gs.PruningSequenceStarts[0])
		}, "pruning sequence start firstport/channel-0: duplicate entry"},
		{"duplicate channel", func(gs *types.GenesisState) { gs.Channels = append(gs.Channels, gs.Channels[0]) }, "channel firstport/channel-0: duplicate entry"},
		{"duplicate send sequence", func(gs *types.GenesisState) { gs.SendSequences = append(gs.SendSequences, gs.SendSequences[0]) }, "send sequence firstport/channel-0: duplicate entry"},
		{"channel without hops", func(gs *types.GenesisState) { gs.Channels[0].ConnectionHops = nil }, "channel firstport/channel-0: connection hops cannot be empty"},
		{"upgrade without hops", func(gs *types.GenesisState) { gs.Upgrades[0].Upgrade.Fields.ConnectionHops = nil }, "upgrade firstport/channel-0: connection hops cannot be empty"},
		{"upgrade with next sequence send 0 and a timeout", func(gs *types.GenesisState) { gs.Upgrades[0].Upgrade.NextSequenceSend = 0 }, "upgrade firstport/channel-0: timeout must be empty when next sequence send is 0"},
		{"upgrade with a wrong next sequence send", func(gs *types.GenesisState) { gs.Upgrades[0].Upgrade.NextSequenceSend = 2 }, "upgrade firstport/channel-0: next sequence send 2 does not match the channel next sequence send 1"},
		{"full upgrade without a send sequence", func(gs *types.GenesisState) { gs.SendSequences = nil }, "upgrade firstport/channel-0: next sequence send 1 but the channel has no next sequence send"},
		{"OPEN with a full upgrade", func(gs *types.GenesisState) {
			gs.Channels[0].State = types.OPEN
			gs.CounterpartyUpgrades = nil
		}, "channel firstport/channel-0 in state STATE_OPEN: cannot have a full upgrade"},
		{"OPEN with a counterparty upgrade", func(gs *types.GenesisState) {
			gs.Channels[0].State = types.OPEN
			gs.Upgrades[0].Upgrade = proposal
		}, "channel firstport/channel-0 in state STATE_OPEN: cannot have a counterparty upgrade"},
		{"FLUSHING with a proposal upgrade", func(gs *types.GenesisState) {
			gs.Channels[0].State = types.FLUSHING
			gs.Upgrades[0].Upgrade = proposal
		}, "channel firstport/channel-0 in state STATE_FLUSHING: requires a full upgrade"},
		{"FLUSHING without an upgrade", func(gs *types.GenesisState) {
			gs.Channels[0].State = types.FLUSHING
			gs.Upgrades, gs.CounterpartyUpgrades = nil, nil
		}, "channel firstport/channel-0 in state STATE_FLUSHING: requires a full upgrade"},
		{"FLUSHCOMPLETE with a proposal upgrade", func(gs *types.GenesisState) { gs.Upgrades[0].Upgrade = proposal }, "channel firstport/channel-0 in state STATE_FLUSHCOMPLETE: requires a full upgrade"},
		{"FLUSHCOMPLETE without an upgrade", func(gs *types.GenesisState) { gs.Upgrades = nil }, "channel firstport/channel-0 in state STATE_FLUSHCOMPLETE: requires a full upgrade"},
		{"FLUSHCOMPLETE without a counterparty upgrade", func(gs *types.GenesisState) { gs.CounterpartyUpgrades = nil }, "channel firstport/channel-0 in state STATE_FLUSHCOMPLETE: requires a counterparty upgrade"},
		{"FLUSHCOMPLETE with a packet commitment", func(gs *types.GenesisState) {
			gs.Commitments = []types.PacketState{types.NewPacketState(testPort1, testChannel1, 1, []byte("commit_hash"))}
		}, "channel firstport/channel-0 in state STATE_FLUSHCOMPLETE: cannot have packet commitments"},
		{"CLOSED with a proposal upgrade and a counterparty upgrade", func(gs *types.GenesisState) {
			gs.Channels[0].State = types.CLOSED
			gs.Upgrades[0].Upgrade = proposal
		}, "channel firstport/channel-0 in state STATE_CLOSED: cannot have a counterparty upgrade without a full upgrade"},
		{"counterparty upgrade without an upgrade", func(gs *types.GenesisState) {
			gs.Channels[0].State = types.CLOSED
			gs.Upgrades = nil
		}, "channel firstport/channel-0 in state STATE_CLOSED: cannot have a counterparty upgrade without a full upgrade"},
		{"upgrade with ordering NONE", func(gs *types.GenesisState) { gs.Upgrades[0].Upgrade.Fields.Ordering = types.NONE }, "upgrade firstport/channel-0: invalid ordering ORDER_NONE_UNSPECIFIED"},
		{"uninitialized state", func(gs *types.GenesisState) { gs.Channels[0].State = types.UNINITIALIZED }, "channel firstport/channel-0 in state STATE_UNINITIALIZED_UNSPECIFIED: invalid state"},
		{"unknown state", func(gs *types.GenesisState) { gs.Channels[0].State = types.State(7) }, "channel firstport/channel-0 in state 7: invalid state"},
		{"recv start without pruning start", func(gs *types.GenesisState) { gs.PruningSequenceStarts = nil }, "channel firstport/channel-0: pruning sequence start and recv start sequence must be set together"},
		{"pruning start without recv start", func(gs *types.GenesisState) { gs.RecvStartSequences = nil }, "channel firstport/channel-0: pruning sequence start and recv start sequence must be set together"},
		{"pruning start 0", func(gs *types.GenesisState) { gs.PruningSequenceStarts[0].Sequence = 0 }, "pruning sequence start firstport/channel-0: sequence cannot be 0"},
	}
	for _, state := range []types.State{types.INIT, types.TRYOPEN} {
		for kind, record := range []string{"upgrade", "counterparty upgrade", "error receipt", "pruning start", "recv start"} {
			testCases = append(testCases, testCase{
				fmt.Sprintf("%s with only a %s", state, record),
				func(gs *types.GenesisState) {
					gs.Channels[0].State = state
					keepOneRecord(gs, kind)
				},
				fmt.Sprintf("channel firstport/channel-0 in state %s: cannot have upgrade records", state),
			})
		}
	}

	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false)

	valid := upgradeGenState()
	require.NoError(t, valid.Validate())
	cacheCtx, _ := ctx.CacheContext()
	require.NotPanics(t, func() { channel.InitGenesis(cacheCtx, app.IBCKeeper.ChannelKeeper, valid) })

	for _, tc := range testCases {
		gs := upgradeGenState()
		tc.malleate(&gs)
		require.EqualError(t, gs.Validate(), tc.expErr, tc.name)
		cacheCtx, _ := ctx.CacheContext()
		require.PanicsWithValue(t, "invalid ibc channel genesis state: "+tc.expErr, func() {
			channel.InitGenesis(cacheCtx, app.IBCKeeper.ChannelKeeper, gs)
		}, tc.name)
	}
}
