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

// BeaconBlock is a fork-agnostic beacon block. The Body's contents depend on
// Version.
type BeaconBlock struct {
	Version       version.DataVersion
	Slot          phase0.Slot
	ProposerIndex phase0.ValidatorIndex
	ParentRoot    phase0.Root
	StateRoot     phase0.Root
	Body          *BeaconBlockBody
}

// viewType returns the fork-specific schema type pointer used as the view
// descriptor for the active Version.
func (b *BeaconBlock) viewType() (any, error) {
	switch b.Version {
	case version.DataVersionPhase0:
		return (*phase0.BeaconBlock)(nil), nil
	case version.DataVersionAltair:
		return (*altair.BeaconBlock)(nil), nil
	case version.DataVersionBellatrix:
		return (*bellatrix.BeaconBlock)(nil), nil
	case version.DataVersionCapella:
		return (*capella.BeaconBlock)(nil), nil
	case version.DataVersionDeneb:
		return (*deneb.BeaconBlock)(nil), nil
	case version.DataVersionElectra,
		version.DataVersionFulu:
		// Fulu reuses the Electra block schema unchanged.
		return (*electra.BeaconBlock)(nil), nil
	case version.DataVersionGloas:
		return (*gloas.BeaconBlock)(nil), nil
	case version.DataVersionHeze:
		return (*heze.BeaconBlock)(nil), nil
	default:
		return nil, fmt.Errorf("BeaconBlock: unsupported version %d", b.Version)
	}
}

// MarshalSSZDyn marshals the block using the view that matches Version.
func (b *BeaconBlock) MarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) ([]byte, error) {
	view, err := b.viewType()
	if err != nil {
		return nil, err
	}

	m, ok := any(b).(sszutils.DynamicViewMarshaler)
	if !ok {
		return nil, errors.New("BeaconBlock: generated SSZ code missing")
	}

	fn := m.MarshalSSZDynView(view)
	if fn == nil {
		return nil, fmt.Errorf("BeaconBlock: no view marshaler for version %d", b.Version)
	}

	return fn(ds, buf)
}

// SizeSSZDyn returns the SSZ size of the block for the active Version.
func (b *BeaconBlock) SizeSSZDyn(ds sszutils.DynamicSpecs) int {
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

// UnmarshalSSZDyn decodes the block into the view that matches Version.
func (b *BeaconBlock) UnmarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) error {
	view, err := b.viewType()
	if err != nil {
		return err
	}

	u, ok := any(b).(sszutils.DynamicViewUnmarshaler)
	if !ok {
		return errors.New("BeaconBlock: generated SSZ code missing")
	}

	fn := u.UnmarshalSSZDynView(view)
	if fn == nil {
		return fmt.Errorf("BeaconBlock: no view unmarshaler for version %d", b.Version)
	}

	if err := fn(ds, buf); err != nil {
		return err
	}

	b.populateVersion(b.Version)

	return nil
}

// populateVersion sets Version and propagates it to any nested versionable
// children allocated by the SSZ unmarshal.
func (b *BeaconBlock) populateVersion(v version.DataVersion) {
	b.Version = v

	if b.Body != nil {
		b.Body.populateVersion(v)
	}
}

// ToView returns a fresh fork-specific BeaconBlock populated with b's fields,
// recursing into Body via copyByName.
func (b *BeaconBlock) ToView() (any, error) {
	return toViewByCopy(b)
}

