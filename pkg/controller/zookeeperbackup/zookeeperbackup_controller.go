package zookeeperbackup

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"io"
	"k8s.io/apimachinery/pkg/api/resource"
	"net/http"
	"strconv"
	"strings"
	"time"

	zookeeperv1beta1 "github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/mitchellh/hashstructure/v2"
)

// ReconcileTime is the delay between reconciliations
const ReconcileTime = 60 * time.Second

var log = logf.Log.WithName("controller_zookeeperbackup")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ZookeeperBackup Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileZookeeperBackup{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("zookeeperbackup-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ZookeeperBackup
	err = c.Watch(&source.Kind{Type: &zookeeperv1beta1.ZookeeperBackup{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to resource CronJob and requeue the owner ZookeeperBackup
	err = c.Watch(&source.Kind{Type: &batchv1beta1.CronJob{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &zookeeperv1beta1.ZookeeperBackup{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to resource PVC and requeue the owner ZookeeperBackup
	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &zookeeperv1beta1.ZookeeperBackup{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileZookeeperBackup implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileZookeeperBackup{}

// ReconcileZookeeperBackup reconciles a ZookeeperBackup object
type ReconcileZookeeperBackup struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	log      logr.Logger
}

// Reconcile reads that state of the cluster for a ZookeeperBackup object and makes changes based on the state read
// and what is in the ZookeeperBackup.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileZookeeperBackup) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.log = log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	r.log.Info("Reconciling ZookeeperBackup")

	// Fetch the ZookeeperBackup instance
	zookeeperBackup := &zookeeperv1beta1.ZookeeperBackup{}
	err := r.client.Get(context.TODO(), request.NamespacedName, zookeeperBackup)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	zookeeperBackup.WithDefaults()

	// Define a new PVC object
	pvc := newPVCForZookeeperBackup(zookeeperBackup)

	// Set ZookeeperBackup instance as the owner and controller
	if err := controllerutil.SetControllerReference(zookeeperBackup, pvc, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if PVC already created
	foundPVC := &corev1.PersistentVolumeClaim{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pvc.Name, Namespace: pvc.Namespace}, foundPVC)
	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Creating a new PersistenVolumeClaim")
		err = r.client.Create(context.TODO(), pvc)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Define a new CronJob object
	cronJob := newCronJobForCR(zookeeperBackup)

	// Set ZookeeperBackup instance as the owner and controller
	if err := controllerutil.SetControllerReference(zookeeperBackup, cronJob, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if zookeeper cluster exists
	foundZookeeperCluster := &zookeeperv1beta1.ZookeeperCluster{}
	zkCluster := zookeeperBackup.Spec.ZookeeperCluster
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: zkCluster, Namespace: zookeeperBackup.Namespace}, foundZookeeperCluster)
	if err != nil && errors.IsNotFound(err) {
		r.log.Error(err, fmt.Sprintf("Zookeeper cluster '%s' not found", zkCluster))
		return reconcile.Result{}, err
	}
	if foundZookeeperCluster.Status.Replicas != foundZookeeperCluster.Status.ReadyReplicas {
		r.log.Info(fmt.Sprintf("Not all cluster replicas are ready: %d/%d. Suspend CronJob",
			foundZookeeperCluster.Status.ReadyReplicas, foundZookeeperCluster.Status.Replicas))
		*cronJob.Spec.Suspend = true
	} else {
		*cronJob.Spec.Suspend = false
	}

	// Get zookeeper leader via zookeeper admin server
	leaderIp, err := r.GetLeaderIP(foundZookeeperCluster)
	if err != nil && errors.IsNotFound(err) {
		return reconcile.Result{}, err
	}
	r.log.Info(fmt.Sprintf("Leader IP (hostname): %s", leaderIp))
	leaderHostname := strings.Split(leaderIp, ".")[0]

	// Landing backup pod on the same node with leader
	podList := &corev1.PodList{}
	opts := []client.ListOption{
		client.InNamespace(request.NamespacedName.Namespace),
		client.MatchingLabels{"app": zkCluster},
	}
	err = r.client.List(context.TODO(), podList, opts...)
	if err != nil {
		if errors.IsNotFound(err) {
			msg := fmt.Sprintf("Pods cannot be found by label app:%s", zookeeperBackup.Name)
			r.log.Error(err, msg)
		}
		return reconcile.Result{}, err
	}

	leaderFound := false
	for _, pod := range podList.Items {
		if pod.Spec.Hostname == leaderHostname {
			leaderFound = true
			r.log.Info(fmt.Sprintf("Leader was found. Pod: %s (node: %s)", pod.Name, pod.Spec.NodeName))
			// Set appropriate NodeSelector and PVC ClaimName
			cronJob.Spec.JobTemplate.Spec.Template.Spec.NodeSelector =
				map[string]string{"kubernetes.io/hostname": pod.Spec.NodeName}
			vol := GetVolumeByName(cronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes, "zookeeperbackup-data")
			vol.VolumeSource.PersistentVolumeClaim.ClaimName = "data-" + pod.Name
			break
		}
	}
	if !leaderFound {
		r.log.Info("Pod with leader role wasn't found. Suspend CronJob")
		*cronJob.Spec.Suspend = true
	}

	if cronJob.Annotations == nil {
		cronJob.Annotations = make(map[string]string)
	}

	// Calculate hash of CronJob Spec
	hash, err := hashstructure.Hash(cronJob.Spec, hashstructure.FormatV2, nil)
	if err != nil {
		panic(err)
	}
	hashStr := strconv.FormatUint(hash, 10)

	// Check if this CronJob already exists
	foundCJ := &batchv1beta1.CronJob{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: cronJob.Name, Namespace: cronJob.Namespace}, foundCJ)
	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Creating a new CronJob", "CronJob.Namespace", cronJob.Namespace, "CronJob.Name", cronJob.Name)
		cronJob.Annotations["last-applied-hash"] = hashStr
		err = r.client.Create(context.TODO(), cronJob)
		if err != nil {
			return reconcile.Result{}, err
		}

		// CronJob created successfully
		r.log.Info("CronJob created successfully.", "RequeueAfter", ReconcileTime)
		return reconcile.Result{RequeueAfter: ReconcileTime}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	if foundCJ.Annotations["last-applied-hash"] == hashStr {
		r.log.Info("CronJob already exists and looks updated", "CronJob.Namespace", foundCJ.Namespace, "CronJob.Name", foundCJ.Name)
	} else {
		cronJob.Annotations["last-applied-hash"] = hashStr
		r.log.Info("Update CronJob", "Namespace", cronJob.Namespace, "Name", cronJob.Name)
		//cronJob.ObjectMeta.ResourceVersion = foundCJ.ObjectMeta.ResourceVersion
		err = r.client.Update(context.TODO(), cronJob)
		if err != nil {
			r.log.Error(err, "CronJob cannot be updated")
			return reconcile.Result{}, err
		}
	}

	// Requeue
	r.log.Info(fmt.Sprintf("Rerun reconclie after %s sec.", ReconcileTime))
	return reconcile.Result{RequeueAfter: ReconcileTime}, nil
}

func (r *ReconcileZookeeperBackup) GetLeaderIP(zkCluster *zookeeperv1beta1.ZookeeperCluster) (string, error) {
	// Get zookeeper leader via zookeeper admin server
	svcAdminName := zkCluster.GetAdminServerServiceName()
	foundSvcAdmin := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      svcAdminName,
		Namespace: zkCluster.Namespace,
	}, foundSvcAdmin)
	if err != nil && errors.IsNotFound(err) {
		r.log.Error(err, fmt.Sprintf("Zookeeper admin service '%s' not found", svcAdminName))
		return "", err
	}

	adminIp  := foundSvcAdmin.Spec.ClusterIP
	svcPort := GetServicePortByName(foundSvcAdmin, "tcp-admin-server")

	resp, err := http.Get(fmt.Sprintf("http://%s:%d/commands/leader", adminIp, svcPort.Port))
	if err != nil {
		r.log.Error(err, "Admin service error response")
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.log.Error(err, "Can't read response body")
		return "", err
	}
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		r.log.Error(err, "Can't unmarshal json")
		return "", err
	}
	leaderIp := result["leader_ip"].(string)
	return leaderIp, nil
}

func GetServicePortByName(service *corev1.Service, name string) *corev1.ServicePort {
	for _, port := range service.Spec.Ports {
		if port.Name == name {
			return &port
		}
	}
	return nil
}

func GetVolumeByName(volumes []corev1.Volume, name string) *corev1.Volume {
	for _, vol := range volumes {
		if vol.Name == name {
			return &vol
		}
	}
	return nil
}

// newPVCForZookeeperBackup returns pob definition
func newPVCForZookeeperBackup(cr *zookeeperv1beta1.ZookeeperBackup) *corev1.PersistentVolumeClaim {
	labels := map[string]string{
		"app": cr.Name,
	}
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pvc",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.PersistentVolumeAccessMode("ReadWriteOnce"),
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(cr.Spec.DataCapacity),
				},
			},
			StorageClassName: &cr.Spec.DataStorageClass,
		},
	}
	return pvc
}

