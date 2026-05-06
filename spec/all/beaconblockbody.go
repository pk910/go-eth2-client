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

	"github.com/ethpandaops/go-eth2-client/spec/altair"
	"github.com/ethpandaops/go-eth2-client/spec/bellatrix"
	"github.com/ethpandaops/go-eth2-client/spec/capella"
	"github.com/ethpandaops/go-eth2-client/spec/deneb"
	"github.com/ethpandaops/go-eth2-client/spec/electra"
	"github.com/ethpandaops/go-eth2-client/spec/gloas"
	"github.com/ethpandaops/go-eth2-client/spec/heze"
	"github.com/ethpandaops/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/go-eth2-client/spec/version"
	dynssz "github.com/pk910/dynamic-ssz"
	"github.com/pk910/dynamic-ssz/sszutils"
)

// BeaconBlockBody is a fork-agnostic beacon block body containing the union
// of fields from every fork. Fields populated on a given instance depend on
// Version.
type BeaconBlockBody struct {
	Version                   version.DataVersion
	RANDAOReveal              phase0.BLSSignature
	ETH1Data                  *phase0.ETH1Data
	Graffiti                  [32]byte
	ProposerSlashings         []*phase0.ProposerSlashing
	AttesterSlashings         []*AttesterSlashing
	Attestations              []*Attestation
	Deposits                  []*phase0.Deposit
	VoluntaryExits            []*phase0.SignedVoluntaryExit
	SyncAggregate             *altair.SyncAggregate
	ExecutionPayload          *ExecutionPayload
	BLSToExecutionChanges     []*capella.SignedBLSToExecutionChange
	BlobKZGCommitments        []deneb.KZGCommitment
	ExecutionRequests         *electra.ExecutionRequests
	SignedExecutionPayloadBid *SignedExecutionPayloadBid
	PayloadAttestations       []*gloas.PayloadAttestation
	ParentExecutionRequests   *electra.ExecutionRequests
}

// viewType returns the fork-specific schema type pointer used as the view
// descriptor for the active Version.
func (b *BeaconBlockBody) viewType() (any, error) {
	switch b.Version {
	case version.DataVersionPhase0:
		return (*phase0.BeaconBlockBody)(nil), nil
	case version.DataVersionAltair:
		return (*altair.BeaconBlockBody)(nil), nil
	case version.DataVersionBellatrix:
		return (*bellatrix.BeaconBlockBody)(nil), nil
	case version.DataVersionCapella:
		return (*capella.BeaconBlockBody)(nil), nil
	case version.DataVersionDeneb:
		return (*deneb.BeaconBlockBody)(nil), nil
	case version.DataVersionElectra:
		return (*electra.BeaconBlockBody)(nil), nil
	case version.DataVersionGloas:
		return (*gloas.BeaconBlockBody)(nil), nil
	case version.DataVersionHeze:
		return (*heze.BeaconBlockBody)(nil), nil
	default:
		return nil, fmt.Errorf("BeaconBlockBody: unsupported version %d", b.Version)
	}
}

// MarshalSSZDyn marshals the body using the view that matches Version.
func (b *BeaconBlockBody) MarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) ([]byte, error) {
	view, err := b.viewType()
	if err != nil {
		return nil, err
	}

	m, ok := any(b).(sszutils.DynamicViewMarshaler)
	if !ok {
		return nil, errors.New("BeaconBlockBody: generated SSZ code missing")
	}

	fn := m.MarshalSSZDynView(view)
	if fn == nil {
		return nil, fmt.Errorf("BeaconBlockBody: no view marshaler for version %d", b.Version)
	}

	return fn(ds, buf)
}

// SizeSSZDyn returns the SSZ size of the body for the active Version.
func (b *BeaconBlockBody) SizeSSZDyn(ds sszutils.DynamicSpecs) int {
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

// UnmarshalSSZDyn decodes the body into the view that matches Version.
func (b *BeaconBlockBody) UnmarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) error {
	view, err := b.viewType()
	if err != nil {
		return err
	}

	u, ok := any(b).(sszutils.DynamicViewUnmarshaler)
	if !ok {
		return errors.New("BeaconBlockBody: generated SSZ code missing")
	}

	fn := u.UnmarshalSSZDynView(view)
	if fn == nil {
		return fmt.Errorf("BeaconBlockBody: no view unmarshaler for version %d", b.Version)
	}

	if err := fn(ds, buf); err != nil {
		return err
	}

	b.populateVersion(b.Version)

	return nil
}

