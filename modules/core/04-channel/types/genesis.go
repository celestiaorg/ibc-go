package types

import (
	"errors"
	"fmt"

	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
)

// NewPacketState creates a new PacketState instance.
func NewPacketState(portID, channelID string, seq uint64, data []byte) PacketState {
	return PacketState{
		PortId:    portID,
		ChannelId: channelID,
		Sequence:  seq,
		Data:      data,
	}
}

// Validate performs basic validation of fields returning an error upon any
// failure.
func (pa PacketState) Validate() error {
	if pa.Data == nil {
		return errors.New("data bytes cannot be nil")
	}
	return validateGenFields(pa.PortId, pa.ChannelId, pa.Sequence)
}

// NewPacketSequence creates a new PacketSequences instance.
func NewPacketSequence(portID, channelID string, seq uint64) PacketSequence {
	return PacketSequence{
		PortId:    portID,
		ChannelId: channelID,
		Sequence:  seq,
	}
}

// Validate performs basic validation of fields returning an error upon any
// failure.
func (ps PacketSequence) Validate() error {
	return validateGenFields(ps.PortId, ps.ChannelId, ps.Sequence)
}

// NewIdentifiedUpgrade creates a new IdentifiedUpgrade instance.
func NewIdentifiedUpgrade(portID, channelID string, upgrade Upgrade) IdentifiedUpgrade {
	return IdentifiedUpgrade{
		PortId:    portID,
		ChannelId: channelID,
		Upgrade:   upgrade,
	}
}

// NewIdentifiedErrorReceipt creates a new IdentifiedErrorReceipt instance.
func NewIdentifiedErrorReceipt(portID, channelID string, errorReceipt ErrorReceipt) IdentifiedErrorReceipt {
	return IdentifiedErrorReceipt{
		PortId:       portID,
		ChannelId:    channelID,
		ErrorReceipt: errorReceipt,
	}
}

// NewGenesisState creates a GenesisState instance. It uses the default params.
// Breakage in v9.0.0 will allow the params to be provided. Please use
// NewGenesisStateWithParams in this version if you want to provide custom params.
func NewGenesisState(
	channels []IdentifiedChannel, acks, receipts, commitments []PacketState,
	sendSeqs, recvSeqs, ackSeqs []PacketSequence, nextChannelSequence uint64,
) GenesisState {
	return GenesisState{
		Channels:            channels,
		Acknowledgements:    acks,
		Commitments:         commitments,
		SendSequences:       sendSeqs,
		RecvSequences:       recvSeqs,
		AckSequences:        ackSeqs,
		NextChannelSequence: nextChannelSequence,
		Params:              DefaultParams(),
	}
}

// NewGenesisStateWithParams creates a GenesisState instance.
func NewGenesisStateWithParams(
	channels []IdentifiedChannel, acks, receipts, commitments []PacketState,
	sendSeqs, recvSeqs, ackSeqs []PacketSequence, nextChannelSequence uint64,
	params Params,
) GenesisState {
	return GenesisState{
		Channels:            channels,
		Acknowledgements:    acks,
		Commitments:         commitments,
		SendSequences:       sendSeqs,
		RecvSequences:       recvSeqs,
		AckSequences:        ackSeqs,
		NextChannelSequence: nextChannelSequence,
		Params:              params,
	}
}

