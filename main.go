package main

import (
	"log"
	"context"
	"encoding/json"
	"fmt"
	"time"
	"github.com/go-redis/redis/v8"
	"errors"
	"github.com/techswarn/playworker/kube"
	"github.com/techswarn/playworker/utils"
	"github.com/techswarn/playworker/database"
)

const deployQueueList = "queue"
const deployTempQueueList = "queue-temp"


func main() {
	client := utils.Client()
	_, err := client.Ping(context.Background()).Result()

	if err != nil {
		log.Fatal("ping failed. could not connect", err)
	}
	fmt.Println("reliable consumer ready")
	database.InitDatabase(utils.GetValue("DB_NAME"))
	for {

		val, err := client.BLMove(context.Background(), deployQueueList, deployTempQueueList, "RIGHT", "LEFT", 2*time.Second).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			log.Println("blmove issue", err)
		}

		job := val
		var jobInfo JobInfo

		err = json.Unmarshal([]byte(job), &jobInfo)
		if err != nil {
			log.Fatal("job info unmarshal issue issue", err)
		}
		//processjobs(jobInfo, client, job, err)
		fmt.Println("received job", jobInfo.JobId)
		fmt.Println("Deploying ", jobInfo.DeployID)
		//ADD CODE FOR DEPLOYMENT
		kube.CreateDeploy(jobInfo.DeployID)
		//GET ALL VALUES FROM DATABASE FROM DEPOYMENT ID


		go func() {
			err = client.LRem(context.Background(), "jobs-temp", 0, job).Err()
			if err != nil {
				log.Fatal("lrem failed for", job, "error", err)
			}
			log.Println("removed job from temporary list", job)
		}()
	}

}

type JobInfo struct {
	JobId string `json:"jobid"`
	DeployID string  `json:"deployid"`
}