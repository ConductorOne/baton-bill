package local

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/conductorone/baton-sdk/pkg/dotc1z"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type localManager struct {
	filePath string
	tmpPath  string
}

func (l *localManager) copyFileToTmp(ctx context.Context) error {
	tmp, err := os.CreateTemp("", "sync-*.c1z")
	if err != nil {
		return err
	}
	defer tmp.Close()

	l.tmpPath = tmp.Name()

	if l.filePath == "" {
		return nil
	}

	if _, err = os.Stat(l.filePath); err == nil {
		f, err := os.Open(l.filePath)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(tmp, f)
		if err != nil {
			return err
		}
	}

	return nil
}

// LoadRaw returns an io.Reader of the bytes in the c1z file.
func (l *localManager) LoadRaw(ctx context.Context) (io.Reader, error) {
	err := l.copyFileToTmp(ctx)
	if err != nil {
		return nil, err
	}

	fBytes, err := os.ReadFile(l.tmpPath)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(fBytes), nil
}

// LoadC1Z loads the C1Z file from the local file system.
func (l *localManager) LoadC1Z(ctx context.Context) (*dotc1z.C1File, error) {
	log := ctxzap.Extract(ctx)

	err := l.copyFileToTmp(ctx)
	if err != nil {
		return nil, err
	}

	log.Debug(
		"successfully loaded c1z locally",
		zap.String("file_path", l.filePath),
		zap.String("temp_path", l.tmpPath),
	)

	return dotc1z.NewC1ZFile(ctx, l.tmpPath)
}

// SaveC1Z saves the C1Z file to the local file system.
func (l *localManager) SaveC1Z(ctx context.Context) error {
	log := ctxzap.Extract(ctx)

	if l.tmpPath == "" {
		return fmt.Errorf("unexpected state - missing temp file path")
	}

	if l.filePath == "" {
		return fmt.Errorf("unexpected state - missing file path")
	}

	tmpFile, err := os.Open(l.tmpPath)
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	dstFile, err := os.Create(l.filePath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	size, err := io.Copy(dstFile, tmpFile)
	if err != nil {
		return err
	}

	log.Debug(
		"successfully saved c1z locally",
		zap.String("file_path", l.filePath),
		zap.String("temp_path", l.tmpPath),
		zap.Int64("bytes", size),
	)

	return nil
}

func (l *localManager) Close(ctx context.Context) error {
	err := os.Remove(l.tmpPath)
	if err != nil {
		return err
	}
	return nil
}

// New returns a new localManager that uses the given filePath.
func New(ctx context.Context, filePath string) (*localManager, error) {
	return &localManager{
		filePath: filePath,
	}, nil
}