// DefaultGenesisState returns the ibc channel submodule's default genesis state.
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Channels:              []IdentifiedChannel{},
		Acknowledgements:      []PacketState{},
		Receipts:              []PacketState{},
		Commitments:           []PacketState{},
		SendSequences:         []PacketSequence{},
		RecvSequences:         []PacketSequence{},
		AckSequences:          []PacketSequence{},
		NextChannelSequence:   0,
		Params:                DefaultParams(),
		Upgrades:              []IdentifiedUpgrade{},
		CounterpartyUpgrades:  []IdentifiedUpgrade{},
		ErrorReceipts:         []IdentifiedErrorReceipt{},
		PruningSequenceStarts: []PacketSequence{},
		RecvStartSequences:    []PacketSequence{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.ValidateUpgradeState(); err != nil {
		return err
	}

	// keep track of the max sequence to ensure it is less than
	// the next sequence used in creating connection identifers.
	var maxSequence uint64

	for i, channel := range gs.Channels {
		sequence, err := ParseChannelSequence(channel.ChannelId)
		if err != nil {
			return err
		}

		if sequence > maxSequence {
			maxSequence = sequence
		}

		// only the first hop is checked; ValidateUpgradeState already rejects a channel without hops
		channel.ConnectionHops = channel.ConnectionHops[:1]
		if err := channel.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid channel %v channel index %d: %w", channel, i, err)
		}
	}

	if maxSequence != 0 && maxSequence >= gs.NextChannelSequence {
		return fmt.Errorf("next channel sequence %d must be greater than maximum sequence used in channel identifier %d", gs.NextChannelSequence, maxSequence)
	}

	for i, ack := range gs.Acknowledgements {
		if err := ack.Validate(); err != nil {
			return fmt.Errorf("invalid acknowledgement %v ack index %d: %w", ack, i, err)
		}
		if len(ack.Data) == 0 {
			return fmt.Errorf("invalid acknowledgement %v ack index %d: data bytes cannot be empty", ack, i)
		}
	}

	for i, receipt := range gs.Receipts {
		if err := receipt.Validate(); err != nil {
			return fmt.Errorf("invalid acknowledgement %v ack index %d: %w", receipt, i, err)
		}
	}

	for i, commitment := range gs.Commitments {
		if err := commitment.Validate(); err != nil {
			return fmt.Errorf("invalid commitment %v index %d: %w", commitment, i, err)
		}
		if len(commitment.Data) == 0 {
			return fmt.Errorf("invalid acknowledgement %v ack index %d: data bytes cannot be empty", commitment, i)
		}
	}

	for i, ss := range gs.SendSequences {
		if err := ss.Validate(); err != nil {
			return fmt.Errorf("invalid send sequence %v index %d: %w", ss, i, err)
		}
	}

	// A recv sequence can be 0: an UNORDERED to ORDERED upgrade copies the counterparty next sequence send.
	for i, rs := range gs.RecvSequences {
		if err := validateGenIdentifiers(rs.PortId, rs.ChannelId); err != nil {
			return fmt.Errorf("invalid receive sequence %v index %d: %w", rs, i, err)
		}
	}

	for i, as := range gs.AckSequences {
		if err := as.Validate(); err != nil {
			return fmt.Errorf("invalid acknowledgement sequence %v index %d: %w", as, i, err)
		}
	}

	for i, u := range gs.Upgrades {
		if err := validateGenIdentifiers(u.PortId, u.ChannelId); err != nil {
			return fmt.Errorf("invalid upgrade %v index %d: %w", u, i, err)
		}
	}

	for i, u := range gs.CounterpartyUpgrades {
		if err := validateGenIdentifiers(u.PortId, u.ChannelId); err != nil {
			return fmt.Errorf("invalid counterparty upgrade %v index %d: %w", u, i, err)
		}
	}

	for i, er := range gs.ErrorReceipts {
		if err := validateGenIdentifiers(er.PortId, er.ChannelId); err != nil {
			return fmt.Errorf("invalid error receipt %v index %d: %w", er, i, err)
		}
	}

	for i, ps := range gs.PruningSequenceStarts {
		if err := ps.Validate(); err != nil {
			return fmt.Errorf("invalid pruning sequence start %v index %d: %w", ps, i, err)
		}
	}

	// A recv start sequence can be 0: it is copied from the counterparty upgrade without validation.
	for i, rs := range gs.RecvStartSequences {
		if err := validateGenIdentifiers(rs.PortId, rs.ChannelId); err != nil {
			return fmt.Errorf("invalid recv start sequence %v index %d: %w", rs, i, err)
		}
	}

	return nil
}

