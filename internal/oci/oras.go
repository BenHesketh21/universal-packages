package oci

import (
	"context"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
)

type OrasClient interface {
	Copy(ctx context.Context, src oras.ReadOnlyTarget, srcRef string, dst oras.Target, dstRef string, options oras.CopyOptions) (v1.Descriptor, error)
}

type OrasClientImpl struct{}

func (c *OrasClientImpl) Copy(ctx context.Context, src oras.ReadOnlyTarget, srcRef string, dst oras.Target, dstRef string, options oras.CopyOptions) (v1.Descriptor, error) {
	return oras.Copy(ctx, src, srcRef, dst, dstRef, options)
}
