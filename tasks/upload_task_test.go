package tasks_test

import (
	"github.com/stretchr/testify/assert"
	"ig_server/config"
	"ig_server/models"
	"ig_server/tasks"
	"ig_server/tests"
	"math/big"
	"testing"
	"time"
)

func TestUploadRightTask(t *testing.T) {

	err := tests.SyncToLatestBlock()
	assert.Equal(t, nil, err, "catchup error")

	uploadTaskChan := make(chan int)
	go tasks.StartUploadTaskParamsWithTerminateChannel(uploadTaskChan)

	sendTaskChan := make(chan int)
	go tasks.StartSendTaskOnChainWithTerminateChannel(sendTaskChan)

	getTaskCreationResultChan := make(chan int)
	go tasks.StartGetTaskCreationResultWithTerminateChannel(getTaskCreationResultChan)

	addresses, privateKeys, err := tests.PrepareAccounts()
	assert.Nil(t, err, "error preparing accounts")

	err = tests.PrepareNetwork(addresses, privateKeys)
	assert.Nil(t, err, "error preparing the network")

	err = tests.PrepareTaskCreatorAccount(addresses[0], privateKeys[0])
	assert.Nil(t, err, "error preparing the task creator account")

	task, err := tests.NewTask()
	assert.Nil(t, err, "error creating task")

	time.Sleep(40 * time.Second)
	task = tests.AssertTaskStatus(t, task.ID, models.InferenceTaskParamsUploaded)

	// Task must be finished before clearing the network
	err = tests.SuccessTaskOnChain(big.NewInt(int64(task.TaskId)), addresses, privateKeys)
	assert.Equal(t, nil, err, "error submitting result on chain")

	t.Cleanup(func() {
		uploadTaskChan <- 1
		sendTaskChan <- 1
		getTaskCreationResultChan <- 1
		err := tests.ClearNetwork(addresses, privateKeys)
		assert.Equal(t, nil, err, "error clearing blockchain network")
		tests.ClearDB()
	})
}

func TestUploadDuplicateTask(t *testing.T) {

	err := tests.SyncToLatestBlock()
	assert.Equal(t, nil, err, "catchup error")

	uploadTaskChan := make(chan int)
	go tasks.StartUploadTaskParamsWithTerminateChannel(uploadTaskChan)

	sendTaskChan := make(chan int)
	go tasks.StartSendTaskOnChainWithTerminateChannel(sendTaskChan)

	getTaskCreationResultChan := make(chan int)
	go tasks.StartGetTaskCreationResultWithTerminateChannel(getTaskCreationResultChan)

	addresses, privateKeys, err := tests.PrepareAccounts()
	assert.Nil(t, err, "error preparing accounts")

	err = tests.PrepareNetwork(addresses, privateKeys)
	assert.Nil(t, err, "error preparing the network")

	err = tests.PrepareTaskCreatorAccount(addresses[0], privateKeys[0])
	assert.Nil(t, err, "error preparing the task creator account")

	task, err := tests.NewTask()
	assert.Nil(t, err, "error creating task")

	time.Sleep(40 * time.Second)
	task = tests.AssertTaskStatus(t, task.ID, models.InferenceTaskParamsUploaded)

	// Let's try to upload the task again
	task, err = models.UpdateStatusForTask(big.NewInt(int64(task.TaskId)), models.InferenceTaskBlockchainConfirmed, config.GetDB())
	assert.Nil(t, err, "error updating task status")

	time.Sleep(10 * time.Second)
	task = tests.AssertTaskStatus(t, task.ID, models.InferenceTaskAborted)

	// Task must be finished before clearing the network
	err = tests.SuccessTaskOnChain(big.NewInt(int64(task.TaskId)), addresses, privateKeys)
	assert.Equal(t, nil, err, "error submitting result on chain")

	t.Cleanup(func() {
		uploadTaskChan <- 1
		sendTaskChan <- 1
		getTaskCreationResultChan <- 1
		err := tests.ClearNetwork(addresses, privateKeys)
		assert.Equal(t, nil, err, "error clearing blockchain network")
		tests.ClearDB()
	})
}
