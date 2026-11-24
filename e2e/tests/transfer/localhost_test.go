//go:build e2e

package transfer

import (
	"context"
	"testing"
	"time"

	test "github.com/strangelove-ventures/interchaintest/v8/testutil"
	testifysuite "github.com/stretchr/testify/suite"

	"github.com/cosmos/ibc-go/e2e/testsuite"
	"github.com/cosmos/ibc-go/e2e/testsuite/query"
	"github.com/cosmos/ibc-go/e2e/testvalues"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	localhost "github.com/cosmos/ibc-go/v8/modules/light-clients/09-localhost"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
)

func TestTransferLocalhostTestSuite(t *testing.T) {
	testifysuite.Run(t, new(LocalhostTransferTestSuite))
}

type LocalhostTransferTestSuite struct {
	testsuite.E2ETestSuite
}

// TestMsgTransfer_Localhost creates two wallets on a single chain and performs MsgTransfers back and forth
// to ensure ibc functions as expected on localhost. This test is largely the same as TestMsgTransfer_Succeeds_Nonincentivized
// except that chain B is replaced with an additional wallet on chainA.
func (s *LocalhostTransferTestSuite) TestMsgTransfer_Localhost() {
	t := s.T()
	ctx := context.TODO()

	_, _ = s.SetupChainsRelayerAndChannel(ctx, s.TransferChannelOptions())
	chainA, _ := s.GetChains()

	chainADenom := chainA.Config().Denom

	rlyWallet := s.CreateUserOnChainA(ctx, testvalues.StartingTokenAmount)
	userAWallet := s.CreateUserOnChainA(ctx, testvalues.StartingTokenAmount)
	userBWallet := s.CreateUserOnChainA(ctx, testvalues.StartingTokenAmount)

	var (
		msgChanOpenInitRes channeltypes.MsgChannelOpenInitResponse
		msgChanOpenTryRes  channeltypes.MsgChannelOpenTryResponse
		ack                []byte
		packet             channeltypes.Packet
	)

	s.Require().NoError(test.WaitForBlocks(ctx, 1, chainA), "failed to wait for blocks")

	if !s.Run("verify begin blocker was executed", func() {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		cs, err := query.ClientState(ctx, chainA, exported.LocalhostClientID)
		s.Require().NoError(err)

		localhostClientState, ok := cs.(*localhost.ClientState)
		s.Require().True(ok)
		originalHeight := localhostClientState.LatestHeight
		latestHeight := originalHeight

		// The consensus RPC height can advance before the application commits.
		// Poll the client state itself so we wait for a committed height update.
		err = test.WaitForCondition(30*time.Second, 100*time.Millisecond, func() (bool, error) {
			cs, err := query.ClientState(ctx, chainA, exported.LocalhostClientID)
			if err != nil {
				return false, err
			}

			localhostClientState, ok := cs.(*localhost.ClientState)
			s.Require().True(ok)
			latestHeight = localhostClientState.LatestHeight
			return latestHeight.GT(originalHeight), nil
		})
		s.Require().NoError(err, "client state height was not incremented: original %s, latest %s", originalHeight, latestHeight)
	}) {
		return
	}

	t.Run("channel open init localhost", func(t *testing.T) {
		msgChanOpenInit := channeltypes.NewMsgChannelOpenInit(
			transfertypes.PortID, transfertypes.Version,
			channeltypes.UNORDERED, []string{exported.LocalhostConnectionID},
			transfertypes.PortID, rlyWallet.FormattedAddress(),
		)

		s.Require().NoError(msgChanOpenInit.ValidateBasic())

		txResp := s.BroadcastMessages(ctx, chainA, rlyWallet, msgChanOpenInit)
		s.AssertTxSuccess(txResp)

		s.Require().NoError(testsuite.UnmarshalMsgResponses(txResp, &msgChanOpenInitRes))
	})

	t.Run("channel open try localhost", func(t *testing.T) {
		msgChanOpenTry := channeltypes.NewMsgChannelOpenTry(
			transfertypes.PortID, transfertypes.Version,
			channeltypes.UNORDERED, []string{exported.LocalhostConnectionID},
			transfertypes.PortID, msgChanOpenInitRes.ChannelId,
			transfertypes.Version, localhost.SentinelProof, clienttypes.ZeroHeight(), rlyWallet.FormattedAddress(),
		)

		txResp := s.BroadcastMessages(ctx, chainA, rlyWallet, msgChanOpenTry)
		s.AssertTxSuccess(txResp)

		s.Require().NoError(testsuite.UnmarshalMsgResponses(txResp, &msgChanOpenTryRes))
	})

	t.Run("channel open ack localhost", func(t *testing.T) {
		msgChanOpenAck := channeltypes.NewMsgChannelOpenAck(
			transfertypes.PortID, msgChanOpenInitRes.ChannelId,
			msgChanOpenTryRes.ChannelId, transfertypes.Version,
			localhost.SentinelProof, clienttypes.ZeroHeight(), rlyWallet.FormattedAddress(),
		)

		txResp := s.BroadcastMessages(ctx, chainA, rlyWallet, msgChanOpenAck)
		s.AssertTxSuccess(txResp)
	})

	t.Run("channel open confirm localhost", func(t *testing.T) {
		msgChanOpenConfirm := channeltypes.NewMsgChannelOpenConfirm(
			transfertypes.PortID, msgChanOpenTryRes.ChannelId,
			localhost.SentinelProof, clienttypes.ZeroHeight(), rlyWallet.FormattedAddress(),
		)

		txResp := s.BroadcastMessages(ctx, chainA, rlyWallet, msgChanOpenConfirm)
		s.AssertTxSuccess(txResp)
	})

	t.Run("query localhost transfer channel ends", func(t *testing.T) {
		channelEndA, err := query.Channel(ctx, chainA, transfertypes.PortID, msgChanOpenInitRes.ChannelId)
		s.Require().NoError(err)
		s.Require().NotNil(channelEndA)

		channelEndB, err := query.Channel(ctx, chainA, transfertypes.PortID, msgChanOpenTryRes.ChannelId)
		s.Require().NoError(err)
		s.Require().NotNil(channelEndB)

		s.Require().Equal(channelEndA.ConnectionHops, channelEndB.ConnectionHops)
	})

	t.Run("send packet localhost ibc transfer", func(t *testing.T) {
		var err error
		txResp := s.Transfer(ctx, chainA, userAWallet, transfertypes.PortID, msgChanOpenInitRes.ChannelId, testvalues.DefaultTransferAmount(chainADenom), userAWallet.FormattedAddress(), userBWallet.FormattedAddress(), clienttypes.NewHeight(1, 100), 0, "")
		s.AssertTxSuccess(txResp)

		packet, err = ibctesting.ParsePacketFromEvents(txResp.Events)
		s.Require().NoError(err)
		s.Require().NotNil(packet)
	})

	t.Run("tokens are escrowed", func(t *testing.T) {
		actualBalance, err := s.GetChainANativeBalance(ctx, userAWallet)
		s.Require().NoError(err)

		expected := testvalues.StartingTokenAmount - testvalues.IBCTransferAmount
		s.Require().Equal(expected, actualBalance)
	})

	t.Run("recv packet localhost ibc transfer", func(t *testing.T) {
		var err error
		msgRecvPacket := channeltypes.NewMsgRecvPacket(packet, localhost.SentinelProof, clienttypes.ZeroHeight(), rlyWallet.FormattedAddress())

		txResp := s.BroadcastMessages(ctx, chainA, rlyWallet, msgRecvPacket)
		s.AssertTxSuccess(txResp)

		ack, err = ibctesting.ParseAckFromEvents(txResp.Events)
		s.Require().NoError(err)
		s.Require().NotNil(ack)
	})

	t.Run("acknowledge packet localhost ibc transfer", func(t *testing.T) {
		msgAcknowledgement := channeltypes.NewMsgAcknowledgement(packet, ack, localhost.SentinelProof, clienttypes.ZeroHeight(), rlyWallet.FormattedAddress())

		txResp := s.BroadcastMessages(ctx, chainA, rlyWallet, msgAcknowledgement)
		s.AssertTxSuccess(txResp)
	})

	t.Run("verify tokens transferred", func(t *testing.T) {
		s.AssertPacketRelayed(ctx, chainA, transfertypes.PortID, msgChanOpenInitRes.ChannelId, 1)

		ibcToken := testsuite.GetIBCToken(chainADenom, transfertypes.PortID, msgChanOpenTryRes.ChannelId)
		actualBalance, err := query.Balance(ctx, chainA, userBWallet.FormattedAddress(), ibcToken.IBCDenom())

		s.Require().NoError(err)

		expected := testvalues.IBCTransferAmount
		s.Require().Equal(expected, actualBalance.Int64())
	})
}
