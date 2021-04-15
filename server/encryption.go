package fileserver

import (
	"github.com/DAv10195/submit_commons/encryption"
	"github.com/spf13/viper"
	"path/filepath"
)

var fsEncryption encryption.Encryption

func InitFsEncryption() error {
	path := filepath.Join(viper.GetString("encryption-key-path"),"submit_server.key")
	if err := encryption.GenerateAesKeyFile(path); err != nil {
		return err
	}
	fsEncryption = &encryption.AesEncryption{KeyFilePath: path}
	return nil
}

func Decrypt(encryptedText string) (string, error) {
	decryptedText, err := fsEncryption.Decrypt(encryptedText)
	if err != nil {
		return "", err
	}
	return decryptedText, nil
}

func Encrypt(unEncryptedText string) (string, error) {
	encryptedText, err := fsEncryption.Encrypt(unEncryptedText)
	if err != nil {
		return "", err
	}
	return encryptedText, nil
}

