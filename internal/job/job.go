// Package job contains job struct
package job

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/kevinle-00/fornax/internal/download"
	"github.com/kevinle-00/fornax/internal/encode"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusDone       Status = "done"
	StatusFailed     Status = "failed"
)

type DownloadInputs struct {
	URL        string
	OutputPath string
	Quality    string
}

type EncodeInputs struct {
	InputPath  string
	OutputPath string
}
type ProcessInputs struct {
	URL        string
	OutputPath string
}

type Job interface {
	Execute(ctx context.Context) error
	GetID() string
	GetStatus() Status
}

type BaseJob struct {
	ID     string
	Status Status
	Error  error
	mu     sync.Mutex
}

func (b *BaseJob) GetID() string {
	return b.ID
}

func (b *BaseJob) GetStatus() Status {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Status
}

func (b *BaseJob) SetStatus(s Status) {
	b.mu.Lock()
	b.Status = s
	b.mu.Unlock()
}

func (b *BaseJob) GetError() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Error
}

func (b *BaseJob) SetError(e error) {
	b.mu.Lock()
	b.Error = e
	b.mu.Unlock()
}

type DownloadJob struct {
	BaseJob
	Inputs     DownloadInputs
	downloader download.Downloader
}

func (d *DownloadJob) Execute(ctx context.Context) error {
	d.SetStatus(StatusProcessing)
	err := d.downloader.Download(ctx, d.Inputs.URL, d.Inputs.OutputPath, d.Inputs.Quality)
	if err != nil {
		d.SetStatus(StatusFailed)
		d.SetError(err)
		return err
	}
	d.SetStatus(StatusDone)
	return nil
}

func NewDownloadJob(inputs DownloadInputs, downloader download.Downloader) *DownloadJob {
	return &DownloadJob{
		BaseJob:    BaseJob{ID: uuid.New().String(), Status: StatusPending},
		Inputs:     inputs,
		downloader: downloader,
	}
}

type EncodeJob struct {
	BaseJob
	Inputs  EncodeInputs
	encoder encode.Encoder
}

func (e *EncodeJob) Execute(ctx context.Context) error {
	e.SetStatus(StatusProcessing)
	err := e.encoder.Encode(ctx, e.Inputs.InputPath, e.Inputs.OutputPath)
	if err != nil {
		e.SetStatus(StatusFailed)
		e.SetError(err)

		return err
	}
	e.SetStatus(StatusDone)
	return nil
}

func NewEncodeJob(inputs EncodeInputs, encoder encode.Encoder) *EncodeJob {
	return &EncodeJob{
		BaseJob: BaseJob{ID: uuid.New().String(), Status: StatusPending},
		Inputs:  inputs,
		encoder: encoder,
	}
}

type ProcessJob struct {
	BaseJob
	Inputs     ProcessInputs
	downloader download.Downloader
	encoder    encode.Encoder
}

func (p *ProcessJob) Execute(ctx context.Context) error {
	p.SetStatus(StatusProcessing)

	tempBase := "/tmp/fornax-temp-" + p.ID
	tempPath := tempBase + ".%(ext)s"

	if err := p.downloader.Download(ctx, p.Inputs.URL, tempPath, ""); err != nil {
		p.SetStatus(StatusFailed)
		p.SetError(err)

		return p.GetError()
	}

	files, err := filepath.Glob(tempBase + ".*")
	if err != nil {
		p.SetStatus(StatusFailed)
		p.SetError(err)

		return err
	}
	if len(files) == 0 {
		p.SetStatus(StatusFailed)
		p.SetError(errors.New("could not find downloaded file in /tmp"))
		return p.GetError()
	}

	tempFile := files[0]
	defer os.Remove(tempFile)

	if err := p.encoder.Encode(ctx, tempFile, p.Inputs.OutputPath); err != nil {
		p.SetStatus(StatusFailed)
		p.SetError(err)

		return p.GetError()
	}

	p.SetStatus(StatusDone)
	return nil
}

func NewProcessJob(inputs ProcessInputs, downloader download.Downloader, encoder encode.Encoder) *ProcessJob {
	return &ProcessJob{
		BaseJob:    BaseJob{ID: uuid.New().String(), Status: StatusPending},
		Inputs:     inputs,
		downloader: downloader,
		encoder:    encoder,
	}
}
