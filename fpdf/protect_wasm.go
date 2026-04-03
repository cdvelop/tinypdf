//go:build wasm

package fpdf

// Advisory bitflag constants that control document activities
const (
	CnProtectPrint      = 4
	CnProtectModify     = 8
	CnProtectCopy       = 16
	CnProtectAnnotForms = 32
)

type protectType struct {
	encrypted     bool
	uValue        []byte
	oValue        []byte
	pValue        int
	padding       []byte
	encryptionKey []byte
	objNum        int
}

func (p *protectType) rc4(n uint32, buf *[]byte) {
}

func (p *protectType) objectKey(n uint32) []byte {
	return nil
}

func oValueGen(userPass, ownerPass []byte) (v []byte) {
	return nil
}

func (p *protectType) uValueGen() (v []byte) {
	return nil
}

func (p *protectType) setProtection(privFlag byte, userPassStr, ownerPassStr string) {
}
