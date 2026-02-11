package kubernetes

import (
	"context"
	"fmt"
	"time"

	"emperror.dev/errors"
	"github.com/apex/log"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pterodactyl/wings/environment"
)

// Create creates the necessary Kubernetes resources for running the server.
// This includes creating the PVC, Pod, and Service.
func (e *Environment) Create() error {
	ctx := context.Background()

	e.log().Info("creating kubernetes resources for server")

	// Step 1: Create PVC
	pvc := e.buildPVCSpec()
	_, err := e.clientset.CoreV1().PersistentVolumeClaims(e.namespace).Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil && !kerrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "kubernetes: failed to create PVC")
	}
	e.log().WithField("pvc", pvc.Name).Debug("created or verified PVC")

	// Step 2: Create Service
	service := e.buildServiceSpec()
	_, err = e.clientset.CoreV1().Services(e.namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil && !kerrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "kubernetes: failed to create Service")
	}
	e.log().WithField("service", service.Name).Debug("created or verified Service")

	// Step 3: Pod is created later during Start()
	e.log().Info("kubernetes resources created successfully")

	return nil
}

// Exists checks if the Pod for this server exists in the cluster.
func (e *Environment) Exists() (bool, error) {
	ctx := context.Background()

	_, err := e.clientset.CoreV1().Pods(e.namespace).Get(ctx, e.getPodName(), metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "kubernetes: failed to check if pod exists")
	}

	return true, nil
}

// IsRunning determines if the server's pod is currently running.
func (e *Environment) IsRunning(ctx context.Context) (bool, error) {
	pod, err := e.clientset.CoreV1().Pods(e.namespace).Get(ctx, e.getPodName(), metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "kubernetes: failed to get pod status")
	}

	// Check if pod is in Running phase and container is actually running
	if pod.Status.Phase == corev1.PodRunning {
		// Verify the main container is running
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Name == "server" && cs.State.Running != nil {
				return true, nil
			}
		}
	}

	return false, nil
}

// Start starts the server by creating the Pod in the cluster.
func (e *Environment) Start(ctx context.Context) error {
	e.log().Info("starting server (creating pod)")

	// Check if pod already exists
	exists, err := e.Exists()
	if err != nil {
		return err
	}

	if exists {
		// Pod already exists, check if it's running
		running, err := e.IsRunning(ctx)
		if err != nil {
			return err
		}
		if running {
			e.log().Warn("pod already running")
			e.SetState(environment.ProcessRunningState)
			return nil
		}
		// If pod exists but not running, delete it first
		e.log().Debug("pod exists but not running, deleting before recreating")
		if err := e.deletePod(ctx); err != nil {
			return err
		}
	}

	// Set state to starting
	e.SetState(environment.ProcessStartingState)

	// Create the Pod
	pod := e.buildPodSpec()
	_, err = e.clientset.CoreV1().Pods(e.namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		e.SetState(environment.ProcessOfflineState)
		return errors.Wrap(err, "kubernetes: failed to create pod")
	}

	e.log().WithField("pod", pod.Name).Info("pod created successfully")

	// Wait for pod to be running (with timeout)
	// This is done async to not block the start operation
	go e.waitForRunning(ctx)

	return nil
}

// Stop stops the server by deleting the Pod.
func (e *Environment) Stop(ctx context.Context) error {
	e.log().Info("stopping server (deleting pod)")

	// Check if pod exists
	exists, err := e.Exists()
	if err != nil {
		return err
	}

	if !exists {
		e.log().Debug("pod does not exist, already stopped")
		e.SetState(environment.ProcessOfflineState)
		return nil
	}

	// Set state to stopping
	e.SetState(environment.ProcessStoppingState)

	// Delete the pod gracefully
	if err := e.deletePod(ctx); err != nil {
		return err
	}

	e.log().Info("pod deleted successfully")
	e.SetState(environment.ProcessOfflineState)

	return nil
}

// Destroy removes all Kubernetes resources associated with this server.
func (e *Environment) Destroy() error {
	ctx := context.Background()

	e.log().Info("destroying kubernetes resources for server")

	// Delete Pod
	if err := e.deletePod(ctx); err != nil {
		e.log().WithError(err).Warn("failed to delete pod during destroy")
	}

	// Delete Service
	err := e.clientset.CoreV1().Services(e.namespace).Delete(ctx, e.getServiceName(), metav1.DeleteOptions{})
	if err != nil && !kerrors.IsNotFound(err) {
		e.log().WithError(err).Warn("failed to delete service during destroy")
	}

	// Delete PVC
	err = e.clientset.CoreV1().PersistentVolumeClaims(e.namespace).Delete(ctx, e.getPVCName(), metav1.DeleteOptions{})
	if err != nil && !kerrors.IsNotFound(err) {
		return errors.Wrap(err, "kubernetes: failed to delete PVC")
	}

	e.log().Info("kubernetes resources destroyed successfully")
	e.SetState(environment.ProcessOfflineState)

	return nil
}

// deletePod deletes the pod with a grace period.
func (e *Environment) deletePod(ctx context.Context) error {
	gracePeriodSeconds := int64(30)
	deleteOptions := metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}

	err := e.clientset.CoreV1().Pods(e.namespace).Delete(ctx, e.getPodName(), deleteOptions)
	if err != nil && !kerrors.IsNotFound(err) {
		return errors.Wrap(err, "kubernetes: failed to delete pod")
	}

	return nil
}

// waitForRunning waits for the pod to be in running state and updates the state accordingly.
func (e *Environment) waitForRunning(ctx context.Context) {
	e.log().Debug("waiting for pod to be running")

	// Poll pod status
	for i := 0; i < 60; i++ { // Poll for up to 60 seconds
		running, err := e.IsRunning(ctx)
		if err != nil {
			e.log().WithError(err).Error("failed to check if pod is running")
			e.SetState(environment.ProcessOfflineState)
			return
		}

		if running {
			e.log().Info("pod is now running")
			e.SetState(environment.ProcessRunningState)
			return
		}

		// Check if pod failed
		pod, err := e.clientset.CoreV1().Pods(e.namespace).Get(ctx, e.getPodName(), metav1.GetOptions{})
		if err != nil {
			e.log().WithError(err).Error("failed to get pod status")
			e.SetState(environment.ProcessOfflineState)
			return
		}

		if pod.Status.Phase == corev1.PodFailed {
			e.log().WithField("reason", pod.Status.Reason).Error("pod failed to start")
			e.SetState(environment.ProcessOfflineState)
			return
		}

		// Wait before next poll
		select {
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Second):
		}
	}

	e.log().Warn("timed out waiting for pod to be running")
	// Don't change state - let it remain in starting state
}

// log returns a logger instance for this environment.
func (e *Environment) log() *log.Entry {
	return log.WithField("environment", "kubernetes").WithField("server_id", e.id)
}
