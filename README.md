# Installation (optional)

```
go get github.com/ashishb/Similar-Ethereum-Address-Finder
```

# Sample usage

Following command will try to find an Ethereum address with prefix/suffix pairs as (57, 5ab), (63, 8c), and (69, 457). Number of threads executed in parallel would be 32. The code will stop at the first match. 

```
$ similar-ethereum-address-finder -p 57,63,69 -s 5Ab,8c,457 -t 32
```

Alternatively, if you have cloned the repository.

```
$ go run similar-ethereum-address-finder.go -p 57,63,69 -s 5Ab,8c,457 -t 32

Finding matches with prefixes = [57 63 69] and suffixes = [5ab 8c 457]
It will take 1048576 attempts for finding a ETH address matching (prefix: "57",suffix: "5ab") with 100% probability.
	524288 attempts suffice for 50% probability of finding a match.
It will take 65536 attempts for finding a ETH address matching (prefix: "63",suffix: "8c") with 100% probability.
	32768 attempts suffice for 50% probability of finding a match.
It will take 1048576 attempts for finding a ETH address matching (prefix: "69",suffix: "457") with 100% probability.
	524288 attempts suffice for 50% probability of finding a match.
Overall number of attempts across all pairs is 1820 for 100% probability and 910 for 50% probability

...

[+] Address with prefix "63" and suffix "8c" found.
Address: 0x634064125a07d3ef655d024a2a50d01ad911af8c
PrivateKey: fafa91a26410e86e15e5965891f771fb0f9a0fae7fead669eebf318b62a0927e
```