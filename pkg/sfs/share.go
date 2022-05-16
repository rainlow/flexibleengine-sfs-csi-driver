/*
Copyright 2020 The Kubernetes Authors.

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

package sfs

import (
	"time"

	"github.com/huaweicloud/golangsdk"
	"github.com/huaweicloud/golangsdk/openstack/sfs/v2/shares"
	"k8s.io/klog"
)

const (
	waitForAvailableShareTimeout = 300

	shareAvailable = "available"

	shareDescription = "provisioned-by=sfs.csi.huaweicloud.org"
)

func getShareExists(client *golangsdk.ServiceClient, name string) (*shares.Share, error) {
	klog.V(2).Infof("List shares with name: %v", name)
	listOpts := shares.ListOpts{
		Name:    name,
		SortKey: shares.SortId,
		SortDir: shares.SortAsc,
	}
	results, err := shares.List(client, listOpts)
	if err != nil {
		klog.V(2).Infof("%v List shares failed: %v", name, err)
		return nil, err
	}

	if len(results) > 1 {
		klog.V(2).Infof("%v There is more then one shares found: %v", name, results)
	}

	if len(results) == 0 {
		return nil, nil
	} else {
		return &results[0], nil
	}
}

func createShare(client *golangsdk.ServiceClient, createOpts *shares.CreateOpts) (*shares.Share, error) {
	klog.V(2).Infof("Begin to list and create share %v", createOpts.Name)
	startTime := time.Now().UnixMilli()

	shareExists, err := getShareExists(client, createOpts.Name)
	if err != nil {
		klog.V(2).Infof("%v List shares failed, create a new share. %v", createOpts.Name, err)
	}
	klog.V(2).Infof("%v Exist share: %+v, list time used: %v ms",
		createOpts.Name, shareExists, time.Now().UnixMilli()-startTime)

	if shareExists == nil {
		klog.V(2).Infof("%v Begin to create a new share", createOpts.Name)
		createOpts.Description = shareDescription
		shareExists, err = shares.Create(client, createOpts).Extract()
		if err != nil {
			klog.V(2).Infof("%v share create err: %v, time used: %v ms",
				createOpts.Name, err, time.Now().UnixMilli()-startTime)
			return nil, err
		}
	}

	klog.V(2).Infof("Begin to wait share available %v, time used: %v ms",
		createOpts.Name, time.Now().UnixMilli()-startTime)
	err = waitForShareStatus(client, shareExists.ID, shareAvailable, waitForAvailableShareTimeout)
	if err != nil {
		klog.V(2).Infof("waitForShareStatus %v(ID: %v) err: %v, time waited: %v",
			createOpts.Name, shareExists.ID, err, time.Now().UnixMilli()-startTime)
		return nil, err
	}
	klog.V(2).Infof("Create share %v(ID: %v) time used: %v ms",
		createOpts.Name, shareExists.ID, time.Now().UnixMilli()-startTime)
	return shareExists, nil
}

func deleteShare(client *golangsdk.ServiceClient, shareID string) error {
	if err := shares.Delete(client, shareID).ExtractErr(); err != nil {
		if _, ok := err.(golangsdk.ErrDefault404); ok {
			klog.V(4).Infof("share %s not found, assuming it to be already deleted", shareID)
		} else {
			return err
		}
	}

	return nil
}

// waitForShareStatus wait for share desired status until timeout
func waitForShareStatus(client *golangsdk.ServiceClient, shareID string, desiredStatus string, timeout int) error {
	return golangsdk.WaitFor(timeout, func() (bool, error) {
		share, err := getShare(client, shareID)
		if err != nil {
			return false, err
		}
		return share.Status == desiredStatus, nil
	})
}

func getShare(client *golangsdk.ServiceClient, shareID string) (*shares.Share, error) {
	return shares.Get(client, shareID).Extract()
}

func isAccessGranted(client *golangsdk.ServiceClient, shareID string, vpcid string) bool {
	ruleList, err := shares.ListAccessRights(client, shareID).ExtractAccessRights()
	if err != nil {
		klog.V(2).Infof("%v Query access rule failed. %v", shareID, err)
		return false
	}

	for _, rule := range ruleList {
		if rule.AccessTo == vpcid {
			klog.V(2).Infof("%v access rule already granted for vpc: %v", shareID, vpcid)
			return true
		}
	}

	klog.V(2).Infof("%v access rule not found for vpc: %v", shareID, vpcid)
	return false
}

func grantAccess(client *golangsdk.ServiceClient, shareID string, vpcid string) error {
	if isAccessGranted(client, shareID, vpcid) {
		return nil
	}

	klog.V(2).Infof("%v begin to grant access for vpc: %v", shareID, vpcid)
	// build GrantAccessOpts
	grantAccessOpts := shares.GrantAccessOpts{}
	grantAccessOpts.AccessLevel = "rw"
	grantAccessOpts.AccessType = "cert"
	grantAccessOpts.AccessTo = vpcid

	// grant access
	_, err := shares.GrantAccess(client, shareID, grantAccessOpts).ExtractAccess()
	if err != nil {
		return err
	}
	return nil
}

func expandShare(client *golangsdk.ServiceClient, shareID string, size int) error {
	expandOpts := shares.ExpandOpts{OSExtend: shares.OSExtendOpts{NewSize: size}}
	expand := shares.Expand(client, shareID, expandOpts)
	if expand.Err != nil {
		return expand.Err
	}
	return nil
}
