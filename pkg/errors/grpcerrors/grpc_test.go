// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpcerrors

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func TestGRPC(t *testing.T) {
	a := assertions.New(t)
	d := &errors.ErrDescriptor{
		MessageFormat: "You do not have access to app with id {app_id}",
		Code:          77,
		Type:          errors.PermissionDenied,
		Namespace:     "pkg/foo",
		SafeAttributes: []string{
			"app_id",
			"count",
		},
	}
	d.Register()

	attributes := errors.Attributes{
		"app_id": "foo",
		"count":  42,
		"unsafe": "secret",
	}

	err := d.New(attributes)

	code := GRPCCode(err)
	a.So(code, should.Equal, codes.PermissionDenied)

	// other errors should be unknown
	other := fmt.Errorf("Foo")
	code = GRPCCode(other)
	a.So(code, should.Equal, codes.Unknown)

	grpcErr := ToGRPC(err)

	got := FromGRPC(grpcErr)
	a.So(got.Code(), should.Equal, d.Code)
	a.So(got.Type(), should.Equal, d.Type)
	a.So(got.Message(), should.Equal, "You do not have access to app with id foo")
	a.So(got.Error(), should.Equal, "pkg/foo[77]: You do not have access to app with id foo")
	a.So(got.ID(), should.Equal, err.ID())

	a.So(got.Attributes(), should.NotBeEmpty)
	a.So(got.Attributes()["app_id"], should.Resemble, attributes["app_id"])
	a.So(got.Attributes()["count"], should.AlmostEqual, attributes["count"])
	a.So(got.Attributes(), should.NotContainKey, "unsafe")
}

func TestFromUnspecifiedGRPC(t *testing.T) {
	a := assertions.New(t)

	err := grpc.Errorf(codes.DeadlineExceeded, "This is an error")

	got := FromGRPC(err)
	a.So(got.Code(), should.Equal, errors.NoCode)
	a.So(got.Type(), should.Equal, errors.Timeout)
	a.So(got.Error(), should.Equal, "This is an error")
	a.So(got.Attributes(), should.BeNil)
	a.So(got.ID(), should.NotBeEmpty)
}

func TestWellKnown(t *testing.T) {
	a := assertions.New(t)

	a.So(errors.ErrEOF.Describes(FromGRPC(ToGRPC(io.EOF))), should.BeTrue)
	a.So(errors.ErrContextCanceled.Describes(FromGRPC(ToGRPC(context.Canceled))), should.BeTrue)
	a.So(errors.ErrContextDeadlineExceeded.Describes(FromGRPC(ToGRPC(context.DeadlineExceeded))), should.BeTrue)
	a.So(ErrClientConnClosing.Describes(FromGRPC(ToGRPC(grpc.ErrClientConnClosing))), should.BeTrue)
	a.So(ErrClientConnTimeout.Describes(FromGRPC(ToGRPC(grpc.ErrClientConnTimeout))), should.BeTrue)
	a.So(ErrServerStopped.Describes(FromGRPC(ToGRPC(grpc.ErrServerStopped))), should.BeTrue)

	a.So(errors.ErrEOF.Describes(FromGRPC(io.EOF)), should.BeTrue)
	a.So(errors.ErrContextCanceled.Describes(FromGRPC(context.Canceled)), should.BeTrue)
	a.So(errors.ErrContextDeadlineExceeded.Describes(FromGRPC(context.DeadlineExceeded)), should.BeTrue)
	a.So(ErrClientConnClosing.Describes(FromGRPC(grpc.ErrClientConnClosing)), should.BeTrue)
	a.So(ErrClientConnTimeout.Describes(FromGRPC(grpc.ErrClientConnTimeout)), should.BeTrue)
	a.So(ErrServerStopped.Describes(FromGRPC(grpc.ErrServerStopped)), should.BeTrue)
}
