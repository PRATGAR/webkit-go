package repo

import (
	"github.com/go-git/go-git/v5"
)

// Mirror clones a remote source repository into a local working directory so
// the integrations pipeline can index its contents. It backs the planned
// source-control integration and is wired up by the worker, not the API.
type Mirror struct {
	workdir string
}

func NewMirror(workdir string) *Mirror {
	return &Mirror{workdir: workdir}
}

func (m *Mirror) Fetch(remoteURL string) (*git.Repository, error) {
	return git.PlainClone(m.workdir, false, &git.CloneOptions{
		URL: remoteURL,
	})
}

// MirrorRepository fetches a remote repository into workdir.
func MirrorRepository(remoteURL, workdir string) error {
	_, err := NewMirror(workdir).Fetch(remoteURL)
	return err
}