// populateVersion sets Version and propagates it to any nested versionable
// children allocated by the SSZ unmarshal.
func (b *BeaconBlockBody) populateVersion(v version.DataVersion) {
	b.Version = v

	for _, as := range b.AttesterSlashings {
		if as != nil {
			as.populateVersion(v)
		}
	}

	for _, a := range b.Attestations {
		if a != nil {
			a.populateVersion(v)
		}
	}

	if b.ExecutionPayload != nil {
		b.ExecutionPayload.populateVersion(v)
	}

	if b.SignedExecutionPayloadBid != nil {
		b.SignedExecutionPayloadBid.populateVersion(v)
	}
}

// ToView returns a fresh fork-specific BeaconBlockBody populated with b's
// fields, recursing into nested versionable children via their ToView.
func (b *BeaconBlockBody) ToView() (any, error) {
	switch b.Version {
	case version.DataVersionPhase0:
		as, err := toViewSlice[*phase0.AttesterSlashing](b.AttesterSlashings, "BeaconBlockBody.AttesterSlashings")
		if err != nil {
			return nil, err
		}

		at, err := toViewSlice[*phase0.Attestation](b.Attestations, "BeaconBlockBody.Attestations")
		if err != nil {
			return nil, err
		}

		return &phase0.BeaconBlockBody{
			RANDAOReveal:      b.RANDAOReveal,
			ETH1Data:          b.ETH1Data,
			Graffiti:          b.Graffiti,
			ProposerSlashings: b.ProposerSlashings,
			AttesterSlashings: as,
			Attestations:      at,
			Deposits:          b.Deposits,
			VoluntaryExits:    b.VoluntaryExits,
		}, nil
	case version.DataVersionAltair:
		as, err := toViewSlice[*phase0.AttesterSlashing](b.AttesterSlashings, "BeaconBlockBody.AttesterSlashings")
		if err != nil {
			return nil, err
		}

		at, err := toViewSlice[*phase0.Attestation](b.Attestations, "BeaconBlockBody.Attestations")
		if err != nil {
			return nil, err
		}

		return &altair.BeaconBlockBody{
			RANDAOReveal:      b.RANDAOReveal,
			ETH1Data:          b.ETH1Data,
			Graffiti:          b.Graffiti,
			ProposerSlashings: b.ProposerSlashings,
			AttesterSlashings: as,
			Attestations:      at,
			Deposits:          b.Deposits,
			VoluntaryExits:    b.VoluntaryExits,
			SyncAggregate:     b.SyncAggregate,
		}, nil
	case version.DataVersionBellatrix:
		as, err := toViewSlice[*phase0.AttesterSlashing](b.AttesterSlashings, "BeaconBlockBody.AttesterSlashings")
		if err != nil {
			return nil, err
		}

		at, err := toViewSlice[*phase0.Attestation](b.Attestations, "BeaconBlockBody.Attestations")
		if err != nil {
			return nil, err
		}

		ep, err := toViewPtr[*bellatrix.ExecutionPayload](b.ExecutionPayload, "BeaconBlockBody.ExecutionPayload")
		if err != nil {
			return nil, err
		}

		return &bellatrix.BeaconBlockBody{
			RANDAOReveal:      b.RANDAOReveal,
			ETH1Data:          b.ETH1Data,
			Graffiti:          b.Graffiti,
			ProposerSlashings: b.ProposerSlashings,
			AttesterSlashings: as,
			Attestations:      at,
			Deposits:          b.Deposits,
			VoluntaryExits:    b.VoluntaryExits,
			SyncAggregate:     b.SyncAggregate,
			ExecutionPayload:  ep,
		}, nil
	case version.DataVersionCapella:
		as, err := toViewSlice[*phase0.AttesterSlashing](b.AttesterSlashings, "BeaconBlockBody.AttesterSlashings")
		if err != nil {
			return nil, err
		}

		at, err := toViewSlice[*phase0.Attestation](b.Attestations, "BeaconBlockBody.Attestations")
		if err != nil {
			return nil, err
		}

		ep, err := toViewPtr[*capella.ExecutionPayload](b.ExecutionPayload, "BeaconBlockBody.ExecutionPayload")
		if err != nil {
			return nil, err
		}

		return &capella.BeaconBlockBody{
			RANDAOReveal:          b.RANDAOReveal,
			ETH1Data:              b.ETH1Data,
			Graffiti:              b.Graffiti,
			ProposerSlashings:     b.ProposerSlashings,
			AttesterSlashings:     as,
			Attestations:          at,
			Deposits:              b.Deposits,
			VoluntaryExits:        b.VoluntaryExits,
			SyncAggregate:         b.SyncAggregate,
			ExecutionPayload:      ep,
			BLSToExecutionChanges: b.BLSToExecutionChanges,
		}, nil
	case version.DataVersionDeneb:
		as, err := toViewSlice[*phase0.AttesterSlashing](b.AttesterSlashings, "BeaconBlockBody.AttesterSlashings")
		if err != nil {
			return nil, err
		}

		at, err := toViewSlice[*phase0.Attestation](b.Attestations, "BeaconBlockBody.Attestations")
		if err != nil {
			return nil, err
		}

		ep, err := toViewPtr[*deneb.ExecutionPayload](b.ExecutionPayload, "BeaconBlockBody.ExecutionPayload")
		if err != nil {
			return nil, err
		}

		return &deneb.BeaconBlockBody{
			RANDAOReveal:          b.RANDAOReveal,
			ETH1Data:              b.ETH1Data,
			Graffiti:              b.Graffiti,
			ProposerSlashings:     b.ProposerSlashings,
			AttesterSlashings:     as,
			Attestations:          at,
			Deposits:              b.Deposits,
			VoluntaryExits:        b.VoluntaryExits,
			SyncAggregate:         b.SyncAggregate,
			ExecutionPayload:      ep,
			BLSToExecutionChanges: b.BLSToExecutionChanges,
			BlobKZGCommitments:    b.BlobKZGCommitments,
		}, nil
	case version.DataVersionElectra:
		as, err := toViewSlice[*electra.AttesterSlashing](b.AttesterSlashings, "BeaconBlockBody.AttesterSlashings")
		if err != nil {
			return nil, err
		}

		at, err := toViewSlice[*electra.Attestation](b.Attestations, "BeaconBlockBody.Attestations")
		if err != nil {
			return nil, err
		}

		ep, err := toViewPtr[*deneb.ExecutionPayload](b.ExecutionPayload, "BeaconBlockBody.ExecutionPayload")
		if err != nil {
			return nil, err
		}

		return &electra.BeaconBlockBody{
			RANDAOReveal:          b.RANDAOReveal,
			ETH1Data:              b.ETH1Data,
			Graffiti:              b.Graffiti,
			ProposerSlashings:     b.ProposerSlashings,
			AttesterSlashings:     as,
			Attestations:          at,
			Deposits:              b.Deposits,
			VoluntaryExits:        b.VoluntaryExits,
			SyncAggregate:         b.SyncAggregate,
			ExecutionPayload:      ep,
			BLSToExecutionChanges: b.BLSToExecutionChanges,
			BlobKZGCommitments:    b.BlobKZGCommitments,
			ExecutionRequests:     b.ExecutionRequests,
		}, nil
	case version.DataVersionGloas:
		as, err := toViewSlice[*electra.AttesterSlashing](b.AttesterSlashings, "BeaconBlockBody.AttesterSlashings")
		if err != nil {
			return nil, err
		}

		at, err := toViewSlice[*electra.Attestation](b.Attestations, "BeaconBlockBody.Attestations")
		if err != nil {
			return nil, err
		}

		bid, err := toViewPtr[*gloas.SignedExecutionPayloadBid](b.SignedExecutionPayloadBid, "BeaconBlockBody.SignedExecutionPayloadBid")
		if err != nil {
			return nil, err
		}

		return &gloas.BeaconBlockBody{
			RANDAOReveal:              b.RANDAOReveal,
			ETH1Data:                  b.ETH1Data,
			Graffiti:                  b.Graffiti,
			ProposerSlashings:         b.ProposerSlashings,
			AttesterSlashings:         as,
			Attestations:              at,
			Deposits:                  b.Deposits,
			VoluntaryExits:            b.VoluntaryExits,
			SyncAggregate:             b.SyncAggregate,
			BLSToExecutionChanges:     b.BLSToExecutionChanges,
			SignedExecutionPayloadBid: bid,
			PayloadAttestations:       b.PayloadAttestations,
			ParentExecutionRequests:   b.ParentExecutionRequests,
		}, nil
	case version.DataVersionHeze:
		as, err := toViewSlice[*electra.AttesterSlashing](b.AttesterSlashings, "BeaconBlockBody.AttesterSlashings")
		if err != nil {
			return nil, err
		}

		at, err := toViewSlice[*electra.Attestation](b.Attestations, "BeaconBlockBody.Attestations")
		if err != nil {
			return nil, err
		}

		bid, err := toViewPtr[*heze.SignedExecutionPayloadBid](b.SignedExecutionPayloadBid, "BeaconBlockBody.SignedExecutionPayloadBid")
		if err != nil {
			return nil, err
		}

		return &heze.BeaconBlockBody{
			RANDAOReveal:              b.RANDAOReveal,
			ETH1Data:                  b.ETH1Data,
			Graffiti:                  b.Graffiti,
			ProposerSlashings:         b.ProposerSlashings,
			AttesterSlashings:         as,
			Attestations:              at,
			Deposits:                  b.Deposits,
			VoluntaryExits:            b.VoluntaryExits,
			SyncAggregate:             b.SyncAggregate,
			BLSToExecutionChanges:     b.BLSToExecutionChanges,
			SignedExecutionPayloadBid: bid,
			PayloadAttestations:       b.PayloadAttestations,
			ParentExecutionRequests:   b.ParentExecutionRequests,
		}, nil
	default:
		return nil, fmt.Errorf("BeaconBlockBody: unsupported version %d", b.Version)
	}
}

// FromView populates b from a fork-specific BeaconBlockBody.
func (b *BeaconBlockBody) FromView(view any) error {
	switch v := view.(type) {
	case *phase0.BeaconBlockBody:
		if b.Version == version.DataVersionUnknown {
			b.Version = version.DataVersionPhase0
		}

		b.RANDAOReveal = v.RANDAOReveal
		b.ETH1Data = v.ETH1Data
		b.Graffiti = v.Graffiti
		b.ProposerSlashings = v.ProposerSlashings
		b.Deposits = v.Deposits
		b.VoluntaryExits = v.VoluntaryExits

		return b.fromPhase0Attestations(v.AttesterSlashings, v.Attestations)
	case *altair.BeaconBlockBody:
		if b.Version == version.DataVersionUnknown {
			b.Version = version.DataVersionAltair
		}

		b.RANDAOReveal = v.RANDAOReveal
		b.ETH1Data = v.ETH1Data
		b.Graffiti = v.Graffiti
		b.ProposerSlashings = v.ProposerSlashings
		b.Deposits = v.Deposits
		b.VoluntaryExits = v.VoluntaryExits
		b.SyncAggregate = v.SyncAggregate

		return b.fromPhase0Attestations(v.AttesterSlashings, v.Attestations)
	case *bellatrix.BeaconBlockBody:
		if b.Version == version.DataVersionUnknown {
			b.Version = version.DataVersionBellatrix
		}

		b.RANDAOReveal = v.RANDAOReveal
		b.ETH1Data = v.ETH1Data
		b.Graffiti = v.Graffiti
		b.ProposerSlashings = v.ProposerSlashings
		b.Deposits = v.Deposits
		b.VoluntaryExits = v.VoluntaryExits
		b.SyncAggregate = v.SyncAggregate

		if err := b.fromPhase0Attestations(v.AttesterSlashings, v.Attestations); err != nil {
			return err
		}

		return b.fromExecutionPayload(v.ExecutionPayload)
	case *capella.BeaconBlockBody:
		if b.Version == version.DataVersionUnknown {
			b.Version = version.DataVersionCapella
		}

		b.RANDAOReveal = v.RANDAOReveal
		b.ETH1Data = v.ETH1Data
		b.Graffiti = v.Graffiti
		b.ProposerSlashings = v.ProposerSlashings
		b.Deposits = v.Deposits
		b.VoluntaryExits = v.VoluntaryExits
		b.SyncAggregate = v.SyncAggregate
		b.BLSToExecutionChanges = v.BLSToExecutionChanges

		if err := b.fromPhase0Attestations(v.AttesterSlashings, v.Attestations); err != nil {
			return err
		}

		return b.fromExecutionPayload(v.ExecutionPayload)
	case *deneb.BeaconBlockBody:
		if b.Version == version.DataVersionUnknown {
			b.Version = version.DataVersionDeneb
		}

		b.RANDAOReveal = v.RANDAOReveal
		b.ETH1Data = v.ETH1Data
		b.Graffiti = v.Graffiti
		b.ProposerSlashings = v.ProposerSlashings
		b.Deposits = v.Deposits
		b.VoluntaryExits = v.VoluntaryExits
		b.SyncAggregate = v.SyncAggregate
		b.BLSToExecutionChanges = v.BLSToExecutionChanges
		b.BlobKZGCommitments = v.BlobKZGCommitments

		if err := b.fromPhase0Attestations(v.AttesterSlashings, v.Attestations); err != nil {
			return err
		}

		return b.fromExecutionPayload(v.ExecutionPayload)
	case *electra.BeaconBlockBody:
		if b.Version == version.DataVersionUnknown {
			b.Version = version.DataVersionElectra
		}

		b.RANDAOReveal = v.RANDAOReveal
		b.ETH1Data = v.ETH1Data
		b.Graffiti = v.Graffiti
		b.ProposerSlashings = v.ProposerSlashings
		b.Deposits = v.Deposits
		b.VoluntaryExits = v.VoluntaryExits
		b.SyncAggregate = v.SyncAggregate
		b.BLSToExecutionChanges = v.BLSToExecutionChanges
		b.BlobKZGCommitments = v.BlobKZGCommitments
		b.ExecutionRequests = v.ExecutionRequests

		if err := b.fromElectraAttestations(v.AttesterSlashings, v.Attestations); err != nil {
			return err
		}

		return b.fromExecutionPayload(v.ExecutionPayload)
	case *gloas.BeaconBlockBody:
		if b.Version == version.DataVersionUnknown {
			b.Version = version.DataVersionGloas
		}

		b.RANDAOReveal = v.RANDAOReveal
		b.ETH1Data = v.ETH1Data
		b.Graffiti = v.Graffiti
		b.ProposerSlashings = v.ProposerSlashings
		b.Deposits = v.Deposits
		b.VoluntaryExits = v.VoluntaryExits
		b.SyncAggregate = v.SyncAggregate
		b.BLSToExecutionChanges = v.BLSToExecutionChanges
		b.PayloadAttestations = v.PayloadAttestations
		b.ParentExecutionRequests = v.ParentExecutionRequests
		b.ExecutionPayload = nil

		if err := b.fromElectraAttestations(v.AttesterSlashings, v.Attestations); err != nil {
			return err
		}

		return b.fromSignedBid(v.SignedExecutionPayloadBid)
	case *heze.BeaconBlockBody:
		if b.Version == version.DataVersionUnknown {
			b.Version = version.DataVersionHeze
		}

		b.RANDAOReveal = v.RANDAOReveal
		b.ETH1Data = v.ETH1Data
		b.Graffiti = v.Graffiti
		b.ProposerSlashings = v.ProposerSlashings
		b.Deposits = v.Deposits
		b.VoluntaryExits = v.VoluntaryExits
		b.SyncAggregate = v.SyncAggregate
		b.BLSToExecutionChanges = v.BLSToExecutionChanges
		b.PayloadAttestations = v.PayloadAttestations
		b.ParentExecutionRequests = v.ParentExecutionRequests
		b.ExecutionPayload = nil

		if err := b.fromElectraAttestations(v.AttesterSlashings, v.Attestations); err != nil {
			return err
		}

		return b.fromSignedBid(v.SignedExecutionPayloadBid)
	default:
		return fmt.Errorf("BeaconBlockBody: unsupported view type %T", view)
	}
}

