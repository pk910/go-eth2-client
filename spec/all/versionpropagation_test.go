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
	"testing"

	"github.com/ethpandaops/go-eth2-client/spec/version"
)

// TestPopulateVersion verifies that populateVersion seeds Version on a top-level
// fork-agnostic type and recurses into nested versionable children.
func TestPopulateVersion(t *testing.T) {
	t.Run("nested ptr", func(t *testing.T) {
		b := &BeaconBlock{
			Body: &BeaconBlockBody{},
		}
		b.populateVersion(version.DataVersionDeneb)

		if b.Version != version.DataVersionDeneb {
			t.Fatalf("BeaconBlock.Version = %d, want %d", b.Version, version.DataVersionDeneb)
		}

		if b.Body.Version != version.DataVersionDeneb {
			t.Fatalf("BeaconBlock.Body.Version = %d, want %d", b.Body.Version, version.DataVersionDeneb)
		}
	})

	t.Run("slice of ptr", func(t *testing.T) {
		body := &BeaconBlockBody{
			AttesterSlashings: []*AttesterSlashing{
				{Attestation1: &IndexedAttestation{}, Attestation2: &IndexedAttestation{}},
				{Attestation1: &IndexedAttestation{}, Attestation2: &IndexedAttestation{}},
			},
		}
		body.populateVersion(version.DataVersionElectra)

		for i, as := range body.AttesterSlashings {
			if as.Version != version.DataVersionElectra {
				t.Fatalf("AttesterSlashings[%d].Version = %d, want %d", i, as.Version, version.DataVersionElectra)
			}

			if as.Attestation1.Version != version.DataVersionElectra {
				t.Fatalf("AttesterSlashings[%d].Attestation1.Version = %d, want %d",
					i, as.Attestation1.Version, version.DataVersionElectra)
			}

			if as.Attestation2.Version != version.DataVersionElectra {
				t.Fatalf("AttesterSlashings[%d].Attestation2.Version = %d, want %d",
					i, as.Attestation2.Version, version.DataVersionElectra)
			}
		}
	})

	t.Run("deeply nested", func(t *testing.T) {
		s := &SignedBeaconBlock{
			Message: &BeaconBlock{
				Body: &BeaconBlockBody{
					ExecutionPayload: &ExecutionPayload{},
					Attestations: []*Attestation{
						{}, {},
					},
				},
			},
		}
		s.populateVersion(version.DataVersionGloas)

		if s.Version != version.DataVersionGloas {
			t.Fatalf("SignedBeaconBlock.Version = %d", s.Version)
		}

		if s.Message.Version != version.DataVersionGloas {
			t.Fatalf("Message.Version = %d", s.Message.Version)
		}

		if s.Message.Body.Version != version.DataVersionGloas {
			t.Fatalf("Body.Version = %d", s.Message.Body.Version)
		}

		if s.Message.Body.ExecutionPayload.Version != version.DataVersionGloas {
			t.Fatalf("ExecutionPayload.Version = %d", s.Message.Body.ExecutionPayload.Version)
		}

		for i, a := range s.Message.Body.Attestations {
			if a.Version != version.DataVersionGloas {
				t.Fatalf("Attestations[%d].Version = %d", i, a.Version)
			}
		}
	})

	t.Run("nil children skipped", func(t *testing.T) {
		// Body is nil — populateVersion must not panic.
		b := &BeaconBlock{}
		b.populateVersion(version.DataVersionPhase0)

		if b.Version != version.DataVersionPhase0 {
			t.Fatalf("Version = %d", b.Version)
		}
	})

	t.Run("primitive arrays not walked", func(t *testing.T) {
		// Sanity: a body with a populated Graffiti ([32]byte) and LogsBloom-sized fields
		// must not blow up the walker.
		body := &BeaconBlockBody{}
		body.Graffiti = [32]byte{1, 2, 3}
		body.populateVersion(version.DataVersionDeneb)

		if body.Version != version.DataVersionDeneb {
			t.Fatalf("Version = %d", body.Version)
		}
	})
}
