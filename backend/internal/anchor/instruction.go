package anchor

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/gagliardetto/solana-go"
)

// DiscriminatorFromName computes the 8-byte Anchor instruction discriminator.
// discriminator = sha256("global:<name>")[0:8]
func DiscriminatorFromName(name string) [8]byte {
	h := sha256.Sum256([]byte("global:" + name))
	var d [8]byte
	copy(d[:], h[:8])
	return d
}

// BuildSubmitMilestoneProofInstruction constructs the submit_milestone_proof instruction
// data for the oracle to send on-chain.
func BuildSubmitMilestoneProofInstruction(
	programID solana.PublicKey,
	oracle solana.PublicKey,
	projectState solana.PublicKey,
	milestoneRecord solana.PublicKey,
	instructionsSysvar solana.PublicKey,
	milestoneIndex uint8,
	proofURI [128]byte,
	proofHash [32]byte,
	oracleSignature [64]byte,
) *GenericInstruction {
	disc := DiscriminatorFromName("submit_milestone_proof")

	// Encode args: milestone_index (u8) + proof_uri ([u8;128]) + proof_hash ([u8;32]) + oracle_signature ([u8;64])
	data := make([]byte, 0, 8+1+128+32+64)
	data = append(data, disc[:]...)
	data = append(data, milestoneIndex)
	data = append(data, proofURI[:]...)
	data = append(data, proofHash[:]...)
	data = append(data, oracleSignature[:]...)

	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(oracle, true, true),
		solana.NewAccountMeta(projectState, true, false),
		solana.NewAccountMeta(milestoneRecord, true, false),
		solana.NewAccountMeta(instructionsSysvar, false, false),
		solana.NewAccountMeta(projectState, false, false), // project (has_one check)
	}

	return &GenericInstruction{
		programID: programID,
		accounts:  accounts,
		data:      data,
	}
}

// BuildReleaseMilestoneFundsInstruction builds the release_milestone_funds instruction.
func BuildReleaseMilestoneFundsInstruction(
	programID solana.PublicKey,
	oracle solana.PublicKey,
	projectState solana.PublicKey,
	milestoneRecord solana.PublicKey,
	escrowVault solana.PublicKey,
	developerUSDCAccount solana.PublicKey,
	usdcTokenProgram solana.PublicKey,
	milestoneIndex uint8,
) *GenericInstruction {
	disc := DiscriminatorFromName("release_milestone_funds")

	data := make([]byte, 0, 8+1)
	data = append(data, disc[:]...)
	data = append(data, milestoneIndex)

	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(oracle, true, true),
		solana.NewAccountMeta(projectState, true, false),
		solana.NewAccountMeta(milestoneRecord, true, false),
		solana.NewAccountMeta(escrowVault, true, false),
		solana.NewAccountMeta(developerUSDCAccount, true, false),
		solana.NewAccountMeta(usdcTokenProgram, false, false),
	}

	return &GenericInstruction{
		programID: programID,
		accounts:  accounts,
		data:      data,
	}
}

// EncodeU64LE encodes a uint64 as little-endian bytes.
func EncodeU64LE(v uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v)
	return b
}

// EncodeU32LE encodes a uint32 as little-endian bytes.
func EncodeU32LE(v uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return b
}

// GenericInstruction implements solana.Instruction for custom instruction data.
type GenericInstruction struct {
	programID solana.PublicKey
	accounts  solana.AccountMetaSlice
	data      []byte
}

func (g *GenericInstruction) ProgramID() solana.PublicKey     { return g.programID }
func (g *GenericInstruction) Accounts() solana.AccountMetaSlice { return g.accounts }
func (g *GenericInstruction) Data() ([]byte, error)           { return g.data, nil }