// newCronJobForCR returns a cronJob with the same name/namespace as the cr
func newCronJobForCR(cr *zookeeperv1beta1.ZookeeperBackup) *batchv1beta1.CronJob {
	labels := map[string]string{
		"app": cr.Name,
	}
	suspend := true
	backupMountPath := "/var/backup"
	return &batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-backup",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: batchv1beta1.CronJobSpec{
			Schedule: cr.Spec.Schedule,
			Suspend: &suspend,
			JobTemplate: batchv1beta1.JobTemplateSpec {
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec {
							RestartPolicy: "Never",
							Containers: []corev1.Container{
								{
									Name: "run-zookeeperbackup",
									Image: cr.Spec.Image.ToString(),
									ImagePullPolicy: cr.Spec.Image.PullPolicy,
									VolumeMounts: []corev1.VolumeMount{
										{
											Name: "zookeeperbackup-vol",
											MountPath: backupMountPath,
										},
										{
											Name: "zookeeperbackup-data",
											MountPath: "/data",
										},
									},
									Env: []corev1.EnvVar{
										{
											Name:  "BACKUPDIR",
											Value: backupMountPath,
										},
										{
											Name:  "ZOOKEEPERDATADIR",
											Value: "/data/version-2/",
										},
										{
											Name:  "BACKUPS_TO_KEEP",
											Value: cr.Spec.BackupsToKeep,
										},
									},
									Command: []string{"/zookeeper/backup.sh"},
								},
							},
							NodeName: "",
							Volumes: []corev1.Volume{
								{
									Name: "zookeeperbackup-vol",
									VolumeSource: corev1.VolumeSource{
										PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
											ClaimName: cr.Name + "-pvc",
										},
									},
								},
								{
									Name: "zookeeperbackup-data",
									VolumeSource: corev1.VolumeSource{
										PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
											ClaimName: "",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
