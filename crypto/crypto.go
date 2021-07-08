package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"math/big"
)

func ParseEcdsaPublicKeyFromPem(publicPEM string) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicPEM))
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *ecdsa.PublicKey:
		return pub, nil
	default:
		break
	}
	return nil, errors.New("key type is not ECDSA")
}

func ParseEcdsaPrivateKeyFromPem (privatePEM string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privatePEM))
	if block == nil || block.Type != "EC PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing ec private key")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func GetPublicECDSAKeyFromCompressedAddress (compressedAddress string) (*ecdsa.PublicKey, error) {

	x, y, err := DeriveYFromCompressed(compressedAddress)
	if err != nil {
		return nil, err
	}
	key := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}
	return key, nil
}

func DeriveYFromCompressed(compressedAddress string) (*big.Int, *big.Int, error) {

	compressedBytes, err := hex.DecodeString(compressedAddress)
	if err != nil {
		return nil, nil, err
	}
	// Addresses consist of: 1 firstByte sign firstByte + 32 bytes X
	if len(compressedBytes) != 33 {
		return nil, nil, errors.New("invalid address length")
	}

	// Split the sign firstByte from the rest
	signByte := uint(compressedBytes[0])
	if signByte != 2 && signByte != 3 {
		return nil, nil, errors.New("invalid sign byte")
	}
	xBytes := compressedBytes[1:]

	// Convert to big Int.
	x := new(big.Int).SetBytes(xBytes)

	// We use Constant 3 a couple of times
	constant := big.NewInt(3)

	// The params for P256
	curveParams := elliptic.P256().Params()

	// The equation is y^2 = x^3 - 3x + b
	// x^3 mod P
	xCubed := new(big.Int).Exp(x, constant, curveParams.P)

	// 3x mod P
	threeX := new(big.Int).Mul(x, constant)
	threeX.Mod(threeX, curveParams.P)

	// x^3 - 3x
	ySquared := new(big.Int).Sub(xCubed, threeX)

	// b mod P
	ySquared.Add(ySquared, curveParams.B)
	ySquared.Mod(ySquared, curveParams.P)

	// Now we need to find the square root mod P.
	// This is where Go's big int library redeems itself.
	y := new(big.Int).ModSqrt(ySquared, curveParams.P)
	if y == nil {
		// If this happens then you're dealing with an invalid point.
		// Panic, return an error, whatever you want.
		return nil, nil, errors.New("invalid point")
	}
	if y.Bit(0) != signByte & 1 {
		y.Neg(y)
		y.Mod(y, curveParams.P)
	}
	return x, y, nil
}

func VerifyECDSASignature(publicKey *ecdsa.PublicKey, hashed string, signature string) (bool, error) {
	hashBytes, err := hex.DecodeString(hashed)
	if err != nil {
		return false, err
	}
	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false, err
	}
	return ecdsa.VerifyASN1(publicKey, hashBytes, signatureBytes), nil
}
