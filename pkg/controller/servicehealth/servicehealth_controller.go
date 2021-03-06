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

package servicehealth

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	ismosbapiv1alpha1 "github.com/pivotal-cf/ism/pkg/apis/osbapi/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ServiceHealth Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileServiceHealth{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("servicehealth-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	instance := ismosbapiv1alpha1.ServiceInstance{}

	// Watch for changes to ServiceInstances
	err = c.Watch(&source.Kind{Type: &instance}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileServiceHealth{}

// ReconcileServiceHealth reconciles a ServiceHealth object
type ReconcileServiceHealth struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ServiceHealth object and makes changes based on the state read
// and what is in the ServiceHealth.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=osbapi.k8s.io,resources=serviceinstance,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=osbapi.k8s.io,resources=serviceinstance/status,verbs=get;update;patch
func (r *ReconcileServiceHealth) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the ServiceHealth instance
	instance := &ismosbapiv1alpha1.ServiceInstance{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if !instance.Status.HealthChecked.IsZero() && instance.Status.HealthChecked.Add(10*time.Second).After(time.Now()) {
		return reconcile.Result{}, nil
	}

	resp, err := http.Get(instance.Status.HealthEndpoint)
	if err != nil {
		return reconcile.Result{}, err
	}

	alive := resp.StatusCode == http.StatusOK

	updatedInstance := instance.DeepCopy()

	updatedInstance.Status.Health = strconv.FormatBool(alive)
	updatedInstance.Status.HealthChecked = time.Now()

	if err := r.Client.Status().Update(context.TODO(), updatedInstance); err != nil {
		return reconcile.Result{}, err
	}

	fmt.Printf("%+v\n", instance)

	return reconcile.Result{RequeueAfter: time.Second * 10}, nil
}
