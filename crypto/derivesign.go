package crypto

import (
	"fmt"
	"math/big"

	"github.com/adiabat/btcd/btcec"
	"github.com/adiabat/btcd/chaincfg/chainhash"
)

var (
	bigZero = new(big.Int).SetInt64(0)
)

// Computes sG, the signature multipled by the generator point, for an arbitrary message based on pubkey R and pubkey A
func ComputeP(oraclePubA, oraclePubR [33]byte, message []byte) ([33]byte, error) {
	return computePubKey(oraclePubA, oraclePubR, message)
}

// calculates P = pubR - h(msg, pubR)pubA
func computePubKey(pubA, pubR [33]byte, msg []byte) ([33]byte, error) {
	var returnValue [33]byte

	// Hardcode curve
	curve := btcec.S256()

	A, err := btcec.ParsePubKey(pubA[:], curve)
	if err != nil {
		return returnValue, err
	}

	R, err := btcec.ParsePubKey(pubR[:], curve)
	if err != nil {
		return returnValue, err
	}

	// e = Hash(messageType, oraclePubQ)
	var hashInput []byte
	hashInput = append(msg, R.X.Bytes()...)
	e := chainhash.HashB(hashInput)

	bigE := new(big.Int).SetBytes(e)

	if bigE.Cmp(curve.N) >= 0 {
		return returnValue, fmt.Errorf("hash of (msg, pubR) too big")
	}

	// e * B
	A.X, A.Y = curve.ScalarMult(A.X, A.Y, e)

	A.Y.Neg(A.Y)

	A.Y.Mod(A.Y, curve.P)

	P := new(btcec.PublicKey)

	// add to R
	P.X, P.Y = curve.Add(A.X, A.Y, R.X, R.Y)
	copy(returnValue[:], P.SerializeCompressed())
	return returnValue, nil
}

// Computes s, the signature for an arbitrary message based on private scalars k and a
func ComputeS(a, k [32]byte, message []byte) ([32]byte, error) {
	return computePrivKey(k, a, message)
}

func computePrivKey(k, a [32]byte, msg []byte) ([32]byte, error) {
	var empty, s [32]byte

	// Hardcode curve
	curve := btcec.S256()

	bigPriv := new(big.Int).SetBytes(a[:])
	a = empty
	bigK := new(big.Int).SetBytes(k[:])

	if bigPriv.Cmp(bigZero) == 0 {
		return empty, fmt.Errorf("priv scalar is zero")
	}
	if bigPriv.Cmp(curve.N) >= 0 {
		return empty, fmt.Errorf("priv scalar is out of bounds")
	}
	if bigK.Cmp(bigZero) == 0 {
		return empty, fmt.Errorf("k scalar is zero")
	}
	if bigK.Cmp(curve.N) >= 0 {
		return empty, fmt.Errorf("k scalar is out of bounds")
	}

	// re-derive R = kG
	var Rx *big.Int
	Rx, _ = curve.ScalarBaseMult(k[:])

	// Ry is always even.  Make it even if it's not.
	//	if Ry.Bit(0) == 1 {
	//		bigK.Mod(bigK, curve.N)
	//		bigK.Sub(curve.N, bigK)
	//	}

	// e = Hash(r, m)

	e := chainhash.HashB(append(msg[:], Rx.Bytes()...))
	bigE := new(big.Int).SetBytes(e)

	// If the hash is bigger than N, fail.  Note that N is
	// FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141
	// So this happens about once every 2**128 signatures.
	if bigE.Cmp(curve.N) >= 0 {
		return empty, fmt.Errorf("hash of (m, R) too big")
	}
	//	fmt.Printf("e: %x\n", e)
	// s = k + e*a
	bigS := new(big.Int)
	// e*a
	bigS.Mul(bigE, bigPriv)
	// k + (e*a)
	bigS.Sub(bigK, bigS)
	bigS.Mod(bigS, curve.N)

	// check if s is 0, and fail if it is.  Can't see how this would happen;
	// looks like it would happen about once every 2**256 signatures
	if bigS.Cmp(bigZero) == 0 {
		str := fmt.Errorf("sig s %v is zero", bigS)
		return empty, str
	}

	// Zero out private key and k in array and bigint form
	// who knows if this really helps...  can't hurt though.
	bigK.SetInt64(0)
	k = empty
	bigPriv.SetInt64(0)

	byteOffset := (256 - bigS.BitLen()) / 8

	copy(s[byteOffset:], bigS.Bytes())

	return s, nil
}
