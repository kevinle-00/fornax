// Package job contains job struct
package job

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
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
	URL             string
	OutputDirectory string
	Quality         string
}

type EncodeInputs struct {
	InputPath       string
	OutputDirectory string
	Format          string
}
type ProcessInputs struct {
	URL             string
	OutputDirectory string
	Format          string
	Quality         string
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
	err := d.downloader.Download(ctx, d.Inputs.URL, d.Inputs.OutputDirectory, d.Inputs.Quality)
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
	fileName := filepath.Base(e.Inputs.InputPath)
	fileNameNoExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	outputPath := filepath.Join(e.Inputs.OutputDirectory, fileNameNoExt+"."+e.Inputs.Format)

	err := e.encoder.Encode(ctx, e.Inputs.InputPath, outputPath)
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

	tempPrefix := "/tmp/fornax-" + p.ID
	downloadTemplate := tempPrefix + "-%(title)s.%(ext)s"

	if err := p.downloader.Download(ctx, p.Inputs.URL, downloadTemplate, p.Inputs.Quality); err != nil {
		p.SetStatus(StatusFailed)
		p.SetError(err)

		return p.GetError()
	}

	files, err := filepath.Glob(tempPrefix + "-*.*")
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

	downloadedFile := files[0]
	defer os.Remove(downloadedFile)

	fileName := filepath.Base(downloadedFile)
	fileExt := filepath.Ext(fileName)
	fileNameNoExt := strings.TrimSuffix(fileName, fileExt)
	videoTitle := strings.TrimPrefix(fileNameNoExt, "fornax-"+p.ID+"-")
	outputPath := filepath.Join(p.Inputs.OutputDirectory, videoTitle+"."+p.Inputs.Format)

	if err := p.encoder.Encode(ctx, downloadedFile, outputPath); err != nil {
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
