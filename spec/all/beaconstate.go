// Copyright © 2023 Attestant Limited.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package all

import (
	"errors"
	"fmt"
	"reflect"

	bitfield "github.com/OffchainLabs/go-bitfield"
	"github.com/ethpandaops/go-eth2-client/spec/altair"
	"github.com/ethpandaops/go-eth2-client/spec/bellatrix"
	"github.com/ethpandaops/go-eth2-client/spec/capella"
	"github.com/ethpandaops/go-eth2-client/spec/deneb"
	"github.com/ethpandaops/go-eth2-client/spec/electra"
	"github.com/ethpandaops/go-eth2-client/spec/fulu"
	"github.com/ethpandaops/go-eth2-client/spec/gloas"
	"github.com/ethpandaops/go-eth2-client/spec/heze"
	"github.com/ethpandaops/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/go-eth2-client/spec/version"
	dynssz "github.com/pk910/dynamic-ssz"
	"github.com/pk910/dynamic-ssz/sszutils"
)

// BeaconState is a fork-agnostic beacon state containing the union of fields
// from every fork. Fields populated on a given instance depend on Version.
type BeaconState struct {
	Version                       version.DataVersion
	GenesisTime                   uint64
	GenesisValidatorsRoot         phase0.Root
	Slot                          phase0.Slot
	Fork                          *phase0.Fork
	LatestBlockHeader             *phase0.BeaconBlockHeader
	BlockRoots                    []phase0.Root
	StateRoots                    []phase0.Root
	HistoricalRoots               []phase0.Root
	ETH1Data                      *phase0.ETH1Data
	ETH1DataVotes                 []*phase0.ETH1Data
	ETH1DepositIndex              uint64
	Validators                    []*phase0.Validator
	Balances                      []phase0.Gwei
	RANDAOMixes                   []phase0.Root
	Slashings                     []phase0.Gwei
	PreviousEpochAttestations     []*phase0.PendingAttestation
	CurrentEpochAttestations      []*phase0.PendingAttestation
	PreviousEpochParticipation    []altair.ParticipationFlags
	CurrentEpochParticipation     []altair.ParticipationFlags
	JustificationBits             bitfield.Bitvector4
	PreviousJustifiedCheckpoint   *phase0.Checkpoint
	CurrentJustifiedCheckpoint    *phase0.Checkpoint
	FinalizedCheckpoint           *phase0.Checkpoint
	InactivityScores              []uint64
	CurrentSyncCommittee          *altair.SyncCommittee
	NextSyncCommittee             *altair.SyncCommittee
	LatestExecutionPayloadHeader  *ExecutionPayloadHeader
	LatestBlockHash               phase0.Hash32
	NextWithdrawalIndex           capella.WithdrawalIndex
	NextWithdrawalValidatorIndex  phase0.ValidatorIndex
	HistoricalSummaries           []*capella.HistoricalSummary
	DepositRequestsStartIndex     uint64
	DepositBalanceToConsume       phase0.Gwei
	ExitBalanceToConsume          phase0.Gwei
	EarliestExitEpoch             phase0.Epoch
	ConsolidationBalanceToConsume phase0.Gwei
	EarliestConsolidationEpoch    phase0.Epoch
	PendingDeposits               []*electra.PendingDeposit
	PendingPartialWithdrawals     []*electra.PendingPartialWithdrawal
	PendingConsolidations         []*electra.PendingConsolidation
	ProposerLookahead             []phase0.ValidatorIndex
	Builders                      []*gloas.Builder
	NextWithdrawalBuilderIndex    gloas.BuilderIndex
	ExecutionPayloadAvailability  []uint8
	BuilderPendingPayments        []*gloas.BuilderPendingPayment
	BuilderPendingWithdrawals     []*gloas.BuilderPendingWithdrawal
	LatestExecutionPayloadBid     *ExecutionPayloadBid
	PayloadExpectedWithdrawals    []*capella.Withdrawal
	PTCWindow                     [][]phase0.ValidatorIndex
}

// viewType returns the fork-specific schema type pointer used as the view
// descriptor for the active Version.
func (b *BeaconState) viewType() (any, error) {
	switch b.Version {
	case version.DataVersionPhase0:
		return (*phase0.BeaconState)(nil), nil
	case version.DataVersionAltair:
		return (*altair.BeaconState)(nil), nil
	case version.DataVersionBellatrix:
		return (*bellatrix.BeaconState)(nil), nil
	case version.DataVersionCapella:
		return (*capella.BeaconState)(nil), nil
	case version.DataVersionDeneb:
		return (*deneb.BeaconState)(nil), nil
	case version.DataVersionElectra:
		return (*electra.BeaconState)(nil), nil
	case version.DataVersionFulu:
		return (*fulu.BeaconState)(nil), nil
	case version.DataVersionGloas:
		return (*gloas.BeaconState)(nil), nil
	case version.DataVersionHeze:
		return (*heze.BeaconState)(nil), nil
	default:
		return nil, fmt.Errorf("BeaconState: unsupported version %d", b.Version)
	}
}

// MarshalSSZDyn marshals the state using the view that matches Version.
func (b *BeaconState) MarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) ([]byte, error) {
	view, err := b.viewType()
	if err != nil {
		return nil, err
	}
	m, ok := any(b).(sszutils.DynamicViewMarshaler)
	if !ok {
		return nil, errors.New("BeaconState: generated SSZ code missing")
	}
	fn := m.MarshalSSZDynView(view)
	if fn == nil {
		return nil, fmt.Errorf("BeaconState: no view marshaler for version %d", b.Version)
	}

	return fn(ds, buf)
}

// SizeSSZDyn returns the SSZ size of the state for the active Version.
func (b *BeaconState) SizeSSZDyn(ds sszutils.DynamicSpecs) int {
	view, err := b.viewType()
	if err != nil {
		return 0
	}
	s, ok := any(b).(sszutils.DynamicViewSizer)
	if !ok {
		return 0
	}
	fn := s.SizeSSZDynView(view)
	if fn == nil {
		return 0
	}

	return fn(ds)
}

