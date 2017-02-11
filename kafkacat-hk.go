package main

import (
	"bytes"
	"encoding/base64"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const DEFAULT_KAFKACAT_BIN = "/usr/bin/kafkacat"

// kafkacat-hk wraps kafkacat and automatically passes SSL arguments and broker URLs based on Herkou-style Kafka env vars
func main() {
	exe := os.Getenv("KAFKACAT_BIN") // default /usr/bin/kafkacat
	if exe == "" {
		exe = DEFAULT_KAFKACAT_BIN
	}

	_, err := os.Stat(exe)
	if err != nil {
		log.Fatalf("%s does not exist. Please set KAFKACAT_BIN to the location of your kafkacat binary\n", DEFAULT_KAFKACAT_BIN)
	}

	ca, crt, key := loadCertsFromEnv()

	caPipe := makePipeFromBytes(ca)
	defer caPipe.Close()

	crtPipe := makePipeFromBytes(crt)
	defer crtPipe.Close()

	keyPipe := makePipeFromBytes(key)
	defer keyPipe.Close()

	cmdFields := sslArgs()
	if os.Getenv("KAFKA_URL") != "" {
		cmdFields = append(cmdFields, "-b", strings.Replace(os.Getenv("KAFKA_URL"), "kafka://", "", -1))
	}
	cmdFields = append(cmdFields, os.Args[1:]...)

	cmd := exec.Command(exe, cmdFields...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// automatically passed to child as fd 3,4,5
	cmd.ExtraFiles = []*os.File{caPipe, crtPipe, keyPipe}

	err = cmd.Run()
	if err, ok := err.(*exec.ExitError); ok {
		if status, ok := err.Sys().(syscall.WaitStatus); ok {
			os.Exit(status.ExitStatus())
		}
	}
}

func sslArgs() []string {
	return []string{"-X", "security.protocol=ssl", "-X", "ssl.ca.location=/dev/fd/3", "-X", "ssl.certificate.location=/dev/fd/4", "-X", "ssl.key.location=/dev/fd/5"}
}

func makePipeFromBytes(input []byte) *os.File {
	r, w, err := os.Pipe()
	if err != nil {
		log.Fatal(err)
	}
	rIn := bytes.NewReader(input)
	go func() {
		io.Copy(w, rIn)
		w.Close()
	}()
	return r
}

func loadCertsFromEnv() ([]byte, []byte, []byte) {
	// I'm using base64 encoded ENV to avoid newlines in a systemd config, but I've made it configurable - Heroku's aren't encoded and don't have quotes, so we can just return them
	if os.Getenv("HEROKU") != "" {
		return []byte(os.Getenv("KAFKA_TRUSTED_CERT")), []byte(os.Getenv("KAFKA_CLIENT_CERT")), []byte(os.Getenv("KAFKA_CLIENT_CERT_KEY"))
	}
	// remove quotes if they are set on the ENV for some reason (don't ask)
	cert_b64 := strings.TrimSuffix(strings.TrimPrefix(os.Getenv("KAFKA_CLIENT_CERT"), `"`), `"`)
	key_b64 := strings.TrimSuffix(strings.TrimPrefix(os.Getenv("KAFKA_CLIENT_CERT_KEY"), `"`), `"`)
	ca_b64 := strings.TrimSuffix(strings.TrimPrefix(os.Getenv("KAFKA_TRUSTED_CERT"), `"`), `"`)
	if cert_b64 == "" || key_b64 == "" || ca_b64 == "" {
		log.Fatal("Must set KAFKA_CLIENT_CERT, KAFKA_CLIENT_CERT_KEY, KAFKA_TRUSTED_CERT env")
	}
	crt, err := base64.StdEncoding.DecodeString(cert_b64)
	if err != nil {
		log.Fatal(err)
	}
	key, err := base64.StdEncoding.DecodeString(key_b64)
	if err != nil {
		log.Fatal(err)
	}
	ca, err := base64.StdEncoding.DecodeString(ca_b64)
	if err != nil {
		log.Fatal(err)
	}
	return ca, crt, key
}
