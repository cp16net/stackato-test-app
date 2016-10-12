package common

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/crypto/openpgp/packet"
)

// VerifySignature takes a key id, file and signature and verifies the signature is valid
func VerifySignature(pubKeyID string, file *os.File, sigFile *os.File) error {
	fileContent, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	pack, err := packet.Read(sigFile)
	if err != nil {
		return err
	}

	signature, ok := pack.(*packet.Signature)
	if !ok {
		return errors.New("Provided signature file is not a valid signature file.")
	}

	reader := strings.NewReader(pubKeyID)
	pack, err = packet.Read(reader)
	if err != nil {
		return err
	}
	publicKey, ok := pack.(*packet.PublicKey)
	if !ok {
		return errors.New("Provided key is not a valid public key. Please contact administrator")
	}

	hash := signature.Hash.New()
	_, err = hash.Write(fileContent)
	if err != nil {
		return err
	}

	return publicKey.VerifySignature(hash, signature)

}
