// Copyright © 2026 Attestant Limited.
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

package all_test

import (
	"reflect"
	"testing"

	"github.com/ethpandaops/go-eth2-client/spec"
	"github.com/ethpandaops/go-eth2-client/spec/all"
	"github.com/ethpandaops/go-eth2-client/spec/electra"
	"github.com/ethpandaops/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/go-eth2-client/spec/version"
)

// TestToVersionedFromVersionedAttestation verifies the ToVersioned/FromVersioned
// round-trip for a leaf type that maps across two view types (phase0 + electra).
func TestToVersionedFromVersionedAttestation(t *testing.T) {
	t.Run("phase0", func(t *testing.T) {
		src := &all.Attestation{
			Version:         version.DataVersionDeneb,
			AggregationBits: []byte{0x01, 0x02},
			Data:            &phase0.AttestationData{Slot: 7},
			Signature:       phase0.BLSSignature{0xab, 0xcd},
		}

		v, err := src.ToVersioned()
		if err != nil {
			t.Fatalf("ToVersioned: %v", err)
		}

		if v.Version != version.DataVersionDeneb {
			t.Fatalf("Version = %d, want %d", v.Version, version.DataVersionDeneb)
		}

		if v.Deneb == nil {
			t.Fatal("Deneb field is nil")
		}

		if v.Phase0 != nil || v.Electra != nil {
			t.Fatalf("non-Deneb fields should be nil; got phase0=%v electra=%v", v.Phase0, v.Electra)
		}

		if v.Deneb.Data.Slot != 7 {
			t.Fatalf("Slot lost: %d", v.Deneb.Data.Slot)
		}

		// Round-trip back.
		rt := &all.Attestation{}
		if err := rt.FromVersioned(v); err != nil {
			t.Fatalf("FromVersioned: %v", err)
		}

		if rt.Version != version.DataVersionDeneb {
			t.Fatalf("rt.Version = %d", rt.Version)
		}

		if rt.Data.Slot != 7 {
			t.Fatalf("rt slot: %d", rt.Data.Slot)
		}
	})

	t.Run("electra carries CommitteeBits", func(t *testing.T) {
		src := &all.Attestation{
			Version:         version.DataVersionElectra,
			AggregationBits: []byte{0x10},
			Data:            &phase0.AttestationData{Slot: 9},
			CommitteeBits:   make([]byte, 8),
		}
		src.CommitteeBits[0] = 0x42

		v, err := src.ToVersioned()
		if err != nil {
			t.Fatalf("ToVersioned: %v", err)
		}

		if v.Electra == nil || v.Electra.CommitteeBits[0] != 0x42 {
			t.Fatalf("CommitteeBits lost: %+v", v.Electra)
		}

		rt := &all.Attestation{}
		if err := rt.FromVersioned(v); err != nil {
			t.Fatalf("FromVersioned: %v", err)
		}

		if rt.CommitteeBits[0] != 0x42 {
			t.Fatalf("rt CommitteeBits: %x", rt.CommitteeBits[0])
		}
	})
}