// FromView populates b from a fork-specific BeaconBlock.
func (b *BeaconBlock) FromView(view any) error {
	v, err := beaconBlockVersion(view)
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

// beaconBlockVersion maps a BeaconBlock view type to its DataVersion.
func beaconBlockVersion(view any) (version.DataVersion, error) {
	switch view.(type) {
	case *phase0.BeaconBlock:
		return version.DataVersionPhase0, nil
	case *altair.BeaconBlock:
		return version.DataVersionAltair, nil
	case *bellatrix.BeaconBlock:
		return version.DataVersionBellatrix, nil
	case *capella.BeaconBlock:
		return version.DataVersionCapella, nil
	case *deneb.BeaconBlock:
		return version.DataVersionDeneb, nil
	case *electra.BeaconBlock:
		return version.DataVersionElectra, nil
	case *gloas.BeaconBlock:
		return version.DataVersionGloas, nil
	case *heze.BeaconBlock:
		return version.DataVersionHeze, nil
	default:
		return version.DataVersionUnknown, fmt.Errorf("BeaconBlock: unsupported view type %T", view)
	}
}

// HashTreeRootWithDyn computes the SSZ hash tree root using the active Version's view.
func (b *BeaconBlock) HashTreeRootWithDyn(ds sszutils.DynamicSpecs, hh sszutils.HashWalker) error {
	view, err := b.viewType()
	if err != nil {
		return err
	}

	h, ok := any(b).(sszutils.DynamicViewHashRoot)
	if !ok {
		return errors.New("BeaconBlock: generated SSZ code missing")
	}

	fn := h.HashTreeRootWithDynView(view)
	if fn == nil {
		return fmt.Errorf("BeaconBlock: no view hasher for version %d", b.Version)
	}

	return fn(ds, hh)
}

// MarshalSSZ implements the fastssz.Marshaler interface.
func (b *BeaconBlock) MarshalSSZ() ([]byte, error) {
	ds := dynssz.GetGlobalDynSsz()

	return b.MarshalSSZDyn(ds, make([]byte, 0, b.SizeSSZDyn(ds)))
}

// MarshalSSZTo implements the fastssz.Marshaler interface.
func (b *BeaconBlock) MarshalSSZTo(dst []byte) ([]byte, error) {
	return b.MarshalSSZDyn(dynssz.GetGlobalDynSsz(), dst)
}

// UnmarshalSSZ implements the fastssz.Unmarshaler interface.
func (b *BeaconBlock) UnmarshalSSZ(buf []byte) error {
	return b.UnmarshalSSZDyn(dynssz.GetGlobalDynSsz(), buf)
}

// SizeSSZ implements the fastssz.Marshaler interface.
func (b *BeaconBlock) SizeSSZ() int {
	return b.SizeSSZDyn(dynssz.GetGlobalDynSsz())
}

// HashTreeRoot implements the fastssz.HashRoot interface.
func (b *BeaconBlock) HashTreeRoot() ([32]byte, error) {
	return dynssz.GetGlobalDynSsz().HashTreeRoot(b)
}

// HashTreeRootWith implements the fastssz.HashRoot interface.
func (b *BeaconBlock) HashTreeRootWith(hh sszutils.HashWalker) error {
	return b.HashTreeRootWithDyn(dynssz.GetGlobalDynSsz(), hh)
}

// ToVersioned converts b into a *spec.VersionedBeaconBlock.
func (b *BeaconBlock) ToVersioned() (*spec.VersionedBeaconBlock, error) {
	out := &spec.VersionedBeaconBlock{}
	if err := toVersioned(b.Version, b, out); err != nil {
		return nil, err
	}

	return out, nil
}

// FromVersioned populates b from src.
func (b *BeaconBlock) FromVersioned(src *spec.VersionedBeaconBlock) error {
	return fromVersioned(b, src)
}

// MarshalJSON delegates to the per-fork BeaconBlock that matches Version.
func (b *BeaconBlock) MarshalJSON() ([]byte, error) {
	return marshalAsView(b)
}

// UnmarshalJSON delegates to the per-fork BeaconBlock that matches Version.
// Caller must set Version before calling.
func (b *BeaconBlock) UnmarshalJSON(data []byte) error {
	if err := unmarshalAsView(b, data); err != nil {
		return err
	}

	b.populateVersion(b.Version)

	return nil
}

// MarshalYAML delegates to the per-fork BeaconBlock that matches Version.
func (b *BeaconBlock) MarshalYAML() ([]byte, error) {
	return marshalAsViewYAML(b)
}

// UnmarshalYAML delegates to the per-fork BeaconBlock that matches Version.
// Caller must set Version before calling.
func (b *BeaconBlock) UnmarshalYAML(data []byte) error {
	if err := unmarshalAsViewYAML(b, data); err != nil {
		return err
	}

	b.populateVersion(b.Version)

	return nil
}
