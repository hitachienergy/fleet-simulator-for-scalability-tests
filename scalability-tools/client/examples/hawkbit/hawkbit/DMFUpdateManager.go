package hawkbit

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/xerrors"
)

/** Update Manager **/

type DMFUpdateManager struct {
	actions map[int64]context.CancelFunc
	*sync.Mutex

	*DMFClient
}

// newDMFUpdateManager creates a new DMFUpdateManager
func newDMFUpdateManager(client *DMFClient) *DMFUpdateManager {
	return &DMFUpdateManager{
		actions:   map[int64]context.CancelFunc{},
		Mutex:     &sync.Mutex{},
		DMFClient: client,
	}
}

// setUpdate records an update in the table
func (u *DMFUpdateManager) setUpdate(actionID int64) (ctx context.Context) {
	u.Lock()
	defer u.Unlock()

	if _, ok := u.actions[actionID]; ok {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	u.actions[actionID] = cancel
	return ctx
}

// resetUpdate removes an update from the table
func (u *DMFUpdateManager) resetUpdate(actionID int64) (exist bool) {
	u.Lock()
	defer u.Unlock()

	cancel, ok := u.actions[actionID]
	if ok {
		cancel()
		delete(u.actions, actionID)
	}

	return ok
}

// func (u *DMFUpdateManager) PrepareUpdate(actionID int64) (deployment *Deployment, err error) { return }

// startUpdate launches the updates
func (u *DMFUpdateManager) startUpdate(action *DMFAction, requireInstall bool) (err error) {
	ctx := u.setUpdate(action.ID)
	if ctx == nil {
		return nil
	}

	defer func() {
		// FIXME: only set actionID for SUCCESS or ERROR
		u.resetUpdate(action.ID)
		u.controller.CompleteTask(err == nil)
	}()

	err = u.reportUpdate(action.ID, LocalUpdateStatus{RUNNING, []string{"Simulation begins!"}})
	if err != nil {
		return err
	}

	cancel, err := u.simulateDownload(ctx, action)
	if err != nil {
		return err
	}

	if !cancel && requireInstall {
		err = u.reportUpdate(action.ID, LocalUpdateStatus{SUCCESSFUL, []string{"Simulation complete!"}})
		return err
	}

	return nil
}

// simulateDownload simulates the downlaod process
func (u *DMFUpdateManager) simulateDownload(ctx context.Context, action *DMFAction) (finish bool, err error) {
	messages := []string{}
	for _, module := range action.SoftwareModules {
		for _, artifact := range module.Artifacts {
			messages = append(messages,
				fmt.Sprintf("Download starts for: %s with SHA1 hash %s and size %d",
					artifact.Filename, artifact.Hashes["sha1"], artifact.Size))
		}
	}

	err = u.reportUpdate(action.ID, LocalUpdateStatus{DOWNLOADING, messages})
	if err != nil {
		return false, err
	}

	u.controller.GetLogger().Debug().Msg("Start downloading")

	result := LocalUpdateStatus{Status: DOWNLOADED}
	for _, chunk := range action.SoftwareModules {
		for _, artifact := range chunk.Artifacts {
			select {
			case <-ctx.Done():
				u.controller.GetLogger().Debug().Msg("Cancel downloading")
				return true, nil
			default:
				status := u.handleArtifact(ctx, &artifact, action.TargetSecurityToken)
				result.StatusMsgs = append(result.StatusMsgs, status.StatusMsgs...)
				if status.Status == ERROR {
					err = xerrors.Errorf(result.StatusMsgs[0])
					result.Status = ERROR
				}
			}
		}
	}
	reportErr := u.reportUpdate(action.ID, LocalUpdateStatus{result.Status, result.StatusMsgs})

	u.controller.GetLogger().Debug().Msgf("Finish downloading. Success: %t", result.Status == DOWNLOADED)

	if err != nil {
		return false, err
	}

	return false, reportErr
}

// handleArtifact verifies the artifact
func (u *DMFUpdateManager) handleArtifact(ctx context.Context, artifact *DMFArtifact, token string) (status LocalUpdateStatus) {
	if url, ok := artifact.Urls["HTTPS"]; ok {
		return u.downloadUrl(ctx, url, token, artifact.Hashes["sha1"], artifact.Size)
	} else if url, ok := artifact.Urls["HTTP"]; ok {
		return u.downloadUrl(ctx, url, token, artifact.Hashes["sha1"], artifact.Size)
	} else {
		return LocalUpdateStatus{Status: ERROR,
			StatusMsgs: []string{fmt.Sprintf("No supported url for artifact %s (Expected: HTTPS, HTTP)", artifact.Filename)}}
	}
}

// downloadUrl performs downloading from the given url using different protocols
func (u *DMFUpdateManager) downloadUrl(ctx context.Context, url string, token string, hash string, size int64) (status LocalUpdateStatus) {
	data, err := u.download(ctx, url, token)
	if err != nil {
		return LocalUpdateStatus{Status: ERROR, StatusMsgs: []string{fmt.Sprintf("Failed to download %s: %s", url, err)}}
	}

	contentSize := int64(len(data))
	if contentSize != size {
		return LocalUpdateStatus{Status: ERROR,
			StatusMsgs: []string{fmt.Sprintf("Download %s has wrong content length (Expected: %d, Got %d)", url, size, contentSize)}}
	}

	myhash := sha1.New()
	myhash.Write(data)
	hashString := strings.ToLower(hex.EncodeToString(myhash.Sum(nil)))
	if hashString != hash {
		return LocalUpdateStatus{Status: ERROR,
			StatusMsgs: []string{fmt.Sprintf("Download %s failed with SHA1 hash missmatch (Expected: %s, Got %s)", url, hash, hashString)}}
	}

	return LocalUpdateStatus{Status: SUCCESSFUL, StatusMsgs: []string{fmt.Sprintf("Download %s successfully (%d bytes)", url, contentSize)}}
}
