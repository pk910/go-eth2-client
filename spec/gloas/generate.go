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

package gloas

//nolint:revive
// Need to `go install github.com/pk910/dynamic-ssz/dynssz-gen@latest` for this to work.
//go:generate rm -f beaconblockbody_ssz.go beaconblock_ssz.go beaconstate_ssz.go builder_ssz.go builderpendingpayment_ssz.go builderpendingwithdrawal_ssz.go executionpayloadbid_ssz.go executionpayloadenvelope_ssz.go indexedpayloadattestation_ssz.go payloadattestation_ssz.go payloadattestationdata_ssz.go payloadattestationmessage_ssz.go signedbeaconblock_ssz.go signedexecutionpayloadbid_ssz.go signedexecutionpayloadenvelope_ssz.go
//go:generate dynssz-gen -package . -legacy -without-dynamic-expressions -types BeaconBlockBody:beaconblockbody_ssz.go,BeaconBlock:beaconblock_ssz.go,BeaconState:beaconstate_ssz.go,Builder:builder_ssz.go,BuilderPendingPayment:builderpendingpayment_ssz.go,BuilderPendingWithdrawal:builderpendingwithdrawal_ssz.go,ExecutionPayloadBid:executionpayloadbid_ssz.go,ExecutionPayloadEnvelope:executionpayloadenvelope_ssz.go,IndexedPayloadAttestation:indexedpayloadattestation_ssz.go,PayloadAttestation:payloadattestation_ssz.go,PayloadAttestationData:payloadattestationdata_ssz.go,PayloadAttestationMessage:payloadattestationmessage_ssz.go,SignedBeaconBlock:signedbeaconblock_ssz.go,SignedExecutionPayloadBid:signedexecutionpayloadbid_ssz.go,SignedExecutionPayloadEnvelope:signedexecutionpayloadenvelope_ssz.go
