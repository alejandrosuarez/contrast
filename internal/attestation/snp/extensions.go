package snp

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"math/big"

	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/kds"
	"github.com/google/go-sev-guest/proto/sevsnp"
	"golang.org/x/exp/constraints"
)

var (
	rootOID = asn1.ObjectIdentifier{1, 3, 9901, 2, 1}

	versionOID  = append(rootOID, 1)
	guestSVNOID = append(rootOID, 2)

	policyABIMajorOID     = append(rootOID, 3)
	policyABIMinorOID     = append(rootOID, 4)
	policyDebugOID        = append(rootOID, 5)
	policyMigrateMAOID    = append(rootOID, 6)
	policySMTOID          = append(rootOID, 7)
	policySingleSocketOID = append(rootOID, 8)

	familyIDOID = append(rootOID, 9)
	imageIDOID  = append(rootOID, 10)
	vmplOID     = append(rootOID, 11)

	currentTCBPartsBlSplOID    = append(rootOID, 12)
	currentTCBPartsSnpSplOID   = append(rootOID, 13)
	currentTCBPartsTeeSplOID   = append(rootOID, 14)
	currentTCBPartsUcodeSplOID = append(rootOID, 15)

	platformInfoSMTEnabledOID  = append(rootOID, 16)
	platformInfoTSMEEnabledOID = append(rootOID, 17)

	singerInfoAuthorKeyEnOID = append(rootOID, 18)
	singerInfoMaskChipKeyOID = append(rootOID, 19)
	singerInfoSigningKeyOID  = append(rootOID, 20)

	reportDataOID      = append(rootOID, 21)
	measurementOID     = append(rootOID, 22)
	hostDataOID        = append(rootOID, 23)
	idKeyDigestOID     = append(rootOID, 24)
	authorKeyDigestOID = append(rootOID, 25)
	reportIDOID        = append(rootOID, 26)
	reportIDMAOID      = append(rootOID, 27)

	reportedTCBPartsBlSplOID    = append(rootOID, 28)
	reportedTCBPartsSnpSplOID   = append(rootOID, 29)
	reportedTCBPartsTeeSplOID   = append(rootOID, 30)
	reportedTCBPartsUcodeSplOID = append(rootOID, 31)

	chipIDOID = append(rootOID, 32)

	committedTCBPartsBlSplOID    = append(rootOID, 32)
	committedTCBPartsSnpSplOID   = append(rootOID, 33)
	committedTCBPartsTeeSplOID   = append(rootOID, 34)
	committedTCBPartsUcodeSplOID = append(rootOID, 35)

	currentBuildOID   = append(rootOID, 36)
	currentMinorOID   = append(rootOID, 37)
	currentMajorOID   = append(rootOID, 38)
	committedBuildOID = append(rootOID, 39)
	committedMinorOID = append(rootOID, 40)
	committedMajorOID = append(rootOID, 41)

	launchTCBPartsBlSplOID    = append(rootOID, 42)
	launchTCBPartsSnpSplOID   = append(rootOID, 43)
	launchTCBPartsTeeSplOID   = append(rootOID, 44)
	launchTCBPartsUcodeSplOID = append(rootOID, 45)
)

type bigIntExtension struct {
	OID asn1.ObjectIdentifier
	Val *big.Int
}

func (b bigIntExtension) toExtension() (pkix.Extension, error) {
	bytes, err := asn1.Marshal(b.Val)
	if err != nil {
		return pkix.Extension{}, fmt.Errorf("marshaling big int: %w", err)
	}
	return pkix.Extension{Id: b.OID, Value: bytes}, nil
}

type bytesExtension struct {
	OID asn1.ObjectIdentifier
	Val []byte
}

func (b bytesExtension) toExtension() (pkix.Extension, error) {
	bytes, err := asn1.Marshal(b.Val)
	if err != nil {
		return pkix.Extension{}, fmt.Errorf("marshaling bytes: %w", err)
	}
	return pkix.Extension{Id: b.OID, Value: bytes}, nil
}

type boolExtension struct {
	OID asn1.ObjectIdentifier
	Val bool
}

func (b boolExtension) toExtension() (pkix.Extension, error) {
	bytes, err := asn1.Marshal(b.Val)
	if err != nil {
		return pkix.Extension{}, fmt.Errorf("marshaling bool: %w", err)
	}
	return pkix.Extension{Id: b.OID, Value: bytes}, nil
}

type extension interface {
	toExtension() (pkix.Extension, error)
}

// ClaimsToCertExtension constructs certificate extensions from a SNP report.
func ClaimsToCertExtension(report *sevsnp.Report) ([]pkix.Extension, error) {
	var extensions []extension
	extensions = append(extensions, bigIntExtension{OID: versionOID, Val: usingedToBigInt(report.Version)})
	extensions = append(extensions, bigIntExtension{OID: guestSVNOID, Val: usingedToBigInt(report.GuestSvn)})

	parsedPolicy, err := abi.ParseSnpPolicy(report.Policy)
	if err != nil {
		return nil, fmt.Errorf("parsing policy: %w", err)
	}

	extensions = append(extensions, bigIntExtension{OID: policyABIMajorOID, Val: usingedToBigInt(parsedPolicy.ABIMajor)})
	extensions = append(extensions, bigIntExtension{OID: policyABIMinorOID, Val: usingedToBigInt(parsedPolicy.ABIMinor)})
	extensions = append(extensions, boolExtension{OID: policySMTOID, Val: parsedPolicy.SMT})
	extensions = append(extensions, boolExtension{OID: policyMigrateMAOID, Val: parsedPolicy.MigrateMA})
	extensions = append(extensions, boolExtension{OID: policyDebugOID, Val: parsedPolicy.Debug})
	extensions = append(extensions, boolExtension{OID: policySingleSocketOID, Val: parsedPolicy.SingleSocket})

	extensions = append(extensions, bytesExtension{OID: familyIDOID, Val: report.FamilyId})
	extensions = append(extensions, bytesExtension{OID: imageIDOID, Val: report.ImageId})
	extensions = append(extensions, bigIntExtension{OID: vmplOID, Val: usingedToBigInt(report.Vmpl)})

	tcbParts := kds.DecomposeTCBVersion(kds.TCBVersion(report.CurrentTcb))
	extensions = append(extensions, bigIntExtension{OID: currentTCBPartsBlSplOID, Val: usingedToBigInt(tcbParts.BlSpl)})
	extensions = append(extensions, bigIntExtension{OID: currentTCBPartsTeeSplOID, Val: usingedToBigInt(tcbParts.TeeSpl)})
	extensions = append(extensions, bigIntExtension{OID: currentTCBPartsSnpSplOID, Val: usingedToBigInt(tcbParts.SnpSpl)})
	extensions = append(extensions, bigIntExtension{OID: currentTCBPartsUcodeSplOID, Val: usingedToBigInt(tcbParts.UcodeSpl)})

	parsedPlatformInfo, err := abi.ParseSnpPlatformInfo(report.PlatformInfo)
	if err != nil {
		return nil, fmt.Errorf("parsing platform info: %w", err)
	}
	extensions = append(extensions, boolExtension{OID: platformInfoSMTEnabledOID, Val: parsedPlatformInfo.SMTEnabled})
	extensions = append(extensions, boolExtension{OID: platformInfoTSMEEnabledOID, Val: parsedPlatformInfo.TSMEEnabled})

	parsedSingerInfo, err := abi.ParseSignerInfo(report.SignerInfo)
	if err != nil {
		return nil, fmt.Errorf("parsing singer info: %w", err)
	}
	extensions = append(extensions, bigIntExtension{OID: singerInfoSigningKeyOID, Val: usingedToBigInt(parsedSingerInfo.SigningKey)})
	extensions = append(extensions, boolExtension{OID: singerInfoAuthorKeyEnOID, Val: parsedSingerInfo.AuthorKeyEn})
	extensions = append(extensions, boolExtension{OID: singerInfoMaskChipKeyOID, Val: parsedSingerInfo.MaskChipKey})

	extensions = append(extensions, bytesExtension{OID: reportDataOID, Val: report.ReportData})
	extensions = append(extensions, bytesExtension{OID: measurementOID, Val: report.Measurement})
	extensions = append(extensions, bytesExtension{OID: hostDataOID, Val: report.HostData})
	extensions = append(extensions, bytesExtension{OID: idKeyDigestOID, Val: report.IdKeyDigest})
	extensions = append(extensions, bytesExtension{OID: authorKeyDigestOID, Val: report.AuthorKeyDigest})
	extensions = append(extensions, bytesExtension{OID: reportIDOID, Val: report.ReportId})
	extensions = append(extensions, bytesExtension{OID: reportIDMAOID, Val: report.ReportIdMa})

	tcbParts = kds.DecomposeTCBVersion(kds.TCBVersion(report.ReportedTcb))
	extensions = append(extensions, bigIntExtension{OID: reportedTCBPartsBlSplOID, Val: usingedToBigInt(tcbParts.BlSpl)})
	extensions = append(extensions, bigIntExtension{OID: reportedTCBPartsTeeSplOID, Val: usingedToBigInt(tcbParts.TeeSpl)})
	extensions = append(extensions, bigIntExtension{OID: reportedTCBPartsSnpSplOID, Val: usingedToBigInt(tcbParts.SnpSpl)})
	extensions = append(extensions, bigIntExtension{OID: reportedTCBPartsUcodeSplOID, Val: usingedToBigInt(tcbParts.UcodeSpl)})

	extensions = append(extensions, bytesExtension{OID: chipIDOID, Val: report.ChipId})

	tcbParts = kds.DecomposeTCBVersion(kds.TCBVersion(report.CommittedTcb))
	extensions = append(extensions, bigIntExtension{OID: committedTCBPartsBlSplOID, Val: usingedToBigInt(tcbParts.BlSpl)})
	extensions = append(extensions, bigIntExtension{OID: committedTCBPartsTeeSplOID, Val: usingedToBigInt(tcbParts.TeeSpl)})
	extensions = append(extensions, bigIntExtension{OID: committedTCBPartsSnpSplOID, Val: usingedToBigInt(tcbParts.SnpSpl)})
	extensions = append(extensions, bigIntExtension{OID: committedTCBPartsUcodeSplOID, Val: usingedToBigInt(tcbParts.UcodeSpl)})

	extensions = append(extensions, bigIntExtension{OID: currentBuildOID, Val: usingedToBigInt(report.CurrentBuild)})
	extensions = append(extensions, bigIntExtension{OID: currentMinorOID, Val: usingedToBigInt(report.CurrentMinor)})
	extensions = append(extensions, bigIntExtension{OID: currentMajorOID, Val: usingedToBigInt(report.CurrentMajor)})
	extensions = append(extensions, bigIntExtension{OID: committedBuildOID, Val: usingedToBigInt(report.CommittedBuild)})
	extensions = append(extensions, bigIntExtension{OID: committedMinorOID, Val: usingedToBigInt(report.CommittedMinor)})
	extensions = append(extensions, bigIntExtension{OID: committedMajorOID, Val: usingedToBigInt(report.CommittedMajor)})

	tcbParts = kds.DecomposeTCBVersion(kds.TCBVersion(report.LaunchTcb))
	extensions = append(extensions, bigIntExtension{OID: launchTCBPartsBlSplOID, Val: usingedToBigInt(tcbParts.BlSpl)})
	extensions = append(extensions, bigIntExtension{OID: launchTCBPartsTeeSplOID, Val: usingedToBigInt(tcbParts.TeeSpl)})
	extensions = append(extensions, bigIntExtension{OID: launchTCBPartsSnpSplOID, Val: usingedToBigInt(tcbParts.SnpSpl)})
	extensions = append(extensions, bigIntExtension{OID: launchTCBPartsUcodeSplOID, Val: usingedToBigInt(tcbParts.UcodeSpl)})

	var exts []pkix.Extension
	for _, extension := range extensions {
		ext, err := extension.toExtension()
		if err != nil {
			return nil, fmt.Errorf("converting extension to pkix: %w", err)
		}
		exts = append(exts, ext)
	}

	return exts, nil
}

func usingedToBigInt[T constraints.Unsigned](value T) *big.Int {
	bigInt := &big.Int{}
	bigInt.SetUint64(uint64(value))
	return bigInt
}
