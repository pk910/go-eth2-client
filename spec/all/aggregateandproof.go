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

	"github.com/ethpandaops/go-eth2-client/spec/electra"
	"github.com/ethpandaops/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/go-eth2-client/spec/version"
	dynssz "github.com/pk910/dynamic-ssz"
	"github.com/pk910/dynamic-ssz/sszutils"
)

// AggregateAndProof is a fork-agnostic aggregate and proof. The wrapped
// Attestation gains CommitteeBits from Electra onwards.
type AggregateAndProof struct {
	Version         version.DataVersion
	AggregatorIndex phase0.ValidatorIndex
	Aggregate       *Attestation
	SelectionProof  phase0.BLSSignature
}

// viewType returns the fork-specific schema type pointer used as the view
// descriptor for the active Version.
func (a *AggregateAndProof) viewType() (any, error) {
	switch a.Version {
	case version.DataVersionPhase0,
		version.DataVersionAltair,
		version.DataVersionBellatrix,
		version.DataVersionCapella,
		version.DataVersionDeneb:
		return (*phase0.AggregateAndProof)(nil), nil
	case version.DataVersionElectra,
		version.DataVersionFulu,
		version.DataVersionGloas,
		version.DataVersionHeze:
		return (*electra.AggregateAndProof)(nil), nil
	default:
		return nil, fmt.Errorf("AggregateAndProof: unsupported version %d", a.Version)
	}
}

// MarshalSSZDyn marshals the proof using the view that matches Version.
func (a *AggregateAndProof) MarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) ([]byte, error) {
	view, err := a.viewType()
	if err != nil {
		return nil, err
	}

	m, ok := any(a).(sszutils.DynamicViewMarshaler)
	if !ok {
		return nil, errors.New("AggregateAndProof: generated SSZ code missing")
	}

	fn := m.MarshalSSZDynView(view)
	if fn == nil {
		return nil, fmt.Errorf("AggregateAndProof: no view marshaler for version %d", a.Version)
	}

	return fn(ds, buf)
}

// SizeSSZDyn returns the SSZ size of the proof for the active Version.
func (a *AggregateAndProof) SizeSSZDyn(ds sszutils.DynamicSpecs) int {
	view, err := a.viewType()
	if err != nil {
		return 0
	}

	s, ok := any(a).(sszutils.DynamicViewSizer)
	if !ok {
		return 0
	}

	fn := s.SizeSSZDynView(view)
	if fn == nil {
		return 0
	}

	return fn(ds)
}

// UnmarshalSSZDyn decodes the proof into the view that matches Version.
func (a *AggregateAndProof) UnmarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) error {
	view, err := a.viewType()
	if err != nil {
		return err
	}

	u, ok := any(a).(sszutils.DynamicViewUnmarshaler)
	if !ok {
		return errors.New("AggregateAndProof: generated SSZ code missing")
	}

	fn := u.UnmarshalSSZDynView(view)
	if fn == nil {
		return fmt.Errorf("AggregateAndProof: no view unmarshaler for version %d", a.Version)
	}

	if err := fn(ds, buf); err != nil {
		return err
	}

	a.populateVersion(a.Version)

	return nil
}

// populateVersion sets Version and propagates it to the inner attestation.
func (a *AggregateAndProof) populateVersion(v version.DataVersion) {
	a.Version = v

	if a.Aggregate != nil {
		a.Aggregate.populateVersion(v)
	}
}

// ToView returns a fresh fork-specific AggregateAndProof populated with a's
// fields, recursing into Aggregate via its ToView.
func (a *AggregateAndProof) ToView() (any, error) {
	var agg any

	var err error

	if a.Aggregate != nil {
		agg, err = a.Aggregate.ToView()
		if err != nil {
			return nil, err
		}
	}

	switch a.Version {
	case version.DataVersionPhase0,
		version.DataVersionAltair,
		version.DataVersionBellatrix,
		version.DataVersionCapella,
		version.DataVersionDeneb:
		pa, err := assertView[*phase0.Attestation](agg, "AggregateAndProof.Aggregate")
		if err != nil {
			return nil, err
		}

		return &phase0.AggregateAndProof{
			AggregatorIndex: a.AggregatorIndex,
			Aggregate:       pa,
			SelectionProof:  a.SelectionProof,
		}, nil
	case version.DataVersionElectra,
		version.DataVersionFulu,
		version.DataVersionGloas,
		version.DataVersionHeze:
		ea, err := assertView[*electra.Attestation](agg, "AggregateAndProof.Aggregate")
		if err != nil {
			return nil, err
		}

		return &electra.AggregateAndProof{
			AggregatorIndex: a.AggregatorIndex,
			Aggregate:       ea,
			SelectionProof:  a.SelectionProof,
		}, nil
	default:
		return nil, fmt.Errorf("AggregateAndProof: unsupported version %d", a.Version)
	}
}

