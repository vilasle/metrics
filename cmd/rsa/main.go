package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"

	"github.com/spf13/pflag"
)

type cliArgs struct {
	name   string
	length int32
	help   bool
}

func parseCliArgs() cliArgs {
	cli := &cliArgs{}

	pflag.Int32VarP(&cli.length, "length", "l", 4096, "Length of the key")
	pflag.StringVarP(&cli.name, "name", "n", "metric", "Name of keys")
	pflag.BoolVarP(&cli.help, "help", "h", false, "Show help")
	pflag.Parse()

	return *cli
}

func main() {

	cli := parseCliArgs()

	if cli.help {
		pflag.Usage()
		return
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, int(cli.length))
	if err != nil {
		log.Fatal(err)
	}

	var privateKeyPEM bytes.Buffer
	pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	var publicKeyPEM bytes.Buffer
	pem.Encode(&publicKeyPEM, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	})

	privatePath, publicPath := cli.name, fmt.Sprintf("%s.pub", cli.name)

	if err := os.WriteFile(privatePath, privateKeyPEM.Bytes(), 0644); err != nil {
		log.Fatalf("can not write private key to '%s' by reason %v", privatePath, err)
	}

	if err := os.WriteFile(publicPath, publicKeyPEM.Bytes(), 0644); err != nil {
		log.Fatalf("can not write public key to '%s' by reason %v", publicPath, err)
	}
}