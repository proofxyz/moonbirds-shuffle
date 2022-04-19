// The shuffler binary shuffles Moonbirds metadata traits using a hash as a
// seed. The 32 bytes of the hash are split into 4 8-byte words that are
// interpreted as uint64s and XORd together. The folded value is used as a seed
// for a rand.Rand that is used to shuffle the metadata.
package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	_ "embed"
)

//go:embed unshuffled.json
var unshuffled []byte

func main() {
	seedHex := flag.String("seed_hash", "", "Hex seed for shuffling")
	flag.Parse()

	// Establish the provenance of the full, unshuffled collection.
	log.Printf("Initial collection: %#x", crypto.Keccak256(unshuffled))
	var metadata []map[string]interface{}
	if err := json.Unmarshal(unshuffled, &metadata); err != nil {
		log.Fatalf("json.Unmarshal([all metadata]): %v", err)
	}
	if n := len(metadata); n != 10_000 {
		log.Fatalf("Unmarshalled %d metadata objects; expecting 10k", n)
	}

	// Convert the seed hex into a seed for a random-number generator used to
	// perform shuffling.
	rng := seededRNG(common.HexToHash(*seedHex))
	rng.Shuffle(len(metadata), func(i, j int) {
		metadata[i], metadata[j] = metadata[j], metadata[i]
	})

	// Establish provenance of the shuffled collection.
	buf, err := json.Marshal(metadata)
	if err != nil {
		log.Fatalf("json.Marshal([shuffled metadata]): %v", err)
	}
	log.Printf("Shuffled collection: %#x", crypto.Keccak256(buf))

	if err := os.WriteFile("shuffled.json", buf, 0644); err != nil {
		log.Fatalf("os.WriteFile([shuffled JSON]): %v", err)
	}
}

// seededRNG folds the seed hash with XOR to obtain a single word of 64 bits,
// which is used as the seed of the returned RNG.
func seededRNG(seedHash common.Hash) *rand.Rand {
	seedBytes := seedHash.Bytes()
	log.Printf("Seed hash: %#x", seedBytes)

	// Fold the four 8-byte words of the seed, using xor, into a seed for the
	// RNG.
	var uSeed uint64
	for i := 0; i < 32; i += 8 {
		word := seedBytes[i : i+8]
		uSeed ^= binary.BigEndian.Uint64(word)
	}
	log.Printf("XOR-folded bytes: %#x", uSeed)
	seed := int64(uSeed)
	log.Printf("RNG seed: %d", seed)

	return rand.New(rand.NewSource(seed))
}
