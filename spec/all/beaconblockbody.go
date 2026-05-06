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

	"github.com/ethpandaops/go-eth2-client/spec"
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
	case version.DataVersionElectra,
		version.DataVersionFulu:
		// Fulu reuses the Electra block-body schema unchanged.
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
// ToView returns a fresh fork-specific BeaconBlockBody populated with b's
// fields. All field copies — including nested versionable children — go
// through copyByName, which walks dst fields by name and recurses into
// nested pointers/slices.
func (b *BeaconBlockBody) ToView() (any, error) {
	return toViewByCopy(b)
}

// FromView populates b from a fork-specific BeaconBlockBody.
func (b *BeaconBlockBody) FromView(view any) error {
	v, err := beaconBlockBodyVersion(view)
	if err != nil {
		return err
	}

	if b.Version == version.DataVersionUnknown {
		b.Version = v
	}

	if err := copyByName(view, b); err != nil {
		return err
	}

	b.populateVersion(b.Version)

	return nil
}

// beaconBlockBodyVersion maps a BeaconBlockBody view type to its DataVersion.
func beaconBlockBodyVersion(view any) (version.DataVersion, error) {
	switch view.(type) {
	case *phase0.BeaconBlockBody:
		return version.DataVersionPhase0, nil
	case *altair.BeaconBlockBody:
		return version.DataVersionAltair, nil
	case *bellatrix.BeaconBlockBody:
		return version.DataVersionBellatrix, nil
	case *capella.BeaconBlockBody:
		return version.DataVersionCapella, nil
	case *deneb.BeaconBlockBody:
		return version.DataVersionDeneb, nil
	case *electra.BeaconBlockBody:
		return version.DataVersionElectra, nil
	case *gloas.BeaconBlockBody:
		return version.DataVersionGloas, nil
	case *heze.BeaconBlockBody:
		return version.DataVersionHeze, nil
	default:
		return version.DataVersionUnknown, fmt.Errorf("BeaconBlockBody: unsupported view type %T", view)
	}
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

// ToVersioned converts b into a *spec.VersionedBeaconBlockBody.
func (b *BeaconBlockBody) ToVersioned() (*spec.VersionedBeaconBlockBody, error) {
	out := &spec.VersionedBeaconBlockBody{}
	if err := toVersioned(b.Version, b, out); err != nil {
		return nil, err
	}

	return out, nil
}

// FromVersioned populates b from src.
func (b *BeaconBlockBody) FromVersioned(src *spec.VersionedBeaconBlockBody) error {
	return fromVersioned(b, src)
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
