package PtUtil

/**
 * Title:Des加解密工具类
 * User: iuoon
 * Date: 2016-9-23
 * Version: 1.0
 */
import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
)

//定义向量
var iv = "15425832"

//定义密钥
var key = "hxxczcwczykjhxxczcwczykj"

//DesEncrypt 3des cbc加密函数，返回加密后的结果长度是8的倍数
func DesEncrypt(origData []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	origData = pKCS5Padding(origData, block.BlockSize())
	// origData = ZeroPadding(origData, block.BlockSize())

	blockMode := cipher.NewCBCEncrypter(block, []byte(iv))
	crypted := make([]byte, len(origData))
	// 根据CryptBlocks方法的说明，如下方式初始化crypted也可以
	// crypted := origData
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func pKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//DesDecrypt 3des cbc解密函数，传入解密内容长度必须是8的倍数
func DesDecrypt(crypted []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, []byte(iv))
	origData := make([]byte, len(crypted))
	// origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	origData = pKCS5UnPadding(origData)
	// origData = ZeroUnPadding(origData)
	return origData, nil
}

func pKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