// TestToVersionedFromVersionedSignedBeaconBlock verifies a deeper tree ToVersioned
// round-trip — including the Fulu→electra schema reuse.
func TestToVersionedFromVersionedSignedBeaconBlock(t *testing.T) {
	src := &all.SignedBeaconBlock{
		Version:   version.DataVersionFulu,
		Signature: phase0.BLSSignature{0xff},
		Message: &all.BeaconBlock{
			Version:       version.DataVersionFulu,
			Slot:          42,
			ProposerIndex: 7,
			Body: &all.BeaconBlockBody{
				Version:           version.DataVersionFulu,
				ETH1Data:          &phase0.ETH1Data{DepositCount: 5},
				ProposerSlashings: []*phase0.ProposerSlashing{},
				AttesterSlashings: []*all.AttesterSlashing{},
				Attestations:      []*all.Attestation{},
				Deposits:          []*phase0.Deposit{},
				VoluntaryExits:    []*phase0.SignedVoluntaryExit{},
			},
		},
	}

	v, err := src.ToVersioned()
	if err != nil {
		t.Fatalf("ToVersioned: %v", err)
	}

	if v.Version != version.DataVersionFulu {
		t.Fatalf("Version = %d", v.Version)
	}

	if v.Fulu == nil {
		t.Fatal("Fulu field is nil")
	}

	if v.Fulu.Message.Slot != 42 {
		t.Fatalf("Slot lost: %d", v.Fulu.Message.Slot)
	}

	rt := &all.SignedBeaconBlock{}
	if err := rt.FromVersioned(v); err != nil {
		t.Fatalf("FromVersioned: %v", err)
	}

	if rt.Version != version.DataVersionFulu {
		t.Fatalf("rt Version: %d", rt.Version)
	}

	if rt.Message.Body.Version != version.DataVersionFulu {
		t.Fatalf("body Version not propagated: %d", rt.Message.Body.Version)
	}

	if rt.Message.Slot != 42 {
		t.Fatalf("rt slot: %d", rt.Message.Slot)
	}
}

// TestFromVersionedNilField verifies an error is returned when the populated
// view field for src.Version is nil.
func TestFromVersionedNilField(t *testing.T) {
	v := &spec.VersionedAttestation{
		Version: version.DataVersionPhase0,
		// Phase0 left nil
	}

	rt := &all.Attestation{}
	if err := rt.FromVersioned(v); err == nil {
		t.Fatal("expected error for nil Phase0 field")
	}
}

// TestVersionedRoundTripWithFromVersionedFresh ensures FromVersioned does not
// require a pre-pinned Version on the destination — it should be inferred
// from the view type via FromView's existing logic.
func TestVersionedRoundTripWithFromVersionedFresh(t *testing.T) {
	src := &all.IndexedAttestation{
		Version:          version.DataVersionElectra,
		AttestingIndices: []uint64{1, 2, 3},
		Data:             &phase0.AttestationData{Slot: 11},
		Signature:        phase0.BLSSignature{0xaa},
	}

	v, err := src.ToVersioned()
	if err != nil {
		t.Fatalf("ToVersioned: %v", err)
	}

	rt := &all.IndexedAttestation{} // Version=Unknown
	if err := rt.FromVersioned(v); err != nil {
		t.Fatalf("FromVersioned: %v", err)
	}

	if rt.Version != version.DataVersionElectra {
		t.Fatalf("Version not inferred: %d", rt.Version)
	}

	if !reflect.DeepEqual(rt.AttestingIndices, src.AttestingIndices) {
		t.Fatalf("AttestingIndices lost: %v", rt.AttestingIndices)
	}
}

// TestExecutionPayloadVersionedFulu verifies that ExecutionPayload with
// Version=Fulu maps to VersionedExecutionPayload.Fulu (which is *deneb.ExecutionPayload).
func TestExecutionPayloadVersionedFulu(t *testing.T) {
	src := &all.ExecutionPayload{
		Version:     version.DataVersionFulu,
		BlockNumber: 100,
		Timestamp:   200,
	}

	v, err := src.ToVersioned()
	if err != nil {
		t.Fatalf("ToVersioned: %v", err)
	}

	if v.Fulu == nil {
		t.Fatal("Fulu field is nil")
	}

	// VersionedExecutionPayload.Fulu is *deneb.ExecutionPayload.
	if v.Fulu.BlockNumber != 100 {
		t.Fatalf("BlockNumber lost: %d", v.Fulu.BlockNumber)
	}

	rt := &all.ExecutionPayload{}
	if err := rt.FromVersioned(v); err != nil {
		t.Fatalf("FromVersioned: %v", err)
	}

	if rt.Version != version.DataVersionFulu {
		t.Fatalf("rt Version: %d", rt.Version)
	}

	if rt.BlockNumber != 100 {
		t.Fatalf("rt BlockNumber: %d", rt.BlockNumber)
	}
}

// _ silences the `electra` import-warning if Go decides AttestingIndices
// isn't enough to keep the import used.
var _ = electra.IndexedAttestation{}