// UnmarshalSSZDyn decodes the state into the view that matches Version.
func (b *BeaconState) UnmarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) error {
	view, err := b.viewType()
	if err != nil {
		return err
	}
	u, ok := any(b).(sszutils.DynamicViewUnmarshaler)
	if !ok {
		return errors.New("BeaconState: generated SSZ code missing")
	}
	fn := u.UnmarshalSSZDynView(view)
	if fn == nil {
		return fmt.Errorf("BeaconState: no view unmarshaler for version %d", b.Version)
	}

	if err := fn(ds, buf); err != nil {
		return err
	}

	b.populateVersion(b.Version)

	return nil
}

// populateVersion sets Version and propagates it to any nested versionable
// children allocated by the SSZ unmarshal.
func (b *BeaconState) populateVersion(v version.DataVersion) {
	b.Version = v

	if b.LatestExecutionPayloadHeader != nil {
		b.LatestExecutionPayloadHeader.populateVersion(v)
	}

	if b.LatestExecutionPayloadBid != nil {
		b.LatestExecutionPayloadBid.populateVersion(v)
	}
}

// ToView returns a fresh fork-specific BeaconState populated with b's fields,
// recursing into LatestExecutionPayloadHeader/LatestExecutionPayloadBid via
// their ToView.
//
//nolint:gocyclo // each version's struct literal is mechanical; one switch per fork
func (b *BeaconState) ToView() (any, error) {
	switch b.Version {
	case version.DataVersionPhase0:
		return &phase0.BeaconState{
			GenesisTime:                 b.GenesisTime,
			GenesisValidatorsRoot:       b.GenesisValidatorsRoot,
			Slot:                        b.Slot,
			Fork:                        b.Fork,
			LatestBlockHeader:           b.LatestBlockHeader,
			BlockRoots:                  b.BlockRoots,
			StateRoots:                  b.StateRoots,
			HistoricalRoots:             b.HistoricalRoots,
			ETH1Data:                    b.ETH1Data,
			ETH1DataVotes:               b.ETH1DataVotes,
			ETH1DepositIndex:            b.ETH1DepositIndex,
			Validators:                  b.Validators,
			Balances:                    b.Balances,
			RANDAOMixes:                 b.RANDAOMixes,
			Slashings:                   b.Slashings,
			PreviousEpochAttestations:   b.PreviousEpochAttestations,
			CurrentEpochAttestations:    b.CurrentEpochAttestations,
			JustificationBits:           b.JustificationBits,
			PreviousJustifiedCheckpoint: b.PreviousJustifiedCheckpoint,
			CurrentJustifiedCheckpoint:  b.CurrentJustifiedCheckpoint,
			FinalizedCheckpoint:         b.FinalizedCheckpoint,
		}, nil
	case version.DataVersionAltair:
		return &altair.BeaconState{
			GenesisTime:                 b.GenesisTime,
			GenesisValidatorsRoot:       b.GenesisValidatorsRoot,
			Slot:                        b.Slot,
			Fork:                        b.Fork,
			LatestBlockHeader:           b.LatestBlockHeader,
			BlockRoots:                  b.BlockRoots,
			StateRoots:                  b.StateRoots,
			HistoricalRoots:             b.HistoricalRoots,
			ETH1Data:                    b.ETH1Data,
			ETH1DataVotes:               b.ETH1DataVotes,
			ETH1DepositIndex:            b.ETH1DepositIndex,
			Validators:                  b.Validators,
			Balances:                    b.Balances,
			RANDAOMixes:                 b.RANDAOMixes,
			Slashings:                   b.Slashings,
			PreviousEpochParticipation:  b.PreviousEpochParticipation,
			CurrentEpochParticipation:   b.CurrentEpochParticipation,
			JustificationBits:           b.JustificationBits,
			PreviousJustifiedCheckpoint: b.PreviousJustifiedCheckpoint,
			CurrentJustifiedCheckpoint:  b.CurrentJustifiedCheckpoint,
			FinalizedCheckpoint:         b.FinalizedCheckpoint,
			InactivityScores:            b.InactivityScores,
			CurrentSyncCommittee:        b.CurrentSyncCommittee,
			NextSyncCommittee:           b.NextSyncCommittee,
		}, nil
	case version.DataVersionBellatrix:
		eph, err := toViewPtr[*bellatrix.ExecutionPayloadHeader](b.LatestExecutionPayloadHeader, "BeaconState.LatestExecutionPayloadHeader")
		if err != nil {
			return nil, err
		}

		return &bellatrix.BeaconState{
			GenesisTime:                  b.GenesisTime,
			GenesisValidatorsRoot:        b.GenesisValidatorsRoot,
			Slot:                         b.Slot,
			Fork:                         b.Fork,
			LatestBlockHeader:            b.LatestBlockHeader,
			BlockRoots:                   b.BlockRoots,
			StateRoots:                   b.StateRoots,
			HistoricalRoots:              b.HistoricalRoots,
			ETH1Data:                     b.ETH1Data,
			ETH1DataVotes:                b.ETH1DataVotes,
			ETH1DepositIndex:             b.ETH1DepositIndex,
			Validators:                   b.Validators,
			Balances:                     b.Balances,
			RANDAOMixes:                  b.RANDAOMixes,
			Slashings:                    b.Slashings,
			PreviousEpochParticipation:   b.PreviousEpochParticipation,
			CurrentEpochParticipation:    b.CurrentEpochParticipation,
			JustificationBits:            b.JustificationBits,
			PreviousJustifiedCheckpoint:  b.PreviousJustifiedCheckpoint,
			CurrentJustifiedCheckpoint:   b.CurrentJustifiedCheckpoint,
			FinalizedCheckpoint:          b.FinalizedCheckpoint,
			InactivityScores:             b.InactivityScores,
			CurrentSyncCommittee:         b.CurrentSyncCommittee,
			NextSyncCommittee:            b.NextSyncCommittee,
			LatestExecutionPayloadHeader: eph,
		}, nil
	case version.DataVersionCapella:
		eph, err := toViewPtr[*capella.ExecutionPayloadHeader](b.LatestExecutionPayloadHeader, "BeaconState.LatestExecutionPayloadHeader")
		if err != nil {
			return nil, err
		}

		return &capella.BeaconState{
			GenesisTime:                  b.GenesisTime,
			GenesisValidatorsRoot:        b.GenesisValidatorsRoot,
			Slot:                         b.Slot,
			Fork:                         b.Fork,
			LatestBlockHeader:            b.LatestBlockHeader,
			BlockRoots:                   b.BlockRoots,
			StateRoots:                   b.StateRoots,
			HistoricalRoots:              b.HistoricalRoots,
			ETH1Data:                     b.ETH1Data,
			ETH1DataVotes:                b.ETH1DataVotes,
			ETH1DepositIndex:             b.ETH1DepositIndex,
			Validators:                   b.Validators,
			Balances:                     b.Balances,
			RANDAOMixes:                  b.RANDAOMixes,
			Slashings:                    b.Slashings,
			PreviousEpochParticipation:   b.PreviousEpochParticipation,
			CurrentEpochParticipation:    b.CurrentEpochParticipation,
			JustificationBits:            b.JustificationBits,
			PreviousJustifiedCheckpoint:  b.PreviousJustifiedCheckpoint,
			CurrentJustifiedCheckpoint:   b.CurrentJustifiedCheckpoint,
			FinalizedCheckpoint:          b.FinalizedCheckpoint,
			InactivityScores:             b.InactivityScores,
			CurrentSyncCommittee:         b.CurrentSyncCommittee,
			NextSyncCommittee:            b.NextSyncCommittee,
			LatestExecutionPayloadHeader: eph,
			NextWithdrawalIndex:          b.NextWithdrawalIndex,
			NextWithdrawalValidatorIndex: b.NextWithdrawalValidatorIndex,
			HistoricalSummaries:          b.HistoricalSummaries,
		}, nil
	case version.DataVersionDeneb:
		eph, err := toViewPtr[*deneb.ExecutionPayloadHeader](b.LatestExecutionPayloadHeader, "BeaconState.LatestExecutionPayloadHeader")
		if err != nil {
			return nil, err
		}

		return &deneb.BeaconState{
			GenesisTime:                  b.GenesisTime,
			GenesisValidatorsRoot:        b.GenesisValidatorsRoot,
			Slot:                         b.Slot,
			Fork:                         b.Fork,
			LatestBlockHeader:            b.LatestBlockHeader,
			BlockRoots:                   b.BlockRoots,
			StateRoots:                   b.StateRoots,
			HistoricalRoots:              b.HistoricalRoots,
			ETH1Data:                     b.ETH1Data,
			ETH1DataVotes:                b.ETH1DataVotes,
			ETH1DepositIndex:             b.ETH1DepositIndex,
			Validators:                   b.Validators,
			Balances:                     b.Balances,
			RANDAOMixes:                  b.RANDAOMixes,
			Slashings:                    b.Slashings,
			PreviousEpochParticipation:   b.PreviousEpochParticipation,
			CurrentEpochParticipation:    b.CurrentEpochParticipation,
			JustificationBits:            b.JustificationBits,
			PreviousJustifiedCheckpoint:  b.PreviousJustifiedCheckpoint,
			CurrentJustifiedCheckpoint:   b.CurrentJustifiedCheckpoint,
			FinalizedCheckpoint:          b.FinalizedCheckpoint,
			InactivityScores:             b.InactivityScores,
			CurrentSyncCommittee:         b.CurrentSyncCommittee,
			NextSyncCommittee:            b.NextSyncCommittee,
			LatestExecutionPayloadHeader: eph,
			NextWithdrawalIndex:          b.NextWithdrawalIndex,
			NextWithdrawalValidatorIndex: b.NextWithdrawalValidatorIndex,
			HistoricalSummaries:          b.HistoricalSummaries,
		}, nil
	case version.DataVersionElectra:
		eph, err := toViewPtr[*deneb.ExecutionPayloadHeader](b.LatestExecutionPayloadHeader, "BeaconState.LatestExecutionPayloadHeader")
		if err != nil {
			return nil, err
		}

		return &electra.BeaconState{
			GenesisTime:                   b.GenesisTime,
			GenesisValidatorsRoot:         b.GenesisValidatorsRoot,
			Slot:                          b.Slot,
			Fork:                          b.Fork,
			LatestBlockHeader:             b.LatestBlockHeader,
			BlockRoots:                    b.BlockRoots,
			StateRoots:                    b.StateRoots,
			HistoricalRoots:               b.HistoricalRoots,
			ETH1Data:                      b.ETH1Data,
			ETH1DataVotes:                 b.ETH1DataVotes,
			ETH1DepositIndex:              b.ETH1DepositIndex,
			Validators:                    b.Validators,
			Balances:                      b.Balances,
			RANDAOMixes:                   b.RANDAOMixes,
			Slashings:                     b.Slashings,
			PreviousEpochParticipation:    b.PreviousEpochParticipation,
			CurrentEpochParticipation:     b.CurrentEpochParticipation,
			JustificationBits:             b.JustificationBits,
			PreviousJustifiedCheckpoint:   b.PreviousJustifiedCheckpoint,
			CurrentJustifiedCheckpoint:    b.CurrentJustifiedCheckpoint,
			FinalizedCheckpoint:           b.FinalizedCheckpoint,
			InactivityScores:              b.InactivityScores,
			CurrentSyncCommittee:          b.CurrentSyncCommittee,
			NextSyncCommittee:             b.NextSyncCommittee,
			LatestExecutionPayloadHeader:  eph,
			NextWithdrawalIndex:           b.NextWithdrawalIndex,
			NextWithdrawalValidatorIndex:  b.NextWithdrawalValidatorIndex,
			HistoricalSummaries:           b.HistoricalSummaries,
			DepositRequestsStartIndex:     b.DepositRequestsStartIndex,
			DepositBalanceToConsume:       b.DepositBalanceToConsume,
			ExitBalanceToConsume:          b.ExitBalanceToConsume,
			EarliestExitEpoch:             b.EarliestExitEpoch,
			ConsolidationBalanceToConsume: b.ConsolidationBalanceToConsume,
			EarliestConsolidationEpoch:    b.EarliestConsolidationEpoch,
			PendingDeposits:               b.PendingDeposits,
			PendingPartialWithdrawals:     b.PendingPartialWithdrawals,
			PendingConsolidations:         b.PendingConsolidations,
		}, nil
	case version.DataVersionFulu:
		eph, err := toViewPtr[*deneb.ExecutionPayloadHeader](b.LatestExecutionPayloadHeader, "BeaconState.LatestExecutionPayloadHeader")
		if err != nil {
			return nil, err
		}

		return &fulu.BeaconState{
			GenesisTime:                   b.GenesisTime,
			GenesisValidatorsRoot:         b.GenesisValidatorsRoot,
			Slot:                          b.Slot,
			Fork:                          b.Fork,
			LatestBlockHeader:             b.LatestBlockHeader,
			BlockRoots:                    b.BlockRoots,
			StateRoots:                    b.StateRoots,
			HistoricalRoots:               b.HistoricalRoots,
			ETH1Data:                      b.ETH1Data,
			ETH1DataVotes:                 b.ETH1DataVotes,
			ETH1DepositIndex:              b.ETH1DepositIndex,
			Validators:                    b.Validators,
			Balances:                      b.Balances,
			RANDAOMixes:                   b.RANDAOMixes,
			Slashings:                     b.Slashings,
			PreviousEpochParticipation:    b.PreviousEpochParticipation,
			CurrentEpochParticipation:     b.CurrentEpochParticipation,
			JustificationBits:             b.JustificationBits,
			PreviousJustifiedCheckpoint:   b.PreviousJustifiedCheckpoint,
			CurrentJustifiedCheckpoint:    b.CurrentJustifiedCheckpoint,
			FinalizedCheckpoint:           b.FinalizedCheckpoint,
			InactivityScores:              b.InactivityScores,
			CurrentSyncCommittee:          b.CurrentSyncCommittee,
			NextSyncCommittee:             b.NextSyncCommittee,
			LatestExecutionPayloadHeader:  eph,
			NextWithdrawalIndex:           b.NextWithdrawalIndex,
			NextWithdrawalValidatorIndex:  b.NextWithdrawalValidatorIndex,
			HistoricalSummaries:           b.HistoricalSummaries,
			DepositRequestsStartIndex:     b.DepositRequestsStartIndex,
			DepositBalanceToConsume:       b.DepositBalanceToConsume,
			ExitBalanceToConsume:          b.ExitBalanceToConsume,
			EarliestExitEpoch:             b.EarliestExitEpoch,
			ConsolidationBalanceToConsume: b.ConsolidationBalanceToConsume,
			EarliestConsolidationEpoch:    b.EarliestConsolidationEpoch,
			PendingDeposits:               b.PendingDeposits,
			PendingPartialWithdrawals:     b.PendingPartialWithdrawals,
			PendingConsolidations:         b.PendingConsolidations,
			ProposerLookahead:             b.ProposerLookahead,
		}, nil
	case version.DataVersionGloas:
		bid, err := toViewPtr[*gloas.ExecutionPayloadBid](b.LatestExecutionPayloadBid, "BeaconState.LatestExecutionPayloadBid")
		if err != nil {
			return nil, err
		}

		return &gloas.BeaconState{
			GenesisTime:                   b.GenesisTime,
			GenesisValidatorsRoot:         b.GenesisValidatorsRoot,
			Slot:                          b.Slot,
			Fork:                          b.Fork,
			LatestBlockHeader:             b.LatestBlockHeader,
			BlockRoots:                    b.BlockRoots,
			StateRoots:                    b.StateRoots,
			HistoricalRoots:               b.HistoricalRoots,
			ETH1Data:                      b.ETH1Data,
			ETH1DataVotes:                 b.ETH1DataVotes,
			ETH1DepositIndex:              b.ETH1DepositIndex,
			Validators:                    b.Validators,
			Balances:                      b.Balances,
			RANDAOMixes:                   b.RANDAOMixes,
			Slashings:                     b.Slashings,
			PreviousEpochParticipation:    b.PreviousEpochParticipation,
			CurrentEpochParticipation:     b.CurrentEpochParticipation,
			JustificationBits:             b.JustificationBits,
			PreviousJustifiedCheckpoint:   b.PreviousJustifiedCheckpoint,
			CurrentJustifiedCheckpoint:    b.CurrentJustifiedCheckpoint,
			FinalizedCheckpoint:           b.FinalizedCheckpoint,
			InactivityScores:              b.InactivityScores,
			CurrentSyncCommittee:          b.CurrentSyncCommittee,
			NextSyncCommittee:             b.NextSyncCommittee,
			LatestBlockHash:               b.LatestBlockHash,
			NextWithdrawalIndex:           b.NextWithdrawalIndex,
			NextWithdrawalValidatorIndex:  b.NextWithdrawalValidatorIndex,
			HistoricalSummaries:           b.HistoricalSummaries,
			DepositRequestsStartIndex:     b.DepositRequestsStartIndex,
			DepositBalanceToConsume:       b.DepositBalanceToConsume,
			ExitBalanceToConsume:          b.ExitBalanceToConsume,
			EarliestExitEpoch:             b.EarliestExitEpoch,
			ConsolidationBalanceToConsume: b.ConsolidationBalanceToConsume,
			EarliestConsolidationEpoch:    b.EarliestConsolidationEpoch,
			PendingDeposits:               b.PendingDeposits,
			PendingPartialWithdrawals:     b.PendingPartialWithdrawals,
			PendingConsolidations:         b.PendingConsolidations,
			ProposerLookahead:             b.ProposerLookahead,
			Builders:                      b.Builders,
			NextWithdrawalBuilderIndex:    b.NextWithdrawalBuilderIndex,
			ExecutionPayloadAvailability:  b.ExecutionPayloadAvailability,
			BuilderPendingPayments:        b.BuilderPendingPayments,
			BuilderPendingWithdrawals:     b.BuilderPendingWithdrawals,
			LatestExecutionPayloadBid:     bid,
			PayloadExpectedWithdrawals:    b.PayloadExpectedWithdrawals,
			PTCWindow:                     b.PTCWindow,
		}, nil
	case version.DataVersionHeze:
		bid, err := toViewPtr[*heze.ExecutionPayloadBid](b.LatestExecutionPayloadBid, "BeaconState.LatestExecutionPayloadBid")
		if err != nil {
			return nil, err
		}

		return &heze.BeaconState{
			GenesisTime:                   b.GenesisTime,
			GenesisValidatorsRoot:         b.GenesisValidatorsRoot,
			Slot:                          b.Slot,
			Fork:                          b.Fork,
			LatestBlockHeader:             b.LatestBlockHeader,
			BlockRoots:                    b.BlockRoots,
			StateRoots:                    b.StateRoots,
			HistoricalRoots:               b.HistoricalRoots,
			ETH1Data:                      b.ETH1Data,
			ETH1DataVotes:                 b.ETH1DataVotes,
			ETH1DepositIndex:              b.ETH1DepositIndex,
			Validators:                    b.Validators,
			Balances:                      b.Balances,
			RANDAOMixes:                   b.RANDAOMixes,
			Slashings:                     b.Slashings,
			PreviousEpochParticipation:    b.PreviousEpochParticipation,
			CurrentEpochParticipation:     b.CurrentEpochParticipation,
			JustificationBits:             b.JustificationBits,
			PreviousJustifiedCheckpoint:   b.PreviousJustifiedCheckpoint,
			CurrentJustifiedCheckpoint:    b.CurrentJustifiedCheckpoint,
			FinalizedCheckpoint:           b.FinalizedCheckpoint,
			InactivityScores:              b.InactivityScores,
			CurrentSyncCommittee:          b.CurrentSyncCommittee,
			NextSyncCommittee:             b.NextSyncCommittee,
			LatestBlockHash:               b.LatestBlockHash,
			NextWithdrawalIndex:           b.NextWithdrawalIndex,
			NextWithdrawalValidatorIndex:  b.NextWithdrawalValidatorIndex,
			HistoricalSummaries:           b.HistoricalSummaries,
			DepositRequestsStartIndex:     b.DepositRequestsStartIndex,
			DepositBalanceToConsume:       b.DepositBalanceToConsume,
			ExitBalanceToConsume:          b.ExitBalanceToConsume,
			EarliestExitEpoch:             b.EarliestExitEpoch,
			ConsolidationBalanceToConsume: b.ConsolidationBalanceToConsume,
			EarliestConsolidationEpoch:    b.EarliestConsolidationEpoch,
			PendingDeposits:               b.PendingDeposits,
			PendingPartialWithdrawals:     b.PendingPartialWithdrawals,
			PendingConsolidations:         b.PendingConsolidations,
			ProposerLookahead:             b.ProposerLookahead,
			Builders:                      b.Builders,
			NextWithdrawalBuilderIndex:    b.NextWithdrawalBuilderIndex,
			ExecutionPayloadAvailability:  b.ExecutionPayloadAvailability,
			BuilderPendingPayments:        b.BuilderPendingPayments,
			BuilderPendingWithdrawals:     b.BuilderPendingWithdrawals,
			LatestExecutionPayloadBid:     bid,
			PayloadExpectedWithdrawals:    b.PayloadExpectedWithdrawals,
			PTCWindow:                     b.PTCWindow,
		}, nil
	default:
		return nil, fmt.Errorf("BeaconState: unsupported version %d", b.Version)
	}
}

