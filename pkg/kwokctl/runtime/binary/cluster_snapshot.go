/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package binary

import (
	"context"
	"os"

	"sigs.k8s.io/kwok/pkg/kwokctl/runtime"
	"sigs.k8s.io/kwok/pkg/kwokctl/utils"
	"sigs.k8s.io/kwok/pkg/kwokctl/vars"
)

// SnapshotSave save the snapshot of cluster
func (c *Cluster) SnapshotSave(ctx context.Context, path string) error {
	conf, err := c.Config()
	if err != nil {
		return err
	}

	bin := utils.PathJoin(conf.Workdir, "bin")
	etcdctlPath := utils.PathJoin(bin, "etcdctl"+vars.BinSuffix)

	err = utils.DownloadWithCacheAndExtract(ctx, conf.CacheDir, conf.EtcdBinaryTar, etcdctlPath, "etcdctl"+vars.BinSuffix, 0755, conf.QuietPull, true)
	if err != nil {
		return err
	}

	err = utils.Exec(ctx, "", utils.IOStreams{}, etcdctlPath, "snapshot", "save", path, "--endpoints=127.0.0.1:"+utils.StringUint32(conf.EtcdPort))
	if err != nil {
		return err
	}

	return nil
}

// SnapshotRestore restore the snapshot of cluster
func (c *Cluster) SnapshotRestore(ctx context.Context, path string) error {
	conf, err := c.Config()
	if err != nil {
		return err
	}

	bin := utils.PathJoin(conf.Workdir, "bin")
	etcdctlPath := utils.PathJoin(bin, "etcdctl"+vars.BinSuffix)

	err = utils.DownloadWithCacheAndExtract(ctx, conf.CacheDir, conf.EtcdBinaryTar, etcdctlPath, "etcdctl"+vars.BinSuffix, 0755, conf.QuietPull, true)
	if err != nil {
		return err
	}

	err = c.Stop(ctx, "etcd")
	if err != nil {
		c.Logger().Error("Failed to stop etcd", err)
	}
	defer func() {
		err = c.Start(ctx, "etcd")
		if err != nil {
			c.Logger().Error("Failed to start etcd", err)
		}
	}()

	etcdDataTmp := utils.PathJoin(conf.Workdir, "etcd-data")
	os.RemoveAll(etcdDataTmp)
	err = utils.Exec(ctx, "", utils.IOStreams{}, etcdctlPath, "snapshot", "restore", path, "--data-dir", etcdDataTmp)
	if err != nil {
		return err
	}

	etcdDataPath := utils.PathJoin(conf.Workdir, runtime.EtcdDataDirName)
	os.RemoveAll(etcdDataPath)
	err = os.Rename(etcdDataTmp, etcdDataPath)
	if err != nil {
		return err
	}
	return nil
}
