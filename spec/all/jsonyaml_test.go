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

package all_test

import (
	"encoding/json"
	"testing"

	"github.com/ethpandaops/go-eth2-client/spec/all"
	"github.com/ethpandaops/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/go-eth2-client/spec/version"
)

// TestPhase0BeaconBlockJSONRoundtrip checks that an all.BeaconBlock pinned to
// phase0 marshals to the same JSON as the equivalent phase0.BeaconBlock and
// round-trips back to an equivalent all.BeaconBlock.
func TestPhase0BeaconBlockJSONRoundtrip(t *testing.T) {
	// Build a phase0 BeaconBlock with realistic-ish data.
	phase0Block := &phase0.BeaconBlock{
		Slot:          12345,
		ProposerIndex: 42,
		ParentRoot:    phase0.Root{0x01, 0x02},
		StateRoot:     phase0.Root{0x03, 0x04},
		Body: &phase0.BeaconBlockBody{
			RANDAOReveal: phase0.BLSSignature{0xaa, 0xbb},
			ETH1Data: &phase0.ETH1Data{
				DepositRoot:  phase0.Root{0x10},
				DepositCount: 7,
				BlockHash:    []byte{0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f},
			},
			Graffiti:          [32]byte{'h', 'i'},
			ProposerSlashings: []*phase0.ProposerSlashing{},
			AttesterSlashings: []*phase0.AttesterSlashing{},
			Attestations:      []*phase0.Attestation{},
			Deposits:          []*phase0.Deposit{},
			VoluntaryExits:    []*phase0.SignedVoluntaryExit{},
		},
	}

	expectedJSON, err := json.Marshal(phase0Block)
	if err != nil {
		t.Fatalf("phase0 marshal failed: %v", err)
	}

	// Now build the equivalent all.BeaconBlock.
	allBlock := &all.BeaconBlock{
		Version:       version.DataVersionPhase0,
		Slot:          phase0Block.Slot,
		ProposerIndex: phase0Block.ProposerIndex,
		ParentRoot:    phase0Block.ParentRoot,
		StateRoot:     phase0Block.StateRoot,
		Body: &all.BeaconBlockBody{
			Version:           version.DataVersionPhase0,
			RANDAOReveal:      phase0Block.Body.RANDAOReveal,
			ETH1Data:          phase0Block.Body.ETH1Data,
			Graffiti:          phase0Block.Body.Graffiti,
			ProposerSlashings: phase0Block.Body.ProposerSlashings,
			AttesterSlashings: []*all.AttesterSlashing{},
			Attestations:      []*all.Attestation{},
			Deposits:          phase0Block.Body.Deposits,
			VoluntaryExits:    phase0Block.Body.VoluntaryExits,
		},
	}

	// Marshal and verify the bytes match the per-fork output.
	gotJSON, err := json.Marshal(allBlock)
	if err != nil {
		t.Fatalf("all marshal failed: %v", err)
	}

	if string(gotJSON) != string(expectedJSON) {
		t.Fatalf("marshaled JSON differs from per-fork output\nwant: %s\n got: %s", expectedJSON, gotJSON)
	}

	// Unmarshal back into a fresh all.BeaconBlock with Version pre-set.
	rt := &all.BeaconBlock{Version: version.DataVersionPhase0}
	if err := json.Unmarshal(gotJSON, rt); err != nil {
		t.Fatalf("all unmarshal failed: %v", err)
	}

	// Version must propagate to the body.
	if rt.Version != version.DataVersionPhase0 {
		t.Fatalf("rt.Version = %d", rt.Version)
	}

	if rt.Body == nil {
		t.Fatalf("rt.Body is nil")
	}

	if rt.Body.Version != version.DataVersionPhase0 {
		t.Fatalf("rt.Body.Version = %d, want %d (populateVersion did not propagate)",
			rt.Body.Version, version.DataVersionPhase0)
	}

	// Top-level fields preserved.
	if rt.Slot != allBlock.Slot {
		t.Fatalf("Slot: got %d want %d", rt.Slot, allBlock.Slot)
	}

	if rt.ProposerIndex != allBlock.ProposerIndex {
		t.Fatalf("ProposerIndex: got %d want %d", rt.ProposerIndex, allBlock.ProposerIndex)
	}

	// Re-marshal and compare to verify lossless round-trip.
	rtJSON, err := json.Marshal(rt)
	if err != nil {
		t.Fatalf("re-marshal failed: %v", err)
	}

	if string(rtJSON) != string(expectedJSON) {
		t.Fatalf("round-tripped JSON differs\nwant: %s\n got: %s", expectedJSON, rtJSON)
	}
}

// TestUnmarshalNoVersionRejected verifies that calling UnmarshalJSON without
// setting Version returns an error rather than producing garbage.
func TestUnmarshalNoVersionRejected(t *testing.T) {
	b := &all.BeaconBlock{}
	if err := json.Unmarshal([]byte(`{}`), b); err == nil {
		t.Fatal("expected error for unset Version, got nil")
	}
}