// FromView populates b from a fork-specific BeaconState.
//
//nolint:gocyclo // each version branch is mechanical field copy
func (b *BeaconState) FromView(view any) error {
	switch v := view.(type) {
	case *phase0.BeaconState:
		if b.Version == version.DataVersionUnknown {
			b.Version = version.DataVersionPhase0
		}

		b.copyPhase0Common(
			v.GenesisTime, v.GenesisValidatorsRoot, v.Slot,
			v.Fork, v.LatestBlockHeader,
			v.BlockRoots, v.StateRoots, v.HistoricalRoots,
			v.ETH1Data, v.ETH1DataVotes, v.ETH1DepositIndex,
			v.Validators, v.Balances, v.RANDAOMixes, v.Slashings,
			v.JustificationBits,
			v.PreviousJustifiedCheckpoint, v.CurrentJustifiedCheckpoint, v.FinalizedCheckpoint,
		)
		b.PreviousEpochAttestations = v.PreviousEpochAttestations
		b.CurrentEpochAttestations = v.CurrentEpochAttestations

		return nil
	case *altair.BeaconState:
		if b.Version == version.DataVersionUnknown {
			b.Version = version.DataVersionAltair
		}

		b.copyPhase0Common(
			v.GenesisTime, v.GenesisValidatorsRoot, v.Slot,
			v.Fork, v.LatestBlockHeader,
			v.BlockRoots, v.StateRoots, v.HistoricalRoots,
			v.ETH1Data, v.ETH1DataVotes, v.ETH1DepositIndex,
			v.Validators, v.Balances, v.RANDAOMixes, v.Slashings,
			v.JustificationBits,
			v.PreviousJustifiedCheckpoint, v.CurrentJustifiedCheckpoint, v.FinalizedCheckpoint,
		)
		b.copyAltairAdditions(
			v.PreviousEpochParticipation, v.CurrentEpochParticipation,
			v.InactivityScores, v.CurrentSyncCommittee, v.NextSyncCommittee,
		)

		return nil
	case *bellatrix.BeaconState:
		if b.Version == version.DataVersionUnknown {
			b.Version = version.DataVersionBellatrix
		}

		b.copyPhase0Common(
			v.GenesisTime, v.GenesisValidatorsRoot, v.Slot,
			v.Fork, v.LatestBlockHeader,
			v.BlockRoots, v.StateRoots, v.HistoricalRoots,
			v.ETH1Data, v.ETH1DataVotes, v.ETH1DepositIndex,
			v.Validators, v.Balances, v.RANDAOMixes, v.Slashings,
			v.JustificationBits,
			v.PreviousJustifiedCheckpoint, v.CurrentJustifiedCheckpoint, v.FinalizedCheckpoint,
		)
		b.copyAltairAdditions(
			v.PreviousEpochParticipation, v.CurrentEpochParticipation,
			v.InactivityScores, v.CurrentSyncCommittee, v.NextSyncCommittee,
		)

		return b.fromExecutionPayloadHeader(v.LatestExecutionPayloadHeader)
	case *capella.BeaconState:
		if b.Version == version.DataVersionUnknown {
			b.Version = version.DataVersionCapella
		}

		b.copyPhase0Common(
			v.GenesisTime, v.GenesisValidatorsRoot, v.Slot,
			v.Fork, v.LatestBlockHeader,
			v.BlockRoots, v.StateRoots, v.HistoricalRoots,
			v.ETH1Data, v.ETH1DataVotes, v.ETH1DepositIndex,
			v.Validators, v.Balances, v.RANDAOMixes, v.Slashings,
			v.JustificationBits,
			v.PreviousJustifiedCheckpoint, v.CurrentJustifiedCheckpoint, v.FinalizedCheckpoint,
		)
		b.copyAltairAdditions(
			v.PreviousEpochParticipation, v.CurrentEpochParticipation,
			v.InactivityScores, v.CurrentSyncCommittee, v.NextSyncCommittee,
		)
		b.NextWithdrawalIndex = v.NextWithdrawalIndex
		b.NextWithdrawalValidatorIndex = v.NextWithdrawalValidatorIndex
		b.HistoricalSummaries = v.HistoricalSummaries

		return b.fromExecutionPayloadHeader(v.LatestExecutionPayloadHeader)
	case *deneb.BeaconState:
		if b.Version == version.DataVersionUnknown {
			b.Version = version.DataVersionDeneb
		}

		b.copyPhase0Common(
			v.GenesisTime, v.GenesisValidatorsRoot, v.Slot,
			v.Fork, v.LatestBlockHeader,
			v.BlockRoots, v.StateRoots, v.HistoricalRoots,
			v.ETH1Data, v.ETH1DataVotes, v.ETH1DepositIndex,
			v.Validators, v.Balances, v.RANDAOMixes, v.Slashings,
			v.JustificationBits,
			v.PreviousJustifiedCheckpoint, v.CurrentJustifiedCheckpoint, v.FinalizedCheckpoint,
		)
		b.copyAltairAdditions(
			v.PreviousEpochParticipation, v.CurrentEpochParticipation,
			v.InactivityScores, v.CurrentSyncCommittee, v.NextSyncCommittee,
		)
		b.NextWithdrawalIndex = v.NextWithdrawalIndex
		b.NextWithdrawalValidatorIndex = v.NextWithdrawalValidatorIndex
		b.HistoricalSummaries = v.HistoricalSummaries

		return b.fromExecutionPayloadHeader(v.LatestExecutionPayloadHeader)
	case *electra.BeaconState:
		if b.Version == version.DataVersionUnknown {
			b.Version = version.DataVersionElectra
		}

		b.copyPhase0Common(
			v.GenesisTime, v.GenesisValidatorsRoot, v.Slot,
			v.Fork, v.LatestBlockHeader,
			v.BlockRoots, v.StateRoots, v.HistoricalRoots,
			v.ETH1Data, v.ETH1DataVotes, v.ETH1DepositIndex,
			v.Validators, v.Balances, v.RANDAOMixes, v.Slashings,
			v.JustificationBits,
			v.PreviousJustifiedCheckpoint, v.CurrentJustifiedCheckpoint, v.FinalizedCheckpoint,
		)
		b.copyAltairAdditions(
			v.PreviousEpochParticipation, v.CurrentEpochParticipation,
			v.InactivityScores, v.CurrentSyncCommittee, v.NextSyncCommittee,
		)
		b.copyCapellaAdditions(v.NextWithdrawalIndex, v.NextWithdrawalValidatorIndex, v.HistoricalSummaries)
		b.copyElectraAdditions(
			v.DepositRequestsStartIndex, v.DepositBalanceToConsume,
			v.ExitBalanceToConsume, v.EarliestExitEpoch,
			v.ConsolidationBalanceToConsume, v.EarliestConsolidationEpoch,
			v.PendingDeposits, v.PendingPartialWithdrawals, v.PendingConsolidations,
		)

		return b.fromExecutionPayloadHeader(v.LatestExecutionPayloadHeader)
	case *fulu.BeaconState:
		if b.Version == version.DataVersionUnknown {
			b.Version = version.DataVersionFulu
		}

		b.copyPhase0Common(
			v.GenesisTime, v.GenesisValidatorsRoot, v.Slot,
			v.Fork, v.LatestBlockHeader,
			v.BlockRoots, v.StateRoots, v.HistoricalRoots,
			v.ETH1Data, v.ETH1DataVotes, v.ETH1DepositIndex,
			v.Validators, v.Balances, v.RANDAOMixes, v.Slashings,
			v.JustificationBits,
			v.PreviousJustifiedCheckpoint, v.CurrentJustifiedCheckpoint, v.FinalizedCheckpoint,
		)
		b.copyAltairAdditions(
			v.PreviousEpochParticipation, v.CurrentEpochParticipation,
			v.InactivityScores, v.CurrentSyncCommittee, v.NextSyncCommittee,
		)
		b.copyCapellaAdditions(v.NextWithdrawalIndex, v.NextWithdrawalValidatorIndex, v.HistoricalSummaries)
		b.copyElectraAdditions(
			v.DepositRequestsStartIndex, v.DepositBalanceToConsume,
			v.ExitBalanceToConsume, v.EarliestExitEpoch,
			v.ConsolidationBalanceToConsume, v.EarliestConsolidationEpoch,
			v.PendingDeposits, v.PendingPartialWithdrawals, v.PendingConsolidations,
		)
		b.ProposerLookahead = v.ProposerLookahead

		return b.fromExecutionPayloadHeader(v.LatestExecutionPayloadHeader)
	case *gloas.BeaconState:
		if b.Version == version.DataVersionUnknown {
			b.Version = version.DataVersionGloas
		}

		b.copyPhase0Common(
			v.GenesisTime, v.GenesisValidatorsRoot, v.Slot,
			v.Fork, v.LatestBlockHeader,
			v.BlockRoots, v.StateRoots, v.HistoricalRoots,
			v.ETH1Data, v.ETH1DataVotes, v.ETH1DepositIndex,
			v.Validators, v.Balances, v.RANDAOMixes, v.Slashings,
			v.JustificationBits,
			v.PreviousJustifiedCheckpoint, v.CurrentJustifiedCheckpoint, v.FinalizedCheckpoint,
		)
		b.copyAltairAdditions(
			v.PreviousEpochParticipation, v.CurrentEpochParticipation,
			v.InactivityScores, v.CurrentSyncCommittee, v.NextSyncCommittee,
		)
		b.copyCapellaAdditions(v.NextWithdrawalIndex, v.NextWithdrawalValidatorIndex, v.HistoricalSummaries)
		b.copyElectraAdditions(
			v.DepositRequestsStartIndex, v.DepositBalanceToConsume,
			v.ExitBalanceToConsume, v.EarliestExitEpoch,
			v.ConsolidationBalanceToConsume, v.EarliestConsolidationEpoch,
			v.PendingDeposits, v.PendingPartialWithdrawals, v.PendingConsolidations,
		)
		b.ProposerLookahead = v.ProposerLookahead
		b.LatestBlockHash = v.LatestBlockHash
		b.Builders = v.Builders
		b.NextWithdrawalBuilderIndex = v.NextWithdrawalBuilderIndex
		b.ExecutionPayloadAvailability = v.ExecutionPayloadAvailability
		b.BuilderPendingPayments = v.BuilderPendingPayments
		b.BuilderPendingWithdrawals = v.BuilderPendingWithdrawals
		b.PayloadExpectedWithdrawals = v.PayloadExpectedWithdrawals
		b.PTCWindow = v.PTCWindow
		b.LatestExecutionPayloadHeader = nil

		return b.fromExecutionPayloadBid(v.LatestExecutionPayloadBid)
	case *heze.BeaconState:
		if b.Version == version.DataVersionUnknown {
			b.Version = version.DataVersionHeze
		}

		b.copyPhase0Common(
			v.GenesisTime, v.GenesisValidatorsRoot, v.Slot,
			v.Fork, v.LatestBlockHeader,
			v.BlockRoots, v.StateRoots, v.HistoricalRoots,
			v.ETH1Data, v.ETH1DataVotes, v.ETH1DepositIndex,
			v.Validators, v.Balances, v.RANDAOMixes, v.Slashings,
			v.JustificationBits,
			v.PreviousJustifiedCheckpoint, v.CurrentJustifiedCheckpoint, v.FinalizedCheckpoint,
		)
		b.copyAltairAdditions(
			v.PreviousEpochParticipation, v.CurrentEpochParticipation,
			v.InactivityScores, v.CurrentSyncCommittee, v.NextSyncCommittee,
		)
		b.copyCapellaAdditions(v.NextWithdrawalIndex, v.NextWithdrawalValidatorIndex, v.HistoricalSummaries)
		b.copyElectraAdditions(
			v.DepositRequestsStartIndex, v.DepositBalanceToConsume,
			v.ExitBalanceToConsume, v.EarliestExitEpoch,
			v.ConsolidationBalanceToConsume, v.EarliestConsolidationEpoch,
			v.PendingDeposits, v.PendingPartialWithdrawals, v.PendingConsolidations,
		)
		b.ProposerLookahead = v.ProposerLookahead
		b.LatestBlockHash = v.LatestBlockHash
		b.Builders = v.Builders
		b.NextWithdrawalBuilderIndex = v.NextWithdrawalBuilderIndex
		b.ExecutionPayloadAvailability = v.ExecutionPayloadAvailability
		b.BuilderPendingPayments = v.BuilderPendingPayments
		b.BuilderPendingWithdrawals = v.BuilderPendingWithdrawals
		b.PayloadExpectedWithdrawals = v.PayloadExpectedWithdrawals
		b.PTCWindow = v.PTCWindow
		b.LatestExecutionPayloadHeader = nil

		return b.fromExecutionPayloadBid(v.LatestExecutionPayloadBid)
	default:
		return fmt.Errorf("BeaconState: unsupported view type %T", view)
	}
}

