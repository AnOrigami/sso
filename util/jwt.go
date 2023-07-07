package util

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"

	"git.blauwelle.com/go/crate/log"
	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// 初始化JWT
func NewJWT(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) *JWT {
	if publicKey == nil {
		publicKey = &privateKey.PublicKey
	}
	return &JWT{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

// 加密过程
func (j *JWT) Sign(ctx context.Context, claims jwt.RegisteredClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(j.privateKey)
	if err != nil {
		log.Error(ctx, err.Error())
	}
	return tokenString, err
}

// 解密过程
func (j *JWT) Verify(ctx context.Context, tokenString string) (jwt.RegisteredClaims, error) {
	var claims jwt.RegisteredClaims
	_, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(token *jwt.Token) (interface{}, error) {
			return j.publicKey, nil
		},
		jwt.WithValidMethods([]string{"RS256"}),
	)
	if err != nil {
		log.Error(ctx, err.Error())
		return jwt.RegisteredClaims{}, err
	}
	return claims, nil
}

func NewJWTFromKeyBytes(keyBytes []byte) (*JWT, error) {
	// 解码给定的 PEM 数据，将其转换为一个 *pem.Block 结构
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	// 将传入的字节块解析为 PKCS1 格式的 RSA 私钥
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return NewJWT(key, nil), nil
}

//func init() {
//	bit := 2048
//	key, err := rsa.GenerateKey(rand.Reader, bit)
//	if err != nil {
//		panic(err)
//	}
//	b := x509.MarshalPKCS1PrivateKey(key)
//	f, _ := os.Create("private.rsa")
//	defer f.Close()
//	pem.Encode(f, &pem.Block{
//		Type:  "RSA PRIVATE KEY",
//		Bytes: b,
//	})
//	pem.Encode(os.Stdout, &pem.Block{
//		Type:  "RSA PUBLIC KEY",
//		Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey),
//	})
//}
