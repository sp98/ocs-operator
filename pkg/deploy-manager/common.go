package deploymanager

import (
	"fmt"
	"time"

	"github.com/onsi/gomega"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
)

// CreateNamespace creates a namespace in the cluster, ignoring if it already exists
func (t *DeployManager) CreateNamespace(namespace string) error {
	label := make(map[string]string)
	// Label required for monitoring this namespace
	label["openshift.io/cluster-monitoring"] = "true"
	ns := &k8sv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   namespace,
			Labels: label,
		},
	}
	_, err := t.k8sClient.CoreV1().Namespaces().Create(ns)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

// DeleteStorageClusterAndWait deletes a storageClusterCR and waits on it to terminate
func (t *DeployManager) DeleteStorageClusterAndWait(namespace string) error {
	err := t.deleteStorageCluster()
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	cephClusters, err := t.rookClient.CephV1().CephClusters(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, cephCluster := range cephClusters.Items {
		_, err = t.rookClient.CephV1().CephClusters(namespace).Patch(cephCluster.GetName(), types.JSONPatchType, []byte(finalizerRemovalPatch))
		if err != nil {
			return err
		}
	}

	timeout := 600 * time.Second
	interval := 10 * time.Second

	// Wait for storagecluster and cephCluster to terminate
	err = utilwait.PollImmediate(interval, timeout, func() (done bool, err error) {
		cephClusters, err := t.rookClient.CephV1().CephClusters(namespace).List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}
		if len(cephClusters.Items) != 0 {
			return false, nil
		}
		_, err = t.getStorageCluster()
		if !errors.IsNotFound(err) {
			return false, nil
		}
		return true, nil
	})

	return err
}

// DeleteNamespaceAndWait deletes a namespace and waits on it to terminate
func (t *DeployManager) DeleteNamespaceAndWait(namespace string) error {
	err := t.DeleteStorageClusterAndWait(namespace)
	if err != nil {
		return err
	}
	err = t.k8sClient.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	timeout := 600 * time.Second
	interval := 10 * time.Second

	// Wait for namespace to terminate
	err = utilwait.PollImmediate(interval, timeout, func() (done bool, err error) {
		_, err = t.k8sClient.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
		if !errors.IsNotFound(err) {
			return false, nil
		}
		return true, nil
	})

	return err
}

// GetDeploymentImage returns the deployment image name for the deployment
func (t *DeployManager) GetDeploymentImage(name string) (string, error) {
	deployment, err := t.k8sClient.AppsV1().Deployments(InstallNamespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return deployment.Spec.Template.Spec.Containers[0].Image, nil
}

// WaitForPVCBound waits for a pvc with a given name and namespace to reach BOUND phase
func (t *DeployManager) WaitForPVCBound(pvcName string, pvcNamespace string) {
	gomega.Eventually(func() error {
		pvc, err := t.k8sClient.CoreV1().PersistentVolumeClaims(pvcNamespace).Get(pvcName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if pvc.Status.Phase == k8sv1.ClaimBound {
			return nil
		}
		return fmt.Errorf("Waiting on pvc %s/%s to reach bound state when it is currently %s", pvcNamespace, pvcName, pvc.Status.Phase)
	}, 200*time.Second, 1*time.Second).ShouldNot(gomega.HaveOccurred())
}

// WaitForJobSucceeded waits for a Job with a given name and namespace to succeed until 200 seconds
func (t *DeployManager) WaitForJobSucceeded(jobName string, jobNamespace string) {
	gomega.Eventually(func() error {
		job, err := t.k8sClient.BatchV1().Jobs(jobNamespace).Get(jobName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if job.Status.Succeeded > 0 {
			return nil
		}
		return fmt.Errorf("Waiting on job %s/%s to succeed when it is currently %d", jobName, jobNamespace, job.Status.Succeeded)
	},
		200*time.Second, 1*time.Second).Should(gomega.Succeed())
}