//nolint:revive // helper takes many params to keep the per-version branches terse
func (b *BeaconState) copyPhase0Common(
	genesisTime uint64, genesisValidatorsRoot phase0.Root, slot phase0.Slot,
	fork *phase0.Fork, latestBlockHeader *phase0.BeaconBlockHeader,
	blockRoots, stateRoots, historicalRoots []phase0.Root,
	eth1Data *phase0.ETH1Data, eth1DataVotes []*phase0.ETH1Data, eth1DepositIndex uint64,
	validators []*phase0.Validator, balances []phase0.Gwei, randaoMixes []phase0.Root,
	slashings []phase0.Gwei, justificationBits bitfield.Bitvector4,
	previousJustified, currentJustified, finalized *phase0.Checkpoint,
) {
	b.GenesisTime = genesisTime
	b.GenesisValidatorsRoot = genesisValidatorsRoot
	b.Slot = slot
	b.Fork = fork
	b.LatestBlockHeader = latestBlockHeader
	b.BlockRoots = blockRoots
	b.StateRoots = stateRoots
	b.HistoricalRoots = historicalRoots
	b.ETH1Data = eth1Data
	b.ETH1DataVotes = eth1DataVotes
	b.ETH1DepositIndex = eth1DepositIndex
	b.Validators = validators
	b.Balances = balances
	b.RANDAOMixes = randaoMixes
	b.Slashings = slashings
	b.JustificationBits = justificationBits
	b.PreviousJustifiedCheckpoint = previousJustified
	b.CurrentJustifiedCheckpoint = currentJustified
	b.FinalizedCheckpoint = finalized
}

func (b *BeaconState) copyAltairAdditions(
	prevPart, currPart []altair.ParticipationFlags,
	inactivityScores []uint64,
	currentSync, nextSync *altair.SyncCommittee,
) {
	b.PreviousEpochParticipation = prevPart
	b.CurrentEpochParticipation = currPart
	b.InactivityScores = inactivityScores
	b.CurrentSyncCommittee = currentSync
	b.NextSyncCommittee = nextSync
}

func (b *BeaconState) copyCapellaAdditions(
	nextWithdrawalIndex capella.WithdrawalIndex,
	nextWithdrawalValidatorIndex phase0.ValidatorIndex,
	historicalSummaries []*capella.HistoricalSummary,
) {
	b.NextWithdrawalIndex = nextWithdrawalIndex
	b.NextWithdrawalValidatorIndex = nextWithdrawalValidatorIndex
	b.HistoricalSummaries = historicalSummaries
}

//nolint:revive // helper takes many params to keep the per-version branches terse
func (b *BeaconState) copyElectraAdditions(
	depositRequestsStartIndex uint64,
	depositBalanceToConsume, exitBalanceToConsume phase0.Gwei,
	earliestExitEpoch phase0.Epoch,
	consolidationBalanceToConsume phase0.Gwei,
	earliestConsolidationEpoch phase0.Epoch,
	pendingDeposits []*electra.PendingDeposit,
	pendingPartialWithdrawals []*electra.PendingPartialWithdrawal,
	pendingConsolidations []*electra.PendingConsolidation,
) {
	b.DepositRequestsStartIndex = depositRequestsStartIndex
	b.DepositBalanceToConsume = depositBalanceToConsume
	b.ExitBalanceToConsume = exitBalanceToConsume
	b.EarliestExitEpoch = earliestExitEpoch
	b.ConsolidationBalanceToConsume = consolidationBalanceToConsume
	b.EarliestConsolidationEpoch = earliestConsolidationEpoch
	b.PendingDeposits = pendingDeposits
	b.PendingPartialWithdrawals = pendingPartialWithdrawals
	b.PendingConsolidations = pendingConsolidations
}

