// Package job defines the job types used by the worker system
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

type Job interface {
	Execute(ctx context.Context) error
	ID() string
	Status() Status
	Error() error
	Requeue() Job
}

type BaseJob struct {
	id string

	mu     sync.Mutex
	status Status
	err    error
}

func (b *BaseJob) ID() string {
	return b.id
}

func (b *BaseJob) Status() Status {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.status
}

func (b *BaseJob) setStatus(s Status) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.status = s
}

func (b *BaseJob) Error() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.err
}

func (b *BaseJob) setError(e error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.err = e
}

type DownloadInputs struct {
	URL             string
	OutputDirectory string
	Quality         string
}

type DownloadJob struct {
	BaseJob
	Inputs     DownloadInputs
	downloader download.Downloader
}

func NewDownloadJob(inputs DownloadInputs, downloader download.Downloader) *DownloadJob {
	return &DownloadJob{
		BaseJob:    BaseJob{id: uuid.New().String(), status: StatusPending},
		Inputs:     inputs,
		downloader: downloader,
	}
}

func (d *DownloadJob) Execute(ctx context.Context) error {
	d.setStatus(StatusProcessing)
	err := d.downloader.Download(ctx, d.Inputs.URL, d.Inputs.OutputDirectory, d.Inputs.Quality)
	if err != nil {
		d.setStatus(StatusFailed)
		d.setError(err)
		return err
	}
	d.setStatus(StatusDone)
	return nil
}

func (d *DownloadJob) Requeue() Job {
	return NewDownloadJob(d.Inputs, d.downloader)
}

type EncodeInputs struct {
	InputPath       string
	OutputDirectory string
	Format          string
}

func (e EncodeInputs) outputPath() string {
	fileName := filepath.Base(e.InputPath)
	fileNameNoExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	return filepath.Join(e.OutputDirectory, fileNameNoExt+"."+e.Format)
}

type EncodeJob struct {
	BaseJob
	Inputs  EncodeInputs
	encoder encode.Encoder
}

func NewEncodeJob(inputs EncodeInputs, encoder encode.Encoder) *EncodeJob {
	return &EncodeJob{
		BaseJob: BaseJob{id: uuid.New().String(), status: StatusPending},
		Inputs:  inputs,
		encoder: encoder,
	}
}

func (e *EncodeJob) Execute(ctx context.Context) error {
	e.setStatus(StatusProcessing)

	err := e.encoder.Encode(ctx, e.Inputs.InputPath, e.Inputs.outputPath())
	if err != nil {
		e.setStatus(StatusFailed)
		e.setError(err)

		return err
	}
	e.setStatus(StatusDone)
	return nil
}

func (e *EncodeJob) Requeue() Job {
	return NewEncodeJob(e.Inputs, e.encoder)
}

type ProcessInputs struct {
	URL             string
	OutputDirectory string
	Format          string
	Quality         string
}

type ProcessJob struct {
	BaseJob
	Inputs     ProcessInputs
	downloader download.Downloader
	encoder    encode.Encoder
}

func NewProcessJob(inputs ProcessInputs, downloader download.Downloader, encoder encode.Encoder) *ProcessJob {
	return &ProcessJob{
		BaseJob:    BaseJob{id: uuid.New().String(), status: StatusPending},
		Inputs:     inputs,
		downloader: downloader,
		encoder:    encoder,
	}
}

func (p *ProcessJob) Execute(ctx context.Context) error {
	p.setStatus(StatusProcessing)

	tempPrefix := filepath.Join(os.TempDir(), "fornax-"+p.id)
	downloadTemplate := tempPrefix + "-%(title)s.%(ext)s"

	if err := p.downloader.Download(ctx, p.Inputs.URL, downloadTemplate, p.Inputs.Quality); err != nil {
		p.setStatus(StatusFailed)
		p.setError(err)

		return err
	}

	files, err := filepath.Glob(tempPrefix + "-*.*")
	if err != nil {
		p.setStatus(StatusFailed)
		p.setError(err)

		return err
	}
	if len(files) == 0 {
		err := errors.New("could not find downloaded file in temp directory")
		p.setStatus(StatusFailed)
		p.setError(err)
		return err
	}

	downloadedFile := files[0]
	defer os.Remove(downloadedFile)

	fileName := filepath.Base(downloadedFile)
	fileExt := filepath.Ext(fileName)
	fileNameNoExt := strings.TrimSuffix(fileName, fileExt)
	videoTitle := strings.TrimPrefix(fileNameNoExt, "fornax-"+p.id+"-")
	outputPath := filepath.Join(p.Inputs.OutputDirectory, videoTitle+"."+p.Inputs.Format)

	if err := p.encoder.Encode(ctx, downloadedFile, outputPath); err != nil {
		p.setStatus(StatusFailed)
		p.setError(err)

		return err
	}

	p.setStatus(StatusDone)
	return nil
}

func (p *ProcessJob) Requeue() Job {
	return NewProcessJob(p.Inputs, p.downloader, p.encoder)
}
