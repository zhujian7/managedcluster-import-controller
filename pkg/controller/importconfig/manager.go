// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package importconfig

import (
	"github.com/stolostron/managedcluster-import-controller/pkg/helpers"
	"github.com/stolostron/managedcluster-import-controller/pkg/source"
	clusterv1 "open-cluster-management.io/api/cluster/v1"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const controllerName = "importconfig-controller"

// Add creates a new importconfig controller and adds it to the Manager.
// The Manager will set fields on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager, clientHolder *helpers.ClientHolder, informerHolder *source.InformerHolder) (string, error) {

	// if err := c.Watch(
	// 	&runtimesource.Kind{Type: &rbacv1.ClusterRole{}},
	// 	&handler.EnqueueRequestForOwner{
	// 		IsController: true,
	// 		OwnerType:    &clusterv1.ManagedCluster{},
	// 	},
	// 	predicate.Predicate(predicate.Funcs{
	// 		GenericFunc: func(e event.GenericEvent) bool { return false },
	// 		CreateFunc:  func(e event.CreateEvent) bool { return false },
	// 		DeleteFunc:  func(e event.DeleteEvent) bool { return true },
	// 		UpdateFunc:  func(e event.UpdateEvent) bool { return true },
	// 	}),
	// ); err != nil {
	// 	return controllerName, err
	// }

	// if err := c.Watch(
	// 	&runtimesource.Kind{Type: &rbacv1.ClusterRoleBinding{}},
	// 	&handler.EnqueueRequestForOwner{
	// 		IsController: true,
	// 		OwnerType:    &clusterv1.ManagedCluster{},
	// 	},
	// 	predicate.Predicate(predicate.Funcs{
	// 		GenericFunc: func(e event.GenericEvent) bool { return false },
	// 		CreateFunc:  func(e event.CreateEvent) bool { return false },
	// 		DeleteFunc:  func(e event.DeleteEvent) bool { return true },
	// 		UpdateFunc:  func(e event.UpdateEvent) bool { return true },
	// 	}),
	// ); err != nil {
	// 	return controllerName, err
	// }

	// if err := c.Watch(
	// 	&runtimesource.Kind{Type: &corev1.ServiceAccount{}},
	// 	&handler.EnqueueRequestForOwner{
	// 		IsController: true,
	// 		OwnerType:    &clusterv1.ManagedCluster{},
	// 	},
	// 	predicate.Predicate(predicate.Funcs{
	// 		GenericFunc: func(e event.GenericEvent) bool { return false },
	// 		CreateFunc:  func(e event.CreateEvent) bool { return false },
	// 		DeleteFunc:  func(e event.DeleteEvent) bool { return true },
	// 		UpdateFunc:  func(e event.UpdateEvent) bool { return true },
	// 	}),
	// ); err != nil {
	// 	return controllerName, err
	// }

	// if err := c.Watch(
	// 	source.NewImportSecretSource(informerHolder.ImportSecretInformer),
	// 	&source.ManagedClusterResourceEventHandler{},
	// 	predicate.Predicate(predicate.Funcs{
	// 		GenericFunc: func(e event.GenericEvent) bool { return false },
	// 		CreateFunc:  func(e event.CreateEvent) bool { return false },
	// 		DeleteFunc:  func(e event.DeleteEvent) bool { return true },
	// 		UpdateFunc:  func(e event.UpdateEvent) bool { return true },
	// 	}),
	// ); err != nil {
	// 	return controllerName, err
	// }

	err := ctrl.NewControllerManagedBy(mgr).Named(controllerName).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: helpers.GetMaxConcurrentReconciles(),
		}).
		Watches(
			&clusterv1.ManagedCluster{},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(predicate.Funcs{
				GenericFunc: func(e event.GenericEvent) bool { return false },
				DeleteFunc:  func(e event.DeleteEvent) bool { return false },
				CreateFunc:  func(e event.CreateEvent) bool { return true },
				UpdateFunc: func(e event.UpdateEvent) bool {
					// handle the labels changes for image registry
					// handle the annotations changes for node placement
					return !equality.Semantic.DeepEqual(e.ObjectOld.GetLabels(), e.ObjectNew.GetLabels()) ||
						!equality.Semantic.DeepEqual(e.ObjectOld.GetAnnotations(), e.ObjectNew.GetAnnotations())
				},
			}),
		).
		Watches(
			&rbacv1.ClusterRole{},
			handler.EnqueueRequestForOwner(
				mgr.GetScheme(),
				mgr.GetRESTMapper(),
				&clusterv1.ManagedCluster{},
				handler.OnlyControllerOwner(),
			),
			builder.WithPredicates(predicate.Funcs{
				GenericFunc: func(e event.GenericEvent) bool { return false },
				CreateFunc:  func(e event.CreateEvent) bool { return false },
				DeleteFunc:  func(e event.DeleteEvent) bool { return true },
				UpdateFunc:  func(e event.UpdateEvent) bool { return true },
			}),
		).
		Watches(
			&rbacv1.ClusterRoleBinding{},
			handler.EnqueueRequestForOwner(
				mgr.GetScheme(),
				mgr.GetRESTMapper(),
				&clusterv1.ManagedCluster{},
				handler.OnlyControllerOwner(),
			),
			builder.WithPredicates(predicate.Funcs{
				GenericFunc: func(e event.GenericEvent) bool { return false },
				CreateFunc:  func(e event.CreateEvent) bool { return false },
				DeleteFunc:  func(e event.DeleteEvent) bool { return true },
				UpdateFunc:  func(e event.UpdateEvent) bool { return true },
			}),
		).
		Watches(
			&corev1.ServiceAccount{},
			handler.EnqueueRequestForOwner(
				mgr.GetScheme(),
				mgr.GetRESTMapper(),
				&clusterv1.ManagedCluster{},
				handler.OnlyControllerOwner(),
			),
			builder.WithPredicates(predicate.Funcs{
				GenericFunc: func(e event.GenericEvent) bool { return false },
				CreateFunc:  func(e event.CreateEvent) bool { return false },
				DeleteFunc:  func(e event.DeleteEvent) bool { return true },
				UpdateFunc:  func(e event.UpdateEvent) bool { return true },
			}),
		).
		WatchesRawSource(
			source.NewImportSecretSource(informerHolder.ImportSecretInformer),
			&source.ManagedClusterResourceEventHandler{},
			builder.WithPredicates(predicate.Funcs{
				GenericFunc: func(e event.GenericEvent) bool { return false },
				CreateFunc:  func(e event.CreateEvent) bool { return false },
				DeleteFunc:  func(e event.DeleteEvent) bool { return true },
				UpdateFunc:  func(e event.UpdateEvent) bool { return true },
			}),
		).
		Complete(&ReconcileImportConfig{
			clientHolder:  clientHolder,
			scheme:        mgr.GetScheme(),
			recorder:      helpers.NewEventRecorder(clientHolder.KubeClient, controllerName),
			workerFactory: &workerFactory{clientHolder: clientHolder},
		})
	return controllerName, err
}