func (b *BeaconState) fromExecutionPayloadHeader(view any) error {
	rv := reflect.ValueOf(view)
	if !rv.IsValid() || (rv.Kind() == reflect.Ptr && rv.IsNil()) {
		b.LatestExecutionPayloadHeader = nil

		return nil
	}

	if b.LatestExecutionPayloadHeader == nil {
		b.LatestExecutionPayloadHeader = &ExecutionPayloadHeader{Version: b.Version}
	}

	return b.LatestExecutionPayloadHeader.FromView(view)
}

func (b *BeaconState) fromExecutionPayloadBid(view any) error {
	rv := reflect.ValueOf(view)
	if !rv.IsValid() || (rv.Kind() == reflect.Ptr && rv.IsNil()) {
		b.LatestExecutionPayloadBid = nil

		return nil
	}

	if b.LatestExecutionPayloadBid == nil {
		b.LatestExecutionPayloadBid = &ExecutionPayloadBid{Version: b.Version}
	}

	return b.LatestExecutionPayloadBid.FromView(view)
}

// HashTreeRootWithDyn computes the SSZ hash tree root using the active Version's view.
func (b *BeaconState) HashTreeRootWithDyn(ds sszutils.DynamicSpecs, hh sszutils.HashWalker) error {
	view, err := b.viewType()
	if err != nil {
		return err
	}
	h, ok := any(b).(sszutils.DynamicViewHashRoot)
	if !ok {
		return errors.New("BeaconState: generated SSZ code missing")
	}
	fn := h.HashTreeRootWithDynView(view)
	if fn == nil {
		return fmt.Errorf("BeaconState: no view hasher for version %d", b.Version)
	}

	return fn(ds, hh)
}

