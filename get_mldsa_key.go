package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
        "os"
        "strconv"

	"github.com/cloudflare/circl/sign/mldsa/mldsa44"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run get_mldsa_key.go <validator_id>")
        os.Exit(1)
    }
    
    n, err := strconv.Atoi(os.Args[1])
    if err != nil {
        fmt.Println("Invalid validator_id:", err)
        os.Exit(1)
    }

	var seed [32]byte
	binary.LittleEndian.PutUint32(seed[:4], uint32(n))

	_, priv := mldsa44.NewKeyFromSeed(&seed)
    
    privBytes := priv.Bytes()

	fmt.Printf("Validator ID: %d\n", n)
	fmt.Printf("Private Key (hex): %s\n", hex.EncodeToString(privBytes))
}
