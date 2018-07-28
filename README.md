# Sample usage

Following command will try to find an Ethereum address with prefix/suffix pairs as (57, 5ab), (63, 8c), and (69, 457). Number of threads executed in parallel would be 32. The code will stop at the first match. 
`
go run similar-ethereum-address-finder.go -p 57,63,69 -s 5Ab,8c,457 -t 32
`