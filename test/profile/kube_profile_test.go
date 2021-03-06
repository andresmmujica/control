package profile

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/supergiant/control/pkg/profile"
	"github.com/supergiant/control/pkg/sgerrors"
	"github.com/supergiant/control/pkg/storage/file"
)

func TestProfileGet(t *testing.T) {
	s, err := file.NewFileRepository(fmt.Sprintf("/tmp/sg-storage-%d", time.Now().UnixNano()))
	require.Nil(t, err, "setup file storage provider")

	testCases := []struct {
		expectedId string
		data       []byte
		err        error
	}{
		{
			expectedId: "1234",
			data:       []byte(`{"id":"1234", "nodes":[{},{}]}`),
			err:        nil,
		},
		{
			data: nil,
			err:  sgerrors.ErrNotFound,
		},
	}

	prefix := "/kube/"

	for _, testCase := range testCases {
		if len(testCase.expectedId) > 0 {
			s.Put(context.Background(), prefix, testCase.expectedId, testCase.data)
		}

		service := profile.NewService(prefix, s)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		p, err := service.Get(ctx, testCase.expectedId)

		if testCase.err != err {
			t.Errorf("Wrong error expected %v actual %v", testCase.err, err)
			return
		}

		if testCase.err == nil && p.ID != testCase.expectedId {
			t.Errorf("Wrong profile id expected %s actual %s", testCase.expectedId, p.ID)
		}
	}
}

func TestKubeProfileCreate(t *testing.T) {
	s, err := file.NewFileRepository(fmt.Sprintf("/tmp/sg-storage-%d", time.Now().UnixNano()))
	require.Nil(t, err, "setup file storage provider")

	prefix := "/profile/"
	key := "key"
	version := "1.8.7"

	kube := &profile.Profile{
		ID:         key,
		K8SVersion: version,
	}

	service := profile.NewService(prefix, s)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = service.Create(ctx, kube)

	if err != nil {
		t.Errorf("Unepexpected error while creating kube profile %v", err)
	}

	kube2, err := service.Get(ctx, kube.ID)

	if err != nil {
		t.Errorf("Unexpected error while getting kube profile %v", err)
	}

	if kube.ID != key || kube.K8SVersion != kube2.K8SVersion {
		t.Errorf("Wrong data in etcd")
	}
}

func TestKubeProfileGetAll(t *testing.T) {
	s, err := file.NewFileRepository(fmt.Sprintf("/tmp/sg-storage-%d", time.Now().UnixNano()))
	require.Nil(t, err, "setup file storage provider")

	prefix := "/profile/"
	key := "key"
	version := "1.8.7"

	service := profile.NewService(prefix, s)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	kube := &profile.Profile{
		ID:         key,
		K8SVersion: version,
	}

	err = service.Create(ctx, kube)

	if err != nil {
		t.Errorf("Unepexpected error while creating kube profile %v", err)
	}

	kubeProfiles, err := service.GetAll(ctx)

	if err != nil {
		t.Errorf("Unexpected error getting kube profiles %v", err)
	}

	if len(kubeProfiles) == 0 {
		t.Error("Kube profiles are empty")
	}
}