// FromView populates a from a fork-specific AggregateAndProof.
func (a *AggregateAndProof) FromView(view any) error {
	var aggView any

	switch v := view.(type) {
	case *phase0.AggregateAndProof:
		if a.Version == version.DataVersionUnknown {
			a.Version = version.DataVersionPhase0
		}

		a.AggregatorIndex = v.AggregatorIndex
		a.SelectionProof = v.SelectionProof
		aggView = v.Aggregate
	case *electra.AggregateAndProof:
		if a.Version == version.DataVersionUnknown {
			a.Version = version.DataVersionElectra
		}

		a.AggregatorIndex = v.AggregatorIndex
		a.SelectionProof = v.SelectionProof

		if v.Aggregate != nil {
			aggView = v.Aggregate
		}
	default:
		return fmt.Errorf("AggregateAndProof: unsupported view type %T", view)
	}

	if aggView == nil {
		a.Aggregate = nil

		return nil
	}

	if a.Aggregate == nil {
		a.Aggregate = &Attestation{Version: a.Version}
	}

	return a.Aggregate.FromView(aggView)
}

// HashTreeRootWithDyn computes the SSZ hash tree root using the active Version's view.
func (a *AggregateAndProof) HashTreeRootWithDyn(ds sszutils.DynamicSpecs, hh sszutils.HashWalker) error {
	view, err := a.viewType()
	if err != nil {
		return err
	}

	h, ok := any(a).(sszutils.DynamicViewHashRoot)
	if !ok {
		return errors.New("AggregateAndProof: generated SSZ code missing")
	}

	fn := h.HashTreeRootWithDynView(view)
	if fn == nil {
		return fmt.Errorf("AggregateAndProof: no view hasher for version %d", a.Version)
	}

	return fn(ds, hh)
}

// MarshalSSZ implements the fastssz.Marshaler interface.
func (a *AggregateAndProof) MarshalSSZ() ([]byte, error) {
	ds := dynssz.GetGlobalDynSsz()

	return a.MarshalSSZDyn(ds, make([]byte, 0, a.SizeSSZDyn(ds)))
}

// MarshalSSZTo implements the fastssz.Marshaler interface.
func (a *AggregateAndProof) MarshalSSZTo(dst []byte) ([]byte, error) {
	return a.MarshalSSZDyn(dynssz.GetGlobalDynSsz(), dst)
}

// UnmarshalSSZ implements the fastssz.Unmarshaler interface.
func (a *AggregateAndProof) UnmarshalSSZ(buf []byte) error {
	return a.UnmarshalSSZDyn(dynssz.GetGlobalDynSsz(), buf)
}

// SizeSSZ implements the fastssz.Marshaler interface.
func (a *AggregateAndProof) SizeSSZ() int {
	return a.SizeSSZDyn(dynssz.GetGlobalDynSsz())
}

// HashTreeRoot implements the fastssz.HashRoot interface.
func (a *AggregateAndProof) HashTreeRoot() (phase0.Root, error) {
	return dynssz.GetGlobalDynSsz().HashTreeRoot(a)
}

// HashTreeRootWith implements the fastssz.HashRoot interface.
func (a *AggregateAndProof) HashTreeRootWith(hh sszutils.HashWalker) error {
	return a.HashTreeRootWithDyn(dynssz.GetGlobalDynSsz(), hh)
}

// MarshalJSON delegates to the per-fork AggregateAndProof that matches Version.
func (a *AggregateAndProof) MarshalJSON() ([]byte, error) {
	return marshalAsView(a)
}

// UnmarshalJSON delegates to the per-fork AggregateAndProof that matches Version.
// Caller must set Version before calling.
func (a *AggregateAndProof) UnmarshalJSON(data []byte) error {
	if err := unmarshalAsView(a, data); err != nil {
		return err
	}

	a.populateVersion(a.Version)

	return nil
}

// MarshalYAML delegates to the per-fork AggregateAndProof that matches Version.
func (a *AggregateAndProof) MarshalYAML() ([]byte, error) {
	return marshalAsViewYAML(a)
}

// UnmarshalYAML delegates to the per-fork AggregateAndProof that matches Version.
// Caller must set Version before calling.
func (a *AggregateAndProof) UnmarshalYAML(data []byte) error {
	if err := unmarshalAsViewYAML(a, data); err != nil {
		return err
	}

	a.populateVersion(a.Version)

	return nil
}
