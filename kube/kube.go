package kube

import (
	"log"
	"time"
	"github.com/techswarn/playworker/database"
    k "github.com/techswarn/k8slib"
	"k8s.io/client-go/kubernetes"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"errors"
	"context"

)

var cs *kubernetes.Clientset
func init() {
    cs = k.Connect()
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
	Instance := &k.Instance {
		Name: d.Name,
	    Image: d.Image,
	    Namespace: d.Namespace,
	}
	log.Printf("Instance : %#v \n", Instance)
	namespace := Instance.CreateNamespace(ctx, cs)
	log.Printf("Creating namespace %#v \n", namespace)
 	
    deployment := Instance.CreateDeployment(ctx, cs, namespace)
	s := Instance.WaitForReadyReplicas(ctx, cs, deployment)
    log.Printf("Replicas ready %#v \n", s)
	var deploy Deploy
	//Once the replica is ready mark the deployment status in mysql to true
	updateDeployStatus := database.DB.Model(deploy).Where("id = ?", id).Update("status", true)
	if updateDeployStatus.RowsAffected == 0 {
		errors.New("Deployment not found in table")
	}
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
