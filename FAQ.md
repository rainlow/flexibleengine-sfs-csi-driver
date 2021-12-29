# FAQ : Frequently Asked Questions on SFS CSI Driver for Kubernetes
 Container Storage Interface (CSI) Plugin makes it possible to use [SFS](https://docs.prod-cloud-ocb.orange-business.com/en-us/sfs/index.html) with your self-built Kubernetes cluster on Flexible Engine. This list can help you facing known errors when installing and using SFS CSI Driver.
___

### **Q:** How can i create access-key & secret-key ?
- AK & SK creation take place in your flexible engine console using your [credentials](https://docs.prod-cloud-ocb.orange-business.com/api/cce/en-us_topic_0035951710.html)

### **Q:** No nodes are available when deploying pod ? / Infinite pending pod ?
They might be multiple reasons:
- One common error is that you did not set the [cloud-config](https://github.com/huaweicloud/huaweicloud-csi-driver/blob/master/deploy/cloud-config) file on **all nodes** in `/etc/sfs/` repository.
- Check if your vpc & vp are running and existing pods in all namespaces.

### **Q:** MountVolume.SetUp failed ?
- Re-install SFS CSI driver from Flexible Engine Github. Be sure to Uninstall existing one first. 
- Check if you specified the right **project & project name** on [cloud-config](https://github.com/huaweicloud/huaweicloud-csi-driver/blob/master/deploy/cloud-config) file

