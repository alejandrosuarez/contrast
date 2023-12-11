package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/google/go-sev-guest/abi"
)

type Report struct {
	Version         uint32 `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"` // Should be 1 for revision 1.51
	GuestSvn        uint32 `protobuf:"varint,2,opt,name=guest_svn,json=guestSvn,proto3" json:"guest_svn,omitempty"`
	Policy          uint64 `protobuf:"varint,3,opt,name=policy,proto3" json:"policy,omitempty"`
	FamilyId        []byte `protobuf:"bytes,4,opt,name=family_id,json=familyId,proto3" json:"family_id,omitempty"` // Should be 16 bytes long
	ImageId         []byte `protobuf:"bytes,5,opt,name=image_id,json=imageId,proto3" json:"image_id,omitempty"`    // Should be 16 bytes long
	Vmpl            uint32 `protobuf:"varint,6,opt,name=vmpl,proto3" json:"vmpl,omitempty"`
	SignatureAlgo   uint32 `protobuf:"varint,7,opt,name=signature_algo,json=signatureAlgo,proto3" json:"signature_algo,omitempty"`
	CurrentTcb      uint64 `protobuf:"varint,8,opt,name=current_tcb,json=currentTcb,proto3" json:"current_tcb,omitempty"`
	PlatformInfo    uint64 `protobuf:"varint,9,opt,name=platform_info,json=platformInfo,proto3" json:"platform_info,omitempty"`
	SignerInfo      uint32 `protobuf:"varint,10,opt,name=signer_info,json=signerInfo,proto3" json:"signer_info,omitempty"`                 // AuthorKeyEn, MaskChipKey, SigningKey
	ReportData      []byte `protobuf:"bytes,11,opt,name=report_data,json=reportData,proto3" json:"report_data,omitempty"`                  // Should be 64 bytes long
	Measurement     []byte `protobuf:"bytes,12,opt,name=measurement,proto3" json:"measurement,omitempty"`                                  // Should be 48 bytes long
	HostData        []byte `protobuf:"bytes,13,opt,name=host_data,json=hostData,proto3" json:"host_data,omitempty"`                        // Should be 32 bytes long
	IdKeyDigest     []byte `protobuf:"bytes,14,opt,name=id_key_digest,json=idKeyDigest,proto3" json:"id_key_digest,omitempty"`             // Should be 48 bytes long
	AuthorKeyDigest []byte `protobuf:"bytes,15,opt,name=author_key_digest,json=authorKeyDigest,proto3" json:"author_key_digest,omitempty"` // Should be 48 bytes long
	ReportId        []byte `protobuf:"bytes,16,opt,name=report_id,json=reportId,proto3" json:"report_id,omitempty"`                        // Should be 32 bytes long
	ReportIdMa      []byte `protobuf:"bytes,17,opt,name=report_id_ma,json=reportIdMa,proto3" json:"report_id_ma,omitempty"`                // Should be 32 bytes long
	ReportedTcb     uint64 `protobuf:"varint,18,opt,name=reported_tcb,json=reportedTcb,proto3" json:"reported_tcb,omitempty"`
	ChipId          []byte `protobuf:"bytes,19,opt,name=chip_id,json=chipId,proto3" json:"chip_id,omitempty"` // Should be 64 bytes long
	CommittedTcb    uint64 `protobuf:"varint,20,opt,name=committed_tcb,json=committedTcb,proto3" json:"committed_tcb,omitempty"`
	// Each build, minor, major triple should be packed together in a uint32
	// packed together at 7:0, 15:8, 23:16 respectively
	CurrentBuild   uint32 `protobuf:"varint,21,opt,name=current_build,json=currentBuild,proto3" json:"current_build,omitempty"`
	CurrentMinor   uint32 `protobuf:"varint,22,opt,name=current_minor,json=currentMinor,proto3" json:"current_minor,omitempty"`
	CurrentMajor   uint32 `protobuf:"varint,23,opt,name=current_major,json=currentMajor,proto3" json:"current_major,omitempty"`
	CommittedBuild uint32 `protobuf:"varint,24,opt,name=committed_build,json=committedBuild,proto3" json:"committed_build,omitempty"`
	CommittedMinor uint32 `protobuf:"varint,25,opt,name=committed_minor,json=committedMinor,proto3" json:"committed_minor,omitempty"`
	CommittedMajor uint32 `protobuf:"varint,26,opt,name=committed_major,json=committedMajor,proto3" json:"committed_major,omitempty"`
	LaunchTcb      uint64 `protobuf:"varint,27,opt,name=launch_tcb,json=launchTcb,proto3" json:"launch_tcb,omitempty"`
	Signature      []byte `protobuf:"bytes,28,opt,name=signature,proto3" json:"signature,omitempty"` // Should be 512 bytes long
}

func main() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("failed to read input: %v", err)
	}
	data, err = hex.DecodeString(string(data))
	if err != nil {
		log.Fatalf("failed to decode input: %v", err)
	}

	var r Report
	// r.Version should be 2, but that's left to validation step.
	r.Version = binary.LittleEndian.Uint32(data[0x00:0x04])
	fmt.Printf("Version: %x\n", r.Version)
	r.GuestSvn = binary.LittleEndian.Uint32(data[0x04:0x08])
	fmt.Printf("GuestSvn: %x\n", r.GuestSvn)
	r.Policy = binary.LittleEndian.Uint64(data[0x08:0x10])
	fmt.Printf("Policy: %x\n", r.Policy)
	if _, err := abi.ParseSnpPolicy(r.Policy); err != nil {
		fmt.Printf("malformed guest policy: %v", err)
	}
	r.FamilyId = clone(data[0x10:0x20])
	fmt.Printf("FamilyId: %x\n", r.FamilyId)
	r.ImageId = clone(data[0x20:0x30])
	fmt.Printf("ImageId: %x\n", r.ImageId)
	r.Vmpl = binary.LittleEndian.Uint32(data[0x30:0x34])
	fmt.Printf("Vmpl: %d\n", r.Vmpl)
	// r.SignatureAlgo = SignatureAlgo(data)
	r.CurrentTcb = binary.LittleEndian.Uint64(data[0x38:0x40])
	fmt.Printf("CurrentTcb: %d\n", r.CurrentTcb)
	r.PlatformInfo = binary.LittleEndian.Uint64(data[0x40:0x48])
	fmt.Printf("PlatformInfo: %d\n", r.PlatformInfo)

	signerInfo, err := abi.ParseSignerInfo(binary.LittleEndian.Uint32(data[0x48:0x4C]))
	if err != nil {
		log.Printf("failed to parse signer info: %v", err)
	}
	r.SignerInfo = abi.ComposeSignerInfo(signerInfo)
	if err := mbz(data, 0x4C, 0x50); err != nil {
		log.Printf("%v", err)
	}
	fmt.Printf("SignerInfo: %v\n", r.SignerInfo)
	r.ReportData = clone(data[0x50:0x90])
	fmt.Printf("ReportData: %x\n", r.ReportData)
	r.Measurement = clone(data[0x90:0xC0])
	fmt.Printf("Measurement: %x\n", r.Measurement)
	r.HostData = clone(data[0xC0:0xE0])
	fmt.Printf("HostData: %x\n", r.HostData)
	r.IdKeyDigest = clone(data[0xE0:0x110])
	fmt.Printf("IdKeyDigest: %x\n", r.IdKeyDigest)
	r.AuthorKeyDigest = clone(data[0x110:0x140])
	fmt.Printf("AuthorKeyDigest: %x\n", r.AuthorKeyDigest)
	r.ReportId = clone(data[0x140:0x160])
	fmt.Printf("ReportId: %x\n", r.ReportId)
	r.ReportIdMa = clone(data[0x160:0x180])
	fmt.Printf("ReportIdMa: %x\n", r.ReportIdMa)
	r.ReportedTcb = binary.LittleEndian.Uint64(data[0x180:0x188])
	if err := mbz(data, 0x188, 0x1A0); err != nil {
		fmt.Printf("%v", err)
	}

	r.ChipId = clone(data[0x1A0:0x1E0])
	fmt.Printf("ChipId: %x\n", r.ChipId)
	r.CommittedTcb = binary.LittleEndian.Uint64(data[0x1E0:0x1E8])
	fmt.Printf("CommittedTcb: %d\n", r.CommittedTcb)
	r.CurrentBuild = uint32(data[0x1E8])
	fmt.Printf("CurrentBuild: %d\n", r.CurrentBuild)
	r.CurrentMinor = uint32(data[0x1E9])
	fmt.Printf("CurrentMinor: %d\n", r.CurrentMinor)
	r.CurrentMajor = uint32(data[0x1EA])
	fmt.Printf("CurrentMajor: %d\n", r.CurrentMajor)
	if err := mbz(data, 0x1EB, 0x1EC); err != nil {
		fmt.Printf("%v", err)
	}
	r.CommittedBuild = uint32(data[0x1EC])
	fmt.Printf("CommittedBuild: %d\n", r.CommittedBuild)
	r.CommittedMinor = uint32(data[0x1ED])
	fmt.Printf("CommittedMinor: %d\n", r.CommittedMinor)
	r.CommittedMajor = uint32(data[0x1EE])
	fmt.Printf("CommittedMajor: %d\n", r.CommittedMajor)
	if err := mbz(data, 0x1EF, 0x1F0); err != nil {
		fmt.Printf("%v", err)
	}
	r.LaunchTcb = binary.LittleEndian.Uint64(data[0x1F0:0x1F8])
	if err := mbz(data, 0x1F8, signatureOffset); err != nil {
		fmt.Printf("%v", err)
	}

	if r.SignatureAlgo == abi.SignEcdsaP384Sha384 {
		if err := mbz(data, signatureOffset+abi.EcdsaP384Sha384SignatureSize, abi.ReportSize); err != nil {
			fmt.Printf("%v", err)
		}
	}
	r.Signature = clone(data[signatureOffset:abi.ReportSize])
	fmt.Printf("Signature: %x\n", r.Signature)
}

func clone(b []byte) []byte {
	result := make([]byte, len(b))
	copy(result, b)
	return result
}

func mbz(data []uint8, lo, hi int) error {
	if findNonZero(data, lo, hi) != hi {
		return fmt.Errorf("mbz range [0x%x:0x%x] not all zero: %s", lo, hi, hex.EncodeToString(data[lo:hi]))
	}
	return nil
}

// findNonZero returns the first index which is not zero, otherwise the length of the slice.
func findNonZero(data []uint8, lo, hi int) int {
	for i := lo; i < hi; i++ {
		if data[i] != 0 {
			return i
		}
	}
	return hi
}

const signatureOffset = 0x2A0