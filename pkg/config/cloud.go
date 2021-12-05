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

package config

import (
	"net/http"
	"os"

	"github.com/huaweicloud/golangsdk"
	"github.com/huaweicloud/golangsdk/openstack"
)

// CloudCredentials define
type CloudCredentials struct {
	Global struct {
		AccessKey   string `gcfg:"access-key"`
		SecretKey   string `gcfg:"secret-key"`
		Region      string `gcfg:"region"`
		ProjectName string `gcfg:"project-name"`
		ProjectId   string `gcfg:"project-id"`
		Password    string `gcfg:"password"`
		Username    string `gcfg:"username"`
		AuthURL     string `gcfg:"auth-url"`
		DomainID    string `gcfg:"domain-id"`
		DomainName  string `gcfg:"domain-name"`
		TenantID    string `gcfg:"tenant-id"`
	}

	Vpc struct {
		Id string `gcfg:"id"`
	}

	CloudClient *golangsdk.ProviderClient
}

// Validate CloudCredentials
func (c *CloudCredentials) Validate() error {
	err := c.newCloudClient()
	if err != nil {
		return err
	}
	return nil
}

// newCloudClient returns new cloud client
func (c *CloudCredentials) newCloudClient() error {
	if c.Global.AccessKey != "" {
		ao := golangsdk.AKSKAuthOptions{
			IdentityEndpoint: c.Global.AuthURL,
			AccessKey:        c.Global.AccessKey,
			SecretKey:        c.Global.SecretKey,
			Region:           c.Global.Region,
			ProjectName:      c.Global.ProjectName,
			ProjectId:        c.Global.ProjectId,
		}
		client, err := openstack.NewClient(ao.IdentityEndpoint)
		if err != nil {
			return err
		}
		var osDebug bool
		if os.Getenv("OS_DEBUG") != "" {
			osDebug = true
		}

		transport := &http.Transport{Proxy: http.ProxyFromEnvironment}
		client.HTTPClient = http.Client{
			Transport: &LogRoundTripper{
				Rt:      transport,
				OsDebug: osDebug,
			},
		}

		err = openstack.Authenticate(client, ao)
		if err != nil {
			return err
		}

		c.CloudClient = client
	} else if c.Global.Password != "" {
		pao := golangsdk.AuthOptions{
			IdentityEndpoint: c.Global.AuthURL,
			DomainID:         c.Global.DomainID,
			DomainName:       c.Global.DomainName,
			Username:         c.Global.Username,
			Password:         c.Global.Password,
			TenantID:         c.Global.TenantID,
		}
		client, err := openstack.NewClient(pao.IdentityEndpoint)
		if err != nil {
			return err
		}
		var osDebug bool
		if os.Getenv("OS_DEBUG") != "" {
			osDebug = true
		}

		transport := &http.Transport{Proxy: http.ProxyFromEnvironment}
		client.HTTPClient = http.Client{
			Transport: &LogRoundTripper{
				Rt:      transport,
				OsDebug: osDebug,
			},
		}

		err = openstack.Authenticate(client, pao)
		if err != nil {
			return err
		}

		c.CloudClient = client
	}

	return nil
}

// SFSV2Client return sfs v2 client
func (c *CloudCredentials) SFSV2Client() (*golangsdk.ServiceClient, error) {
	return openstack.NewHwSFSV2(c.CloudClient, golangsdk.EndpointOpts{
		Region:       c.Global.Region,
		Availability: golangsdk.AvailabilityPublic,
	})
}
