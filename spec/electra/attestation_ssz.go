// Code generated by fastssz. DO NOT EDIT.
// Hash: 5a2d052da32c0d16b124cd692d1cc933fcbfc680beed934cdfa374656ab41afe
// Version: 0.1.3
package electra

import (
	"github.com/attestantio/go-eth2-client/spec/phase0"
	ssz "github.com/ferranbt/fastssz"
)

// MarshalSSZ ssz marshals the Attestation object
func (a *Attestation) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(a)
}

// MarshalSSZTo ssz marshals the Attestation object to a target array
func (a *Attestation) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(236)

	// Offset (0) 'AggregationBits'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(a.AggregationBits)

	// Field (1) 'Data'
	if a.Data == nil {
		a.Data = new(phase0.AttestationData)
	}
	if dst, err = a.Data.MarshalSSZTo(dst); err != nil {
		return
	}

	// Field (2) 'Signature'
	dst = append(dst, a.Signature[:]...)

	// Field (3) 'CommitteeBits'
	if size := len(a.CommitteeBits); size != 8 {
		err = ssz.ErrBytesLengthFn("Attestation.CommitteeBits", size, 8)
		return
	}
	dst = append(dst, a.CommitteeBits...)

	// Field (0) 'AggregationBits'
	if size := len(a.AggregationBits); size > 131072 {
		err = ssz.ErrBytesLengthFn("Attestation.AggregationBits", size, 131072)
		return
	}
	dst = append(dst, a.AggregationBits...)

	return
}

// UnmarshalSSZ ssz unmarshals the Attestation object
func (a *Attestation) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 236 {
		return ssz.ErrSize
	}

	tail := buf
	var o0 uint64

	// Offset (0) 'AggregationBits'
	if o0 = ssz.ReadOffset(buf[0:4]); o0 > size {
		return ssz.ErrOffset
	}

	if o0 < 236 {
		return ssz.ErrInvalidVariableOffset
	}

	// Field (1) 'Data'
	if a.Data == nil {
		a.Data = new(phase0.AttestationData)
	}
	if err = a.Data.UnmarshalSSZ(buf[4:132]); err != nil {
		return err
	}

	// Field (2) 'Signature'
	copy(a.Signature[:], buf[132:228])

	// Field (3) 'CommitteeBits'
	if cap(a.CommitteeBits) == 0 {
		a.CommitteeBits = make([]byte, 0, len(buf[228:236]))
	}
	a.CommitteeBits = append(a.CommitteeBits, buf[228:236]...)

	// Field (0) 'AggregationBits'
	{
		buf = tail[o0:]
		if err = ssz.ValidateBitlist(buf, 131072); err != nil {
			return err
		}
		if cap(a.AggregationBits) == 0 {
			a.AggregationBits = make([]byte, 0, len(buf))
		}
		a.AggregationBits = append(a.AggregationBits, buf...)
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the Attestation object
func (a *Attestation) SizeSSZ() (size int) {
	size = 236

	// Field (0) 'AggregationBits'
	size += len(a.AggregationBits)

	return
}

// HashTreeRoot ssz hashes the Attestation object
func (a *Attestation) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(a)
}

// HashTreeRootWith ssz hashes the Attestation object with a hasher
func (a *Attestation) HashTreeRootWith(hh ssz.HashWalker) (err error) {
	indx := hh.Index()

	// Field (0) 'AggregationBits'
	if len(a.AggregationBits) == 0 {
		err = ssz.ErrEmptyBitlist
		return
	}
	hh.PutBitlist(a.AggregationBits, 131072)

	// Field (1) 'Data'
	if a.Data == nil {
		a.Data = new(phase0.AttestationData)
	}
	if err = a.Data.HashTreeRootWith(hh); err != nil {
		return
	}

	// Field (3) 'Signature'
	hh.PutBytes(a.Signature[:])

	// Field (3) 'CommitteeBits'
	if size := len(a.CommitteeBits); size != 8 {
		err = ssz.ErrBytesLengthFn("Attestation.CommitteeBits", size, 8)
		return
	}
	hh.PutBytes(a.CommitteeBits)

	hh.Merkleize(indx)
	return
}

// GetTree ssz hashes the Attestation object
func (a *Attestation) GetTree() (*ssz.Node, error) {
	return ssz.ProofTree(a)
}