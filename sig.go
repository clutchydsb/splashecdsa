// MIT License

// Copyright (c) 2019 Kassius Barker <kasscrypto@gmail.com>

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package splashecdsa

import (
	"crypto/elliptic"
	"math/big"

	"github.com/CryptoKass/splashecdsa/ecmath"
)

// SplashSignature an ECDSA signature which allows for
// Multisig and PublicKey reconstruction.
type SplashSignature struct {
	R, S *big.Int
	V, O uint8 // V is a reconstruction flag and O a multi sig order
}

// GetY will calculate the two possible Y values for a given X value on the
// Curve.
func GetY(x *big.Int, curve *elliptic.CurveParams) (*big.Int, *big.Int) {
	x3 := new(big.Int).Mul(x, x)
	x3.Mul(x3, x)

	threeX := new(big.Int).Lsh(x, 1)
	threeX.Add(threeX, x)

	x3.Sub(x3, threeX)
	x3.Add(x3, curve.B)
	x3.Mod(x3, curve.P)

	y1 := new(big.Int).ModSqrt(x3, curve.P)

	y2 := new(big.Int).Sub(curve.P, y1)
	y2.Mod(y2, curve.P)

	//return x3.ModSqrt(x3, curve.P), x3.Sub(curve.P, x3).Mod(x3, curve.P)
	return y1, y2
}

// ReconstructPublicKey reconstructs a public from a signature
// using the message hash and a given Curve. Follows the
// forumula pub = r^-1(sR−zG)`
func (sig *SplashSignature) ReconstructPublicKey(msgHash []byte, C elliptic.Curve) (pub SplashPublicKey) {

	// get the curve parameters
	curve := C.Params()

	// calulate point K this is also known as point R
	// in the formula: `pub = r^-1(sR−zG)`.
	// K is the point where x = r, however due to the nature
	// of elliptic curves this point will have 2 possible
	// values. The signatures V value will determin which of
	// these values we should use...
	kX := sig.R
	kY, potentialY := ecmath.GetY(sig.R, curve)

	// which y is the correct one
	if sig.V == 1 {
		kY = potentialY
	}

	// calculate sK and zG points
	sKx, sKy := curve.ScalarMult(kX, kY, sig.S.Bytes())
	zGx, zGy := curve.ScalarBaseMult(msgHash[:32]) // z the first 32 bytes of the data

	// subtract zk from sK by inverting zK and adding
	izGx, izGy := ecmath.InversePoint(zGx, zGy, curve)
	XX, YY := curve.Add(sKx, sKy, izGx, izGy)

	// finally multiply XX, YY by r mod p
	RMod := new(big.Int).ModInverse(sig.R, curve.N)
	x, y := curve.ScalarMult(XX, YY, RMod.Bytes())

	// wrap the values in SplashPublicKey
	pub.X, pub.Y = x, y
	pub.Curve = C

	return
}
