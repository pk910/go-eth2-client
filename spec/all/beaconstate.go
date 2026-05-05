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

	return fn(ds, buf)
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
