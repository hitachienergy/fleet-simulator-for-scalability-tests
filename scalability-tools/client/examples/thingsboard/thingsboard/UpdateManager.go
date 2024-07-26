package thingsboard

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"strings"
	"sync"

	"github.com/spaolacci/murmur3"
	"golang.org/x/xerrors"
)

type CommunicationService interface {
	GetFirmware(fwinfo FirmwareInfo) (data []byte, err error)
	ReportUpdateState(state FWUpdateState) (err error)
	GetID() string
}

type UpdateModule interface {
	StartUpdate(fw FirmwareInfo) (err error)
	ReportCurrState() error
	CheckFw(fw FirmwareInfo) bool
}

type UpdateManager struct {
	UpdateModule
	*HTTPClient
	*sync.RWMutex

	currFW     FWUpdateState
	isUpdating bool
}

// newUpdateManager creates an instance of updatemanager
func newUpdateManager(c *HTTPClient) *UpdateManager {
	return &UpdateManager{
		HTTPClient: c,
		RWMutex:    &sync.RWMutex{},
	}
}

// CheckFw implements UpdateModule.CheckFw
func (u *UpdateManager) CheckFw(fw FirmwareInfo) bool {
	u.Lock()
	defer u.Unlock()

	if len(fw.Title) == 0 || len(fw.Version) == 0 {
		return false
	}

	if u.currFW.Title == fw.Title && u.currFW.Version == fw.Version {
		return false
	}
	if u.isUpdating {
		return false
	}

	u.isUpdating = true
	return true
}

// ReportCurrState implements UpdateModule.ReportCurrState
func (u *UpdateManager) ReportCurrState() error {
	u.RLock()
	state := u.currFW
	u.RUnlock()

	return u.reportUpdateState(state)
}

// StartUpdate implements UpdateModule.StartUpdate
func (u *UpdateManager) StartUpdate(fw FirmwareInfo) (err error) {
	var updateFw FWUpdateState

	u.HTTPClient.controller.GetLogger().Debug().Msg("Start downloading")

	updateFw = FWUpdateState{
		Title:   fw.Title,
		Version: fw.Version,
		State:   UPDATE_DOWNLOADING,
	}
	u.reportUpdateState(updateFw)

	data, err := u.getFirmware(fw)
	if err != nil {
		updateFw.State = UPDATE_FAILED
		u.reportUpdateState(updateFw)
		return err
	}
	updateFw.State = UPDATE_DOWNLOADED
	u.HTTPClient.controller.GetLogger().Debug().Msg("Finished downloading")
	u.reportUpdateState(updateFw)

	u.HTTPClient.controller.GetLogger().Debug().Msg("Verify checksum")

	err = u.verifyChecksum(fw, data)
	if err != nil {
		updateFw.State = UPDATE_FAILED
		u.reportUpdateState(updateFw)
		return err
	}
	updateFw.State = UPDATE_VERIFIED
	u.reportUpdateState(updateFw)

	u.HTTPClient.controller.GetLogger().Debug().Msg("Update")

	updateFw.State = UPDATE_UPDATING
	u.reportUpdateState(updateFw)

	u.HTTPClient.controller.GetLogger().Debug().Msg("Updated")

	updateFw.State = UPDATE_UPDATED
	u.reportUpdateState(updateFw)

	u.Lock()
	if err == nil {
		u.currFW = updateFw
	}
	u.isUpdating = false
	u.Unlock()

	u.controller.CompleteTask(err == nil)

	return nil
}

// verifyChecksum verifies the integrity of firmware based on the firmware info
func (u *UpdateManager) verifyChecksum(fw FirmwareInfo, data []byte) error {
	if len(data) == 0 {
		return xerrors.Errorf("Empty Firmware data")
	}
	if len(fw.Checksum) == 0 {
		// if len(fw.ChecksumAlg) == 0 {
		u.controller.GetLogger().Warn().Msg("No checksum provided")
		return nil
		// }
		// return xerrors.Errorf("No checksum provided")
	}

	var hashString string
	switch fw.ChecksumAlg {
	case SHA256:
		myhash := sha256.New()
		myhash.Write(data)
		hashString = strings.ToLower(hex.EncodeToString(myhash.Sum(nil)))
	case SHA348:
		myhash := sha512.New384()
		myhash.Write(data)
		hashString = strings.ToLower(hex.EncodeToString(myhash.Sum(nil)))
	case SHA512:
		myhash := sha512.New()
		myhash.Write(data)
		hashString = strings.ToLower(hex.EncodeToString(myhash.Sum(nil)))
	case MD5:
		myhash := md5.New()
		myhash.Write(data)
		hashString = hex.EncodeToString(myhash.Sum(nil))
	case MURMUR3_32:
		// TODO: check whether this output matches python mmh3
		myhash := murmur3.New32()
		myhash.Write(data)
		hashString = hex.EncodeToString(myhash.Sum(nil))
	case MURMUR3_128:
		myhash := murmur3.New128()
		myhash.Write(data)
		hashString = hex.EncodeToString(myhash.Sum(nil))
	case CRC32:
		checksum := crc32.ChecksumIEEE(data)
		hashString = fmt.Sprintf("%08X", checksum)
	default:
		return xerrors.Errorf("Unrecognized checksum algorithm: %s", fw.ChecksumAlg)
	}
	if !strings.EqualFold(hashString, fw.Checksum) {
		return xerrors.Errorf("Checksume unmatched. Checksum Algo: %s", fw.ChecksumAlg)
	}
	return nil
}
