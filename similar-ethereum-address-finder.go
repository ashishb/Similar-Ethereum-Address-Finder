package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

// The following code takes inspiration from and generalizes the code at https://github.com/chrsow/geth-vanitygen

// Command line flag parsing
type stringsFlag struct {
	set   bool
	value []string
}

// set variable checks for the existence of a flag
func (sf *stringsFlag) Set(x string) error {
	sf.value = strings.Split(x, ",")
	sf.set = true
	return nil
}

func (sf *stringsFlag) String() string {
	return strings.Join(sf.value, ",")
}

// Word validation code
func validateWord(word string) {
	// Only accept lowercase to avoid the upper/lower case mismatches.
	r, _ := regexp.MatchString(`^[0-9a-f]+$`, word)
	if !r {
		fmt.Printf("[-] %s: is not a valid hexadecimal.\n", word)
		os.Exit(1)
	} else if len(word) > 40 {
		fmt.Println("[-] You can't generate matching Ethereum address for more than 40 characters (20 bytes).")
		os.Exit(1)
	}
}

func generateAccount() string {
	// 1. generate private key, ECDSA(private key)  => public key
	key, _ = crypto.GenerateKey()
	pubKey := key.PublicKey
	// 2. public key => address
	address := crypto.PubkeyToAddress(pubKey)
	addressHex := hex.EncodeToString(address[:])
	return addressHex
}

func searchAddress(prefix []string, suffix []string) (string, string) {
	n := len(prefix)
	if len(prefix) != len(suffix) {
		fmt.Printf("Length of prefix and suffix arrays doesn't match: %v %v\n", prefix, suffix);
		os.Exit(1)
	}

	prefixLengths := make([]int, len(prefix))
	// Small optimization when all prefixes have same length
	allPrefixesHaveSameLengths := true
	commonPrefixLength := len(prefix[0])
	for i, _ := range prefixLengths {
		prefixLengths[i] = len(prefix[i])
		if prefixLengths[i] != commonPrefixLength {
			allPrefixesHaveSameLengths = false
		}
	}

	suffixLengths := make([]int, len(suffix))
	// Small optimization when all suffixes have same length
	allSuffixesHaveSameLengths := true
	commonSuffixLength := len(suffix[0])

	for i, _ := range prefixLengths {
		suffixLengths[i] = len(suffix[i])
		if suffixLengths[i] != commonSuffixLength {
			allSuffixesHaveSameLengths = false
		}
	}

	found := false
	var address string
	count := 0

	for !found {
		if (count % 10000) == 0 {
			fmt.Printf("Attempt: %d\n", count)
		}
		count++
		address = generateAccount()
		var addressPrefix string
		var addressSuffix string
		if allPrefixesHaveSameLengths {
			addressPrefix = address[:commonPrefixLength]
		}
		if allSuffixesHaveSameLengths {
			addressSuffix = address[40-commonSuffixLength:]
		}
		for i := 0; i < n; i++ {
			if !allPrefixesHaveSameLengths {
				addressPrefix = address[:prefixLengths[i]]
			}
			if !allSuffixesHaveSameLengths {
				addressSuffix = address[40-suffixLengths[i]:]
			}
			//fmt.Printf("Checking \"%s (%s, %s)\" against \"%s\" and \"%s\"\n",
			//	address,
			//	addressPrefix,
			//	addressSuffix,
			//	prefix[i],
			//	suffix[i])
			if addressPrefix == prefix[i] && addressSuffix == suffix[i] {
				fmt.Printf("[+] Address with prefix \"%s\" and suffix \"%s\" found.\n", prefix[i], suffix[i])
				found = true
				break
			}
		}
	}

	privateKey := hex.EncodeToString(crypto.FromECDSA(key))
	return address, privateKey
}

func foundAddress(address string, privateKey string) {
	fmt.Printf("Address: 0x%s\n", address)
	fmt.Printf("PrivateKey: %s\n\n", privateKey)
}

// prefix, suffix from cli
var prefixes stringsFlag
var suffixes stringsFlag
var key *ecdsa.PrivateKey

func init() {
	flag.Var(&prefixes, "p", "Comma-separated list of prefixes")
	flag.Var(&suffixes, "s", "Comma-separated list of suffixes")
}


// Usage: -p 12,13,14 -s 89,678,56 -> This will try to find an address with
// 1. prefix = 12 and suffix = 89 or,
// 2. prefix = 13 and suffix = 678 or,
// 3. prefix = 14 and suffix = 56
// At least one of the -p and -s should be provided.
// If both are provided then they should have same number of elements.
// The program execution stops at the first match.
func main() {
	threadCount := 16
	var word string
	flag.Parse()
	ch := make(chan bool)
	for i, _ := range prefixes.value {
		prefixes.value[i] = strings.ToLower(prefixes.value[i])
		validateWord(prefixes.value[i])
	}
	for i, _ := range suffixes.value {
		suffixes.value[i] = strings.ToLower(suffixes.value[i])
		validateWord(suffixes.value[i])
	}
	// If prefixes are provided but not suffixes then init suffixes as empty array
	if prefixes.set && !suffixes.set {
		suffixes.value = make([]string, len(prefixes.value))
	}
	// If suffixes are provided but not suffixes then init prefixes as empty array
	if suffixes.set && !prefixes.set {
		prefixes.value = make([]string, len(suffixes.value))
	}
	fmt.Printf("Finding matches with prefixes = %v and suffixes = %v\n", prefixes.value, suffixes.value)
	for i := 0; i < threadCount; i++ {
		go findTheMatch(prefixes.value, suffixes.value, word, ch)
	}
	<-ch
}

func findTheMatch(prefixes []string, suffixes []string, word string, ch chan bool) {
	address, privateKey := searchAddress(prefixes, suffixes)
	foundAddress(address, privateKey)
	ch <- true
}