type channelKey struct{ portID, channelID string }

// ValidateUpgradeState checks that the channels and their upgrade records form a state the
// channel handshake and upgrade handlers can write. It is also run by InitGenesis.
func (gs GenesisState) ValidateUpgradeState() error {
	channels := make(map[channelKey]bool, len(gs.Channels))
	for _, ch := range gs.Channels {
		key := channelKey{ch.PortId, ch.ChannelId}
		if channels[key] {
			return fmt.Errorf("channel %s/%s: duplicate entry", ch.PortId, ch.ChannelId)
		}
		if len(ch.ConnectionHops) == 0 {
			return fmt.Errorf("channel %s/%s: connection hops cannot be empty", ch.PortId, ch.ChannelId)
		}
		channels[key] = true
	}

	sendSeqs := make(map[channelKey]uint64, len(gs.SendSequences))
	for _, ss := range gs.SendSequences {
		key := channelKey{ss.PortId, ss.ChannelId}
		if _, found := sendSeqs[key]; found {
			return fmt.Errorf("send sequence %s/%s: duplicate entry", ss.PortId, ss.ChannelId)
		}
		sendSeqs[key] = ss.Sequence
	}

	// indexRecord checks that a record names a known channel at most once.
	indexRecord := func(seen map[channelKey]bool, record, portID, channelID string) error {
		key := channelKey{portID, channelID}
		if !channels[key] {
			return fmt.Errorf("%s %s/%s: channel not found", record, portID, channelID)
		}
		if seen[key] {
			return fmt.Errorf("%s %s/%s: duplicate entry", record, portID, channelID)
		}
		seen[key] = true
		return nil
	}

	upgrades := make(map[channelKey]Upgrade, len(gs.Upgrades))
	upgradeSeen := make(map[channelKey]bool, len(gs.Upgrades))
	for _, u := range gs.Upgrades {
		if err := indexRecord(upgradeSeen, "upgrade", u.PortId, u.ChannelId); err != nil {
			return err
		}
		if len(u.Upgrade.Fields.ConnectionHops) == 0 {
			return fmt.Errorf("upgrade %s/%s: connection hops cannot be empty", u.PortId, u.ChannelId)
		}
		if u.Upgrade.Fields.Ordering != ORDERED && u.Upgrade.Fields.Ordering != UNORDERED {
			return fmt.Errorf("upgrade %s/%s: invalid ordering %s", u.PortId, u.ChannelId, u.Upgrade.Fields.Ordering)
		}
		key := channelKey{u.PortId, u.ChannelId}
		// A proposal has no timeout and no next sequence send; a full upgrade copies the channel's next sequence send.
		if u.Upgrade.NextSequenceSend == 0 {
			if u.Upgrade.Timeout.IsValid() {
				return fmt.Errorf("upgrade %s/%s: timeout must be empty when next sequence send is 0", u.PortId, u.ChannelId)
			}
		} else {
			// startFlushing always sets an absolute timeout.
			if !u.Upgrade.Timeout.IsValid() {
				return fmt.Errorf("upgrade %s/%s: timeout cannot be empty when next sequence send is set", u.PortId, u.ChannelId)
			}
			sendSeq, found := sendSeqs[key]
			if !found {
				return fmt.Errorf("upgrade %s/%s: next sequence send %d but the channel has no next sequence send", u.PortId, u.ChannelId, u.Upgrade.NextSequenceSend)
			}
			if u.Upgrade.NextSequenceSend != sendSeq {
				return fmt.Errorf("upgrade %s/%s: next sequence send %d does not match the channel next sequence send %d", u.PortId, u.ChannelId, u.Upgrade.NextSequenceSend, sendSeq)
			}
		}
		upgrades[key] = u.Upgrade
	}

	counterpartyUpgrades := make(map[channelKey]bool, len(gs.CounterpartyUpgrades))
	for _, u := range gs.CounterpartyUpgrades {
		if err := indexRecord(counterpartyUpgrades, "counterparty upgrade", u.PortId, u.ChannelId); err != nil {
			return err
		}
	}

	errorReceipts := make(map[channelKey]bool, len(gs.ErrorReceipts))
	for _, er := range gs.ErrorReceipts {
		if err := indexRecord(errorReceipts, "error receipt", er.PortId, er.ChannelId); err != nil {
			return err
		}
	}

	pruningStarts := make(map[channelKey]bool, len(gs.PruningSequenceStarts))
	for _, ps := range gs.PruningSequenceStarts {
		if err := indexRecord(pruningStarts, "pruning sequence start", ps.PortId, ps.ChannelId); err != nil {
			return err
		}
		if ps.Sequence == 0 {
			return fmt.Errorf("pruning sequence start %s/%s: sequence cannot be 0", ps.PortId, ps.ChannelId)
		}
	}

	recvStarts := make(map[channelKey]bool, len(gs.RecvStartSequences))
	for _, rs := range gs.RecvStartSequences {
		if err := indexRecord(recvStarts, "recv start sequence", rs.PortId, rs.ChannelId); err != nil {
			return err
		}
	}

	commitments := make(map[channelKey]bool)
	for _, c := range gs.Commitments {
		commitments[channelKey{c.PortId, c.ChannelId}] = true
	}

	for _, ch := range gs.Channels {
		key := channelKey{ch.PortId, ch.ChannelId}
		upgrade, hasUpgrade := upgrades[key]
		hasFullUpgrade := hasUpgrade && upgrade.NextSequenceSend != 0
		hasCounterpartyUpgrade := counterpartyUpgrades[key]

		var reason string
		switch ch.State {
		case INIT, TRYOPEN:
			if hasUpgrade || hasCounterpartyUpgrade || errorReceipts[key] || pruningStarts[key] || recvStarts[key] {
				reason = "cannot have upgrade records"
			}
		case OPEN:
			switch {
			case hasFullUpgrade:
				reason = "cannot have a full upgrade"
			case hasCounterpartyUpgrade:
				reason = "cannot have a counterparty upgrade"
			}
		case FLUSHING:
			if !hasFullUpgrade {
				reason = "requires a full upgrade"
			}
		case FLUSHCOMPLETE:
			switch {
			case !hasFullUpgrade:
				reason = "requires a full upgrade"
			case !hasCounterpartyUpgrade:
				reason = "requires a counterparty upgrade"
			case commitments[key]:
				reason = "cannot have packet commitments"
			}
		case CLOSED:
			if hasCounterpartyUpgrade && !hasFullUpgrade {
				reason = "cannot have a counterparty upgrade without a full upgrade"
			}
		default:
			reason = "invalid state"
		}
		if reason != "" {
			return fmt.Errorf("channel %s/%s in state %s: %s", ch.PortId, ch.ChannelId, ch.State, reason)
		}

		if pruningStarts[key] != recvStarts[key] {
			return fmt.Errorf("channel %s/%s: pruning sequence start and recv start sequence must be set together", ch.PortId, ch.ChannelId)
		}
	}

	return nil
}

func validateGenFields(portID, channelID string, sequence uint64) error {
	if err := validateGenIdentifiers(portID, channelID); err != nil {
		return err
	}
	if sequence == 0 {
		return errors.New("sequence cannot be 0")
	}
	return nil
}

func validateGenIdentifiers(portID, channelID string) error {
	if err := host.PortIdentifierValidator(portID); err != nil {
		return fmt.Errorf("invalid port Id: %w", err)
	}
	if err := host.ChannelIdentifierValidator(channelID); err != nil {
		return fmt.Errorf("invalid channel Id: %w", err)
	}
	return nil
}