// MarshalSSZ implements the fastssz.Marshaler interface.
func (b *BeaconState) MarshalSSZ() ([]byte, error) {
	ds := dynssz.GetGlobalDynSsz()

	return b.MarshalSSZDyn(ds, make([]byte, 0, b.SizeSSZDyn(ds)))
}

// MarshalSSZTo implements the fastssz.Marshaler interface.
func (b *BeaconState) MarshalSSZTo(dst []byte) ([]byte, error) {
	return b.MarshalSSZDyn(dynssz.GetGlobalDynSsz(), dst)
}

// UnmarshalSSZ implements the fastssz.Unmarshaler interface.
func (b *BeaconState) UnmarshalSSZ(buf []byte) error {
	return b.UnmarshalSSZDyn(dynssz.GetGlobalDynSsz(), buf)
}

// SizeSSZ implements the fastssz.Marshaler interface.
func (b *BeaconState) SizeSSZ() int {
	return b.SizeSSZDyn(dynssz.GetGlobalDynSsz())
}

// HashTreeRoot implements the fastssz.HashRoot interface.
func (b *BeaconState) HashTreeRoot() (phase0.Root, error) {
	return dynssz.GetGlobalDynSsz().HashTreeRoot(b)
}

// HashTreeRootWith implements the fastssz.HashRoot interface.
func (b *BeaconState) HashTreeRootWith(hh sszutils.HashWalker) error {
	return b.HashTreeRootWithDyn(dynssz.GetGlobalDynSsz(), hh)
}

// MarshalJSON delegates to the per-fork BeaconState that matches Version.
func (b *BeaconState) MarshalJSON() ([]byte, error) {
	return marshalAsView(b)
}

// UnmarshalJSON delegates to the per-fork BeaconState that matches Version.
// Caller must set Version before calling.
func (b *BeaconState) UnmarshalJSON(data []byte) error {
	if err := unmarshalAsView(b, data); err != nil {
		return err
	}

	b.populateVersion(b.Version)

	return nil
}

// MarshalYAML delegates to the per-fork BeaconState that matches Version.
func (b *BeaconState) MarshalYAML() ([]byte, error) {
	return marshalAsViewYAML(b)
}

// UnmarshalYAML delegates to the per-fork BeaconState that matches Version.
// Caller must set Version before calling.
func (b *BeaconState) UnmarshalYAML(data []byte) error {
	if err := unmarshalAsViewYAML(b, data); err != nil {
		return err
	}

	b.populateVersion(b.Version)

	return nil
}
