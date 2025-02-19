package util

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

func Download(ctx context.Context, from, to string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, from, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request url %q: %w", from, err)
	}
	defer resp.Body.Close()

	file, err := os.Create(to)
	if err != nil {
		return fmt.Errorf("failed to create %q file: %w", to, err)
	}

	var downloaded bool
	defer func() {
		file.Close()
		if !downloaded {
			RemoveFile(to)
		}
	}()

	defer file.Close()

	if _, err := io.Copy(file, NewReaderCtx(ctx, resp.Body)); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	downloaded = true
	return nil
}

type ContextReader struct {
	ctx context.Context
	r   io.Reader
}

func NewReaderCtx(ctx context.Context, r io.Reader) io.Reader {
	return &ContextReader{
		ctx: ctx,
		r:   r,
	}
}

func (r *ContextReader) Read(p []byte) (n int, err error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
		return r.r.Read(p)
	}
}
