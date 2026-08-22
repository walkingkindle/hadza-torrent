package download

import "sync"

type DownloadStatus struct {
	Mu         sync.Mutex
	Done       []bool
	InProgress []bool
}
