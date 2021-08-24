package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

// The following code takes inspiration from and generalizes the code at https://github.com/chrsow/geth-vanitygen

// Command line flag parsing
type stringsFlag struct {
	set bool
	value []string
}

// Set checks variable for the existence of a flag
func (sf *stringsFlag) Set(x string) error {
	sf.value = strings.Split(x, ",")
	sf.set = true
	return nil
}

func (sf *stringsFlag) String() string {
	return strings.Join(sf.value, ",")
}

type IntFlag struct {
	set bool
	value int
}

func (intFlag *IntFlag) Set(x string) error {
	value, error := strconv.Atoi(x)
	if error != nil {
		return error
	}
	intFlag.value = value
	intFlag.set = true
	return nil
}

func (intFlag *IntFlag) String() string {
	return string(intFlag.value)
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
	key, _ := crypto.GenerateKey()
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
		if (count % 50000) == 0 {
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
var threadCount IntFlag
var key *ecdsa.PrivateKey

const defaultThreadCount = 16

func init() {
	flag.Var(&prefixes, "p", "Comma-separated list of prefixes")
	flag.Var(&suffixes, "s", "Comma-separated list of suffixes")
	flag.Var(&threadCount, "t", fmt.Sprintf("Num threads (default: %d)", defaultThreadCount))
}


// Usage: -p 12,13,14 -s 89,678,56 -> This will try to find an address with
// 1. prefix = 12 and suffix = 89 or,
// 2. prefix = 13 and suffix = 678 or,
// 3. prefix = 14 and suffix = 56
// At least one of the -p and -s should be provided.
// If both are provided then they should have same number of elements.
// The program execution stops at the first match.
func main() {
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
	numThreads := defaultThreadCount
	if threadCount.set {
		numThreads = threadCount.value
	}
	printAttemptEstimates(prefixes.value, suffixes.value, numThreads)
	for i := 0; i < numThreads; i++ {
		go findTheMatch(prefixes.value, suffixes.value, word, ch)
	}
	<-ch
}

func findTheMatch(prefixes []string, suffixes []string, word string, ch chan bool) {
	address, privateKey := searchAddress(prefixes, suffixes)
	foundAddress(address, privateKey)
	ch <- true
}

func printAttemptEstimates(prefixes []string, suffixes []string, threadCount int) {
	harmonicSum := 0.0
	for i, _ := range prefixes {
		numNibbles := len(prefixes[i]) + len(suffixes[i])
		numBits := 4 * numNibbles
		numAttempts := math.Pow(2, float64(numBits))
		harmonicSum += 1/float64(numAttempts)
		fmt.Printf(
			"It will take %.1f attempts for finding a ETH address matching (prefix: \"%s\",suffix: \"%s\") " +
				"with 100%% probability.\n\t%.1f attempts suffice for 50%% probability of finding a match.\n",
			numAttempts, prefixes[i], suffixes[i], numAttempts / 2)
	}
	numAttempts := int(1.0 / harmonicSum) / threadCount
	fmt.Printf("Overall number of attempts across all pairs is %d for 100%% probability and %d for 50%% probability\n",
		numAttempts, numAttempts / 2)
}
