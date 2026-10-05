package handlers

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func createSSHClient(host string, port int, username, password, privateKey string) (*ssh.Client, error) {
	var authMethods []ssh.AuthMethod

	if privateKey != "" {
		signer, err := parsePrivateKey(privateKey)
		if err == nil {
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		}
	}

	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("没有可用的认证方式")
	}

	config := &ssh.ClientConfig{
		User:            username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}

	return client, nil
}

func parsePrivateKey(key string) (signer ssh.Signer, err error) {
	var signerInterface interface{}

	if strings.Contains(key, "BEGIN OPENSSH PRIVATE KEY") || strings.Contains(key, "BEGIN PRIVATE KEY") {
		signerInterface, err = ssh.ParsePrivateKey([]byte(key))
	} else if strings.Contains(key, "BEGIN RSA PRIVATE KEY") || strings.Contains(key, "BEGIN DSA PRIVATE KEY") || strings.Contains(key, "BEGIN EC PRIVATE KEY") {
		signerInterface, err = ssh.ParsePrivateKeyWithPassphrase([]byte(key), []byte{})
	} else {
		signerInterface, err = ssh.ParsePrivateKey([]byte(key))
	}

	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %w", err)
	}
	return signerInterface.(ssh.Signer), nil
}
