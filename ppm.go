/*
 * TODO Refactor
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
	"sort"
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

	fmt.Println("Master password")
	password, err := getPass()
	if err != nil {
		log.Fatal(err)
	}
	key := sha256.Sum256([]byte(password))

	err = extract(dataPath, key[:])
	if err != nil {
		fmt.Println("Right password?")
		log.Fatal(err)
	}
	switch os.Args[2] {
	case "set":
		err := set(os.Args[3])
		if err != nil {
			log.Fatal(err)
		}
		err = save(os.Args[1], key[:])
		if err != nil {
			log.Fatal(err)
		}
	case "get":
		fmt.Println(get(os.Args[3]))
	case "gen":
		len, err := strconv.Atoi(os.Args[4])
		if err != nil {
			log.Fatal(err)
		}
		err = gen(os.Args[3], len)
		if err != nil {
			log.Fatal(err)
		}
		err = save(os.Args[1], key[:])
		if err != nil {
			log.Fatal(err)
		}
	case "exp":
		err := exp()
		if err != nil {
			log.Fatal(err)
		}
	case "imp":
		err := imp(dataPath, key[:])
		if err != nil {
			log.Fatal(err)
		}
	case "list":
		fmt.Print(list())
	case "pipe":
		/* Set default pipe path */
		pipe, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		pipe += "/.ppm"
		if checkFileNotExists(pipe) {
			err = os.Mkdir(pipe, 0777)
			if err != nil {
				log.Fatal(err)
			}
		}
		pipe += "/pipe"
		/* Create named pipe */
		if checkFileNotExists(pipe) {
			out := exec.Command("mkfifo", pipe)
			err = out.Run()
			if err != nil {
				log.Println(out.Run())
			}
		}
		err = handlePipe(pipe)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func handlePipe(pipe string) error {
	var err error
	var account []byte
	var accountStr string
	/* Equals 3 passwords with the example cliente */
	for {
		account, err = ioutil.ReadFile(pipe)
		accountStr = strings.Replace(string(account), "\n", "", -1)
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
/* Non echoing prompt */
func getPass() (string, error) {
	prompt := promptui.Prompt{
		Label: "Enter",
		Mask:  '+',
	}
	pass, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return pass, err
}

/* Encrypts a byte array */
func enc(key, plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cipherText := make([]byte, len(plainText)+block.BlockSize())

	iv := cipherText[:block.BlockSize()]
	_, err = rand.Read(iv)
	if err != nil {
		return nil, err
	}

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
func set(account string) error {
	var err error
	if err != nil {
		return err
	}
	fmt.Println()
	accounts[account], err = getPass()
	if err != nil {
		return err
	}
	return nil
}

/* Prints account data */
func get(account string) string {
	return string(accounts[account])
}

/* Generates random account data */
func gen(account string, length int) error {
	word := ""
	/* make default ascii list */
	var ascii string
	for i := 0x21; i < 0x7f; i++ {
		ascii += string(i)
	}
	/* make nice random string */
	seedSlice := make([]byte, 8)
	_, err := rand.Read(seedSlice)
	if err != nil {
		return err
	}
	var seed int64
	for i, v := range seedSlice {
		seed += int64(v << i * 8)
	}
	mrand.Seed(int64(seed))

	skip := make([]byte, 1)
	for i := 0; i < length; i++ {
		_, err = rand.Read(skip)
		if err != nil {
			return err
		}
		for k := byte(0); k < skip[0]; k++ {
			mrand.Int()
		}
		word += string(ascii[mrand.Int()%len(ascii)])
	}
	accounts[account] = word
	return nil
}

/* For complete data set */
/* Prints all data as plain text */
func exp() error {
	jsonData, err := json.Marshal(accounts)
	if err != nil {
		return err
	}
	fmt.Print(string(jsonData))
	fmt.Println()
	return nil
}

/* Imports plain text into specified file (if exists adds and overwrites it) */
func imp(dataPath string, key []byte) error {
	var impData string
	var impAccounts map[string]string
	fmt.Scan(&impData)
	err := json.Unmarshal([]byte(impData), &impAccounts)
	if err != nil {
		return err
	}
	for k, v := range impAccounts {
		accounts[k] = v
	}
	save(dataPath, key[:])
	return nil
}

/* Prints a list of all accounts */
func list() string {
	var strRand []string
	var strSorted string
	for k := range accounts {
		strRand = append(strRand, k+"\n")
	}
	sort.Strings(strRand)
	for _, v := range strRand {
		strSorted += v
	}
	return strSorted
}

/* Extract data from file */
func extract(dataPath string, key []byte) error {
	if !checkFileNotExists(dataPath) {
		cipherText, err := ioutil.ReadFile(dataPath)
		if err != nil {
			return err
		}
		jsonData, err := dec(key[:], cipherText)
		if err != nil {
			return err
		}
		err = json.Unmarshal(jsonData, &accounts)
		if err != nil {
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
	err = ioutil.WriteFile(dataPath, cipherText, 0660)
	if err != nil {
		return err
	}
	return nil
}

/* Return true if file does not exist */
func checkFileNotExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return os.IsNotExist(err)
}
