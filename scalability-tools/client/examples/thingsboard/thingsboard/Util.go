package thingsboard

import "strings"

var FIRMWARE_SHARED_KEYS = strings.Join([]string{
	"fw_checksum_algorithm",
	"fw_checksum",
	"fw_size",
	"fw_title",
	"fw_version"}, ",")

type ChecksumAlg string

const (
	SHA256      ChecksumAlg = "SHA256"
	SHA348      ChecksumAlg = "SHA348"
	SHA512      ChecksumAlg = "SHA512"
	MD5         ChecksumAlg = "MD5"
	MURMUR3_32  ChecksumAlg = "MURMUR3_32"
	MURMUR3_128 ChecksumAlg = "MURMUR3_128"
	CRC32       ChecksumAlg = "CRC32"
)

type FirmwareInfo struct {
	Checksum    string      `json:"fw_checksum"`
	ChecksumAlg ChecksumAlg `json:"fw_checksum_algorithm"`
	Size        float64     `json:"fw_size"`
	Title       string      `json:"fw_title"`
	Version     string      `json:"fw_version"`
}

type UpdateState string

const (
	UPDATE_DOWNLOADING UpdateState = "DOWNLOADING"
	UPDATE_DOWNLOADED  UpdateState = "DOWNLOADED"
	UPDATE_VERIFIED    UpdateState = "VERIFIED"
	UPDATE_UPDATING    UpdateState = "UPDATING"
	UPDATE_UPDATED     UpdateState = "UPDATED"
	UPDATE_FAILED      UpdateState = "FAILED"
)

type FWUpdateState struct {
	Title   string      `json:"current_fw_title"`
	Version string      `json:"current_fw_version"`
	State   UpdateState `json:"fw_state"`
}

// ---------------------- HTTP ----------------------

type HTTPAttributes struct {
	Client map[string]interface{} `json:"client"`
	Shared FirmwareInfo           `json:"shared"`
}
