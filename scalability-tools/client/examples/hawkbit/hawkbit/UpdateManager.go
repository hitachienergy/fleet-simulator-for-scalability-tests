package hawkbit

type UpdateStage int

const (
	SUCCESSFUL UpdateStage = iota
	ERROR
	RUNNING
	DOWNLOADING
	DOWNLOADED
	CONFIRMED
	CANCEL
)

type LocalUpdateStatus struct {
	Status     UpdateStage
	StatusMsgs []string
}

type UpdateManagerDDI interface {
	PrepareUpdate(actionID int64) (deployment *Deployment, err error)
	StartUpdate(actionID int64, deployment *Deployment) (err error)
}

type UpdateManagerDMF interface {
	StartUpdate(action *DMFAction, requireInstall bool) (err error)
	ResetUpdate(actionID int64) bool
}
