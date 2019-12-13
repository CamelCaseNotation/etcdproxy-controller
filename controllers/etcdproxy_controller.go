/*

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

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	// utilerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	// _ "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	_ "sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	// TODO: Choose between the below 2 options for recording events
	// "k8s.io/client-go/tools/record"
	// "sigs.k8s.io/controller-runtime/pkg/recorder"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	// _ "sigs.k8s.io/controller-runtime/pkg/source"

	etcdv1alpha1 "github.com/camelcasenotation/etcdproxy-controller/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	// Below import needed for PodPreset creation
	// settingsv1alpha1 "k8s.io/api/settings/v1alpha1"
)

// EtcdProxyReconciler reconciles a EtcdProxy object
type EtcdProxyReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=etcd.camelcasenotation.io,resources=etcdproxies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=etcd.camelcasenotation.io,resources=etcdproxies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile handles EtcdProxy objects
func (r *EtcdProxyReconciler) Reconcile(req ctrl.Request) (_ ctrl.Result, reterr error) {
	ctx := context.Background()
	log := r.Log.WithValues("etcdproxy", req.Name)

	// Get EtcdProxy object
	ep := &etcdv1alpha1.EtcdProxy{}
	err := r.Get(ctx, req.NamespacedName, ep)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return reconcile.Result{}, err
	}

	// Don't do anything if it has been deleted
	// NOTE: May not be necessary, so commenting out... will need to test if necessary with tests
	// if !ep.DeletionTimestamp.IsZero() {
	// 	log.Info("EtcdProxy %s is being terminated, not reconciling", ep.Name)
	// 	return reconcile.Result{}, nil
	// }

	secret := &corev1.Secret{}
	err = r.Get(ctx, ep.ClientSecret(), secret)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return r.createClientSecret(ctx, log, ep)
		}
		return reconcile.Result{}, err
	}

	service := &corev1.Service{}
	err = r.Get(ctx, ep.Service(), service)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return r.createService(ctx, log, ep)
		}
		return reconcile.Result{}, err
	}

	deployment := &appsv1.Deployment{}
	err = r.Get(ctx, ep.Deployment(), deployment)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return r.createDeployment(ctx, log, ep)
		}
		return reconcile.Result{}, err
	}

	// TODO: Propagate changes from EtcdProxy resource to above resources, mainly Service & Deployment

	// TODO: Clean this up
	// return ctrl.Result{}, utilerrors.NewAggregate(errs)
	return reconcile.Result{}, nil
}

func (r *EtcdProxyReconciler) createClientSecret(ctx context.Context, log logr.Logger, ep *etcdv1alpha1.EtcdProxy) (reconcile.Result, error) {
	log.Info(fmt.Sprintf("Creating Secret: %s", ep.ClientSecret().String()))
	err := r.Client.Create(ctx, newClientCertSecret(ep))
	return reconcile.Result{}, err
}

func (r *EtcdProxyReconciler) createService(ctx context.Context, log logr.Logger, ep *etcdv1alpha1.EtcdProxy) (reconcile.Result, error) {
	log.Info(fmt.Sprintf("Creating Service: %s", ep.Service().String()))
	err := r.Client.Create(ctx, newService(ep))
	return reconcile.Result{}, err
}

func (r *EtcdProxyReconciler) createDeployment(ctx context.Context, log logr.Logger, ep *etcdv1alpha1.EtcdProxy) (reconcile.Result, error) {
	log.Info(fmt.Sprintf("Creating Deployment: %s", ep.Deployment().String()))
	err := r.Client.Create(ctx, newDeployment(ep))
	return reconcile.Result{}, err
}

// SetupWithManager defines what resources this controller should Reconcile against.
// It will trigger on events for resources it owns: Deployments, Services, & Secrets
func (r *EtcdProxyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&etcdv1alpha1.EtcdProxy{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		// Prevents Reconcile triggers from Deployment events
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
