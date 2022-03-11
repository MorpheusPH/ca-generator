package main

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"
)

const (
	organizationKey = "ORGANIZATION"
	serviceNameKey  = "SERVICENAME"
	namespaceKey    = "NAMESPACE"
)

func main() {
	var caPEM, serverCertPEM, serverPrivKeyPEM *bytes.Buffer
	var dnsNames []string
	if os.Getenv(organizationKey) == "" {
		log.Printf("environment variable %s is not set! Please set it.", organizationKey)
		os.Exit(1)
	}

	if os.Getenv(serviceNameKey) == "" {
		log.Printf("environment variable %s is not set! Please set it.", serviceNameKey)
		os.Exit(1)
	}

	if os.Getenv(namespaceKey) == "" {
		log.Printf("environment variable %s is not set! Please set it.", namespaceKey)
		os.Exit(1)
	}

	organization := os.Getenv(organizationKey)
	service_name := os.Getenv(serviceNameKey)
	namespace := os.Getenv(namespaceKey)

	// CA config
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2020),
		Subject: pkix.Name{
			Organization: []string{organization},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// CA private key
	caPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		fmt.Println(err)
	}

	// Self signed CA certificate
	caBytes, err := x509.CreateCertificate(cryptorand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		fmt.Println(err)
	}

	// PEM encode CA cert
	caPEM = new(bytes.Buffer)
	_ = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	// dnsNames := []string{"nfs-subdir-webhook-service",
	// 	"nfs-subdir-webhook-service.nfs-subdir-webhook-system", "nfs-subdir-webhook-service.nfs-subdir-webhook-system.svc"}
	commonName := fmt.Sprintf("%s.%s.%s", service_name, namespace, "svc")
	dnsNames = append(dnsNames, service_name)
	dnsNames = append(dnsNames, fmt.Sprintf("%s.%s", service_name, namespace))
	dnsNames = append(dnsNames, commonName)

	// server cert config
	cert := &x509.Certificate{
		DNSNames:     dnsNames,
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{organization},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// server private key
	serverPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		fmt.Println(err)
	}

	// sign the server cert
	serverCertBytes, err := x509.CreateCertificate(cryptorand.Reader, cert, ca, &serverPrivKey.PublicKey, caPrivKey)
	if err != nil {
		fmt.Println(err)
	}

	// PEM encode the  server cert and key
	serverCertPEM = new(bytes.Buffer)
	_ = pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})

	serverPrivKeyPEM = new(bytes.Buffer)
	_ = pem.Encode(serverPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	})

	//err = os.MkdirAll("./certs/", 0666)
	//if err != nil {
	//	log.Panic(err)
	//}
	err = WriteFile("tls.crt", serverCertPEM)
	if err != nil {
		log.Panic(err)
	}

	err = WriteFile("tls.key", serverPrivKeyPEM)
	if err != nil {
		log.Panic(err)
	}

}

// WriteFile writes data in the file at the given path
func WriteFile(filepath string, sCert *bytes.Buffer) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(sCert.Bytes())
	if err != nil {
		return err
	}
	return nil
}
