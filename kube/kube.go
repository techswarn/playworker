package kube

import (
	"log"
	"time"
	"github.com/techswarn/playworker/database"
	"github.com/techswarn/playworker/utils"

	"k8s.io/client-go/kubernetes"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/Azure/go-autorest/autorest/to"

	"errors"
	"context"

)

var cs *kubernetes.Clientset
func init() {
    cs, _ = utils.GetKubehandle()
}
//Struct for deploymnent details
type Deploy struct {
	Id string `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
	Image string `json:"image"`
	Namespace string `json:namespace`
	Status bool `json:status`
	CreatedAt time.Time `json:"createdat"`
}

type Replica struct {
	Name string
	Status bool
}

func CreateDeploy(id string) {
	log.Println("Create deploymnt")
	d, err := GetDeployDetails(id)
	if err != nil {
		errors.New("Deployment not found in table")
	}
	log.Printf(" Deployment details for id %#v \n", d)

	//Create the deploy from the details fetched.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	namespace := createNamespace(ctx, cs, d.Namespace)
	log.Printf("Creating namespace %#v \n", namespace)
 	
    deployment := createDeployment(ctx, cs, namespace, d.Name, d.Image)
	s := waitForReadyReplicas(ctx, cs, deployment)
    log.Printf("Replicas ready %#v \n", s)

}

 func GetDeployDetails(id string) (Deploy , error) {
	 var deploy Deploy
	 result := database.DB.First(&deploy, "id = ?", id)
	 // if the item data is not found, return an error
	if result.RowsAffected == 0 {
		return Deploy{}, errors.New("Deployment not found in table")
	}

	res := Deploy{
		Id: deploy.Id,
		Name: deploy.Name,
		Image: deploy.Image,
		Namespace: deploy.Namespace,
		Status: deploy.Status,
		CreatedAt: deploy.CreatedAt,
	}

	return res, nil
 }

//CREATE DEPLOYMENT
func createDeployment(ctx context.Context, clientSet *kubernetes.Clientset, ns *corev1.Namespace, name string, image string) *appv1.Deployment {
	log.Printf("Printing namespace================= %#s \n", ns.Name)
	var (
		matchLabel = map[string]string{"app": "nginx"}
		objMeta    = metav1.ObjectMeta{
			Name:      name,
			Namespace: ns.Name,
			Labels:    matchLabel,
		}
	)

	deployment := &appv1.Deployment{
		ObjectMeta: objMeta,
		Spec: appv1.DeploymentSpec{
			Replicas: to.Int32Ptr(1),
			Selector: &metav1.LabelSelector{MatchLabels: matchLabel},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: matchLabel,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},						
							},
							Command: []string{"/bin/sh", "-c", "sleep 50000"},
						},
					},
				},
			},
		},
	}
	deployment, err := clientSet.AppsV1().Deployments(ns.Name).Create(ctx, deployment, metav1.CreateOptions{})
	//log.Printf(" deployment: %#v", deployment)
	panicIfError(err)
	return deployment
}

func waitForReadyReplicas(ctx context.Context, clientSet *kubernetes.Clientset, deployment *appv1.Deployment) *Replica {

	log.Printf("Waiting for ready replicas in deployment %q\n", deployment.Name)
	for {
		expectedReplicas := *deployment.Spec.Replicas
		readyReplicas := getReadyReplicasForDeployment(ctx, clientSet, deployment)
		if readyReplicas == expectedReplicas {
			log.Printf("replicas are ready!\n\n")
			return &Replica{
				Name: deployment.Name,
				Status: true,
			}
			break
		}

		log.Printf("replicas are not ready yet. %d/%d\n", readyReplicas, expectedReplicas)
		time.Sleep(1 * time.Second)
	}

	return &Replica{
		Name: "",
		Status: false,
	}
}

func createNamespace(ctx context.Context, clientSet *kubernetes.Clientset, name string) *corev1.Namespace {

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	ns, err := clientSet.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	panicIfError(err)
	return ns
}

func getReadyReplicasForDeployment(ctx context.Context, clientSet *kubernetes.Clientset, deployment *appv1.Deployment) int32 {
	dep, err := clientSet.AppsV1().Deployments(deployment.Namespace).Get(ctx, deployment.Name, metav1.GetOptions{})
	panicIfError(err)

	return dep.Status.ReadyReplicas
}


func panicIfError(err error) {
	if err != nil {
		panic(err.Error())
	}
}
