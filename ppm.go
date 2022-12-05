/* TODO
 * Export function
 * Import function
 * Timers for security
 */
package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	mrand "math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
)

var accounts map[string]string

func main() {
	if len(os.Args) < 3 {
		println("You have to specify a file and an argument!")
		return
	}
	dataPath := os.Args[1]
	accounts = make(map[string]string)

	password := getPass()
	key := sha256.Sum256([]byte(password))

	err := extract(dataPath, key[:])
	if err != nil {
		log.Fatal(err)
	}
	switch os.Args[2] {
	case "set":
		set(os.Args[3])
		save(os.Args[1], key[:])
	case "get":
		fmt.Println(get(os.Args[3]))
	case "gen":
		len, err := strconv.Atoi(os.Args[4])
		if err != nil {
			log.Fatal(err)
		}
		gen(os.Args[3], len)
		save(os.Args[1], key[:])
	case "exp":
		exp()
	case "imp":
		imp()
	case "list":
		fmt.Print(list())
		fmt.Println([]byte(list()))
	case "pipe":
		/* Set default pipe path */
		pipe, _ := os.UserHomeDir()
		pipe += "/.ppm"
		os.Mkdir(pipe, 0777)
		pipe += "/pipe"
		/* Create named pipe */
		out := exec.Command("mkfifo", pipe)
		out.Run()

		handlePipe(pipe)
	}
}

func handlePipe(pipe string) error {
	var account []byte
	var accountStr string
	var err error
	for {
		account, err = ioutil.ReadFile(pipe)
		accountStr = strings.Replace(string(account), "\n", "", -1)
		fmt.Println(accountStr)
		if err != nil {
			return err
		}
		if string(accountStr) == "list" {
			ioutil.WriteFile(pipe, []byte(list()), 0660)

		} else {
			ioutil.WriteFile(pipe, []byte(get(string(accountStr))), 0660)
		}
	}
	return nil
}

/* Encryption */
/* Prompts for the master password */
func getPass() string {
	prompt := promptui.Prompt{
		Label: "Enter Master Password: ",
		Mask:  '+',
	}
	pass, err := prompt.Run()
	if err != nil {
		log.Fatal("Can't read master password!")
	}
	return pass
}

/* Encrypts a byte array */
func enc(key, plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cipherText := make([]byte, len(plainText)+block.BlockSize())

	iv := cipherText[:block.BlockSize()]
	rand.Read(iv)

	encrypter := cipher.NewCFBEncrypter(block, iv)
	encrypter.XORKeyStream(cipherText[block.BlockSize():], plainText)

	return cipherText, err
}

/* Decrypts a string */
func dec(key, cipherText []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plainText := make([]byte, len(cipherText)-block.BlockSize())

	iv := cipherText[:block.BlockSize()]

	decrypter := cipher.NewCFBDecrypter(block, iv)
	decrypter.XORKeyStream(plainText, cipherText[block.BlockSize():])

	return plainText, err
}

/* Account specific */
/* Sets account data */
func set(account string) {
	fmt.Println("Give the password here: ")
	accounts[account] = getPass()
}

/* Prints account data */
func get(account string) string {
	return string(accounts[account])
}

/* Generates random account data */
func gen(account string, length int) {
	word := ""
	/* make default ascii list */
	var ascii string
	for i := 0x21; i < 0x7f; i++ {
		ascii += string(i)
	}
	/* make nice random string */
	seedSlice := make([]byte, 8)
	rand.Read(seedSlice)
	var seed int64
	for i, v := range seedSlice {
		seed += int64(v << i * 8)
	}
	mrand.Seed(int64(seed))

	skip := make([]byte, 1)
	for i := 0; i < length; i++ {
		rand.Read(skip)
		for k := byte(0); k < skip[0]; k++ {
			mrand.Int()
		}
		word += string(ascii[mrand.Int()%len(ascii)])
	}
	accounts[account] = word
}

/* For complete data set */
/* Prints all data as plain text */
func exp() {

}

/* Imports plain text into specified file (if exists adds and overwrites it) */
func imp() {

}

/* Prints a list of all accounts */
func list() string {
	var str string
	for k := range accounts {
		str += k + "\n"
	}
	return str
}

/* Extract data from file */
func extract(dataPath string, key []byte) error {
	/* Decryption of database */
	cipherText, err := ioutil.ReadFile(dataPath)
	if err == nil {
		jsonData, err := dec(key[:], cipherText)
		if err != nil {
			return err
		}
		err = json.Unmarshal(jsonData, &accounts)
		if err != nil {
			return err
		}
	} else {
		noFileErr := "open " + dataPath + ": no such file or directory"
		if fmt.Sprint(err) == noFileErr {
			return nil
		} else {
			return err
		}
	}
	return nil
}
func save(dataPath string, key []byte) error {
	/* clean data */
	delete(accounts, "")
	for k, v := range accounts {
		if v == "" {
			delete(accounts, k)
		}
	}

	jsonData, err := json.Marshal(accounts)
	if err != nil {
		return err
	}
	cipherText, err := enc(key[:], jsonData)
	if err != nil {
		return err
	}
	ioutil.WriteFile(dataPath, cipherText, 0660)
	return nil
}
