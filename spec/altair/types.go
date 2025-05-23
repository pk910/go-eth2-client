// Copyright © 2021 Attestant Limited.
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

package altair

// ParticipationFlag is an individual participation flag for a validator.
type ParticipationFlag int

const (
	// TimelySourceFlagIndex is set when an attestation has a timely source value.
	TimelySourceFlagIndex ParticipationFlag = iota
	// TimelyTargetFlagIndex is set when an attestation has a timely target value.
	TimelyTargetFlagIndex
	// TimelyHeadFlagIndex is set when an attestation has a timely head value.
	TimelyHeadFlagIndex
)
