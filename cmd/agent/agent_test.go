package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"os"
	"sync"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func Test_collectorAgent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	collector := NewMockCollector(ctrl)
	collector.EXPECT().Collect().AnyTimes()
	collector.EXPECT().AllMetrics().AnyTimes()
	collector.EXPECT().ResetCounter("PollCount").AnyTimes()

	sender := NewMockSender(ctrl)
	sender.EXPECT().Send(gomock.Any()).AnyTimes()
	sender.EXPECT().Close()

	agent := newCollectorAgent(
		collector, sender,
		newDelay(time.Second*2, time.Second*1))

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	agent.run(ctx, wg)

	time.Sleep(time.Second * 5)

	cancel()

	time.Sleep(time.Second * 2)

}

func Test_createSender(t *testing.T) {
	hash := sha256.New().Sum([]byte("test"))

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)

	defer os.Remove("hash.key")
	defer os.Remove("cert.pub")

	var publicKeyPEM bytes.Buffer
	pem.Encode(&publicKeyPEM, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	})
	err = os.WriteFile("cert.pub", publicKeyPEM.Bytes(), 0644)
	require.NoError(t, err)

	err = os.WriteFile("hash.key", hash, 0644)
	require.NoError(t, err)

	_, err = createSender("hash.key", "cert.pub", "localhost:8080", 1)
	require.NoError(t, err)

}