func (b *BeaconBlockBody) fromPhase0Attestations(slashings []*phase0.AttesterSlashing, attestations []*phase0.Attestation) error {
	asOut := make([]*AttesterSlashing, len(slashings))
	for i, s := range slashings {
		if s == nil {
			continue
		}

		asOut[i] = &AttesterSlashing{Version: b.Version}
		if err := asOut[i].FromView(s); err != nil {
			return fmt.Errorf("attesterSlashings[%d]: %w", i, err)
		}
	}

	b.AttesterSlashings = asOut

	atOut := make([]*Attestation, len(attestations))
	for i, a := range attestations {
		if a == nil {
			continue
		}

		atOut[i] = &Attestation{Version: b.Version}
		if err := atOut[i].FromView(a); err != nil {
			return fmt.Errorf("attestations[%d]: %w", i, err)
		}
	}

	b.Attestations = atOut

	return nil
}

func (b *BeaconBlockBody) fromElectraAttestations(slashings []*electra.AttesterSlashing, attestations []*electra.Attestation) error {
	asOut := make([]*AttesterSlashing, len(slashings))
	for i, s := range slashings {
		if s == nil {
			continue
		}

		asOut[i] = &AttesterSlashing{Version: b.Version}
		if err := asOut[i].FromView(s); err != nil {
			return fmt.Errorf("attesterSlashings[%d]: %w", i, err)
		}
	}

	b.AttesterSlashings = asOut

	atOut := make([]*Attestation, len(attestations))
	for i, a := range attestations {
		if a == nil {
			continue
		}

		atOut[i] = &Attestation{Version: b.Version}
		if err := atOut[i].FromView(a); err != nil {
			return fmt.Errorf("attestations[%d]: %w", i, err)
		}
	}

	b.Attestations = atOut

	return nil
}

func (b *BeaconBlockBody) fromExecutionPayload(view any) error {
	if view == nil {
		b.ExecutionPayload = nil

		return nil
	}

	if b.ExecutionPayload == nil {
		b.ExecutionPayload = &ExecutionPayload{Version: b.Version}
	}

	return b.ExecutionPayload.FromView(view)
}

func (b *BeaconBlockBody) fromSignedBid(view any) error {
	rv := reflect.ValueOf(view)
	if !rv.IsValid() || (rv.Kind() == reflect.Ptr && rv.IsNil()) {
		b.SignedExecutionPayloadBid = nil

		return nil
	}

	if b.SignedExecutionPayloadBid == nil {
		b.SignedExecutionPayloadBid = &SignedExecutionPayloadBid{Version: b.Version}
	}

	return b.SignedExecutionPayloadBid.FromView(view)
}

// HashTreeRootWithDyn computes the SSZ hash tree root using the active Version's view.
func (b *BeaconBlockBody) HashTreeRootWithDyn(ds sszutils.DynamicSpecs, hh sszutils.HashWalker) error {
	view, err := b.viewType()
	if err != nil {
		return err
	}

	h, ok := any(b).(sszutils.DynamicViewHashRoot)
	if !ok {
		return errors.New("BeaconBlockBody: generated SSZ code missing")
	}

	fn := h.HashTreeRootWithDynView(view)
	if fn == nil {
		return fmt.Errorf("BeaconBlockBody: no view hasher for version %d", b.Version)
	}

	return fn(ds, hh)
}

// MarshalSSZ implements the fastssz.Marshaler interface.
func (b *BeaconBlockBody) MarshalSSZ() ([]byte, error) {
	ds := dynssz.GetGlobalDynSsz()

	return b.MarshalSSZDyn(ds, make([]byte, 0, b.SizeSSZDyn(ds)))
}

// MarshalSSZTo implements the fastssz.Marshaler interface.
func (b *BeaconBlockBody) MarshalSSZTo(dst []byte) ([]byte, error) {
	return b.MarshalSSZDyn(dynssz.GetGlobalDynSsz(), dst)
}

// UnmarshalSSZ implements the fastssz.Unmarshaler interface.
func (b *BeaconBlockBody) UnmarshalSSZ(buf []byte) error {
	return b.UnmarshalSSZDyn(dynssz.GetGlobalDynSsz(), buf)
}

// SizeSSZ implements the fastssz.Marshaler interface.
func (b *BeaconBlockBody) SizeSSZ() int {
	return b.SizeSSZDyn(dynssz.GetGlobalDynSsz())
}

// HashTreeRoot implements the fastssz.HashRoot interface.
func (b *BeaconBlockBody) HashTreeRoot() ([32]byte, error) {
	return dynssz.GetGlobalDynSsz().HashTreeRoot(b)
}

// HashTreeRootWith implements the fastssz.HashRoot interface.
func (b *BeaconBlockBody) HashTreeRootWith(hh sszutils.HashWalker) error {
	return b.HashTreeRootWithDyn(dynssz.GetGlobalDynSsz(), hh)
}

// MarshalJSON delegates to the per-fork BeaconBlockBody that matches Version.
func (b *BeaconBlockBody) MarshalJSON() ([]byte, error) {
	return marshalAsView(b)
}

// UnmarshalJSON delegates to the per-fork BeaconBlockBody that matches Version.
// Caller must set Version before calling.
func (b *BeaconBlockBody) UnmarshalJSON(data []byte) error {
	if err := unmarshalAsView(b, data); err != nil {
		return err
	}

	b.populateVersion(b.Version)

	return nil
}

// MarshalYAML delegates to the per-fork BeaconBlockBody that matches Version.
func (b *BeaconBlockBody) MarshalYAML() ([]byte, error) {
	return marshalAsViewYAML(b)
}

// UnmarshalYAML delegates to the per-fork BeaconBlockBody that matches Version.
// Caller must set Version before calling.
func (b *BeaconBlockBody) UnmarshalYAML(data []byte) error {
	if err := unmarshalAsViewYAML(b, data); err != nil {
		return err
	}

	b.populateVersion(b.Version)

	return nil
}
