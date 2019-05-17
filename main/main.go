// Copyright (c) 2018 Couchbase, Inc.
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
// except in compliance with the License. You may obtain a copy of the License at
//   http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software distributed under the
// License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
// either express or implied. See the License for the specific language governing permissions
// and limitations under the License.

package main

import (
	"flag"
	"fmt"
	"github.com/nelio2k/xdcrDiffer/base"
	"os"
)

var done = make(chan bool)

var options struct {
	sourceUrl                         string
	sourceUsername                    string
	sourcePassword                    string
	sourceBucketName                  string
	remoteClusterName                 string
	sourceFileDir                     string
	targetUrl                         string
	targetUsername                    string
	targetPassword                    string
	targetBucketName                  string
	targetFileDir                     string
	numberOfSourceDcpClients          uint64
	numberOfWorkersPerSourceDcpClient uint64
	numberOfTargetDcpClients          uint64
	numberOfWorkersPerTargetDcpClient uint64
	numberOfWorkersForFileDiffer      uint64
	numberOfWorkersForMutationDiffer  uint64
	numberOfBins                      uint64
	numberOfFileDesc                  uint64
	// the duration that the tools should be run, in minutes
	completeByDuration uint64
	// whether tool should complete after processing all mutations at tool start time
	completeBySeqno bool
	// directory for checkpoint files
	checkpointFileDir string
	// name of source cluster checkpoint file to load from when tool starts
	// if not specified, source cluster will start from 0
	oldSourceCheckpointFileName string
	// name of target cluster checkpoint file to load from when tool starts
	// if not specified, target cluster will start from 0
	oldTargetCheckpointFileName string
	// name of new checkpoint file to write to when tool shuts down
	// if not specified, tool will not save checkpoint files
	newCheckpointFileName string
	// directory for storing diffs generated by file differ
	fileDifferDir string
	// input directory for mutation differ
	// if this directory is not specified, it indicates that mutation differ is expected to read diff keys generated by file differ,
	// i.e., diffFileDir/base.DiffKeysFileName
	// if this directory is specified, it indicates that mutation differ is expected to read diff keys generated by mutation differ itself
	// i.e., inputDiffKeysFileDir/base.MutationDiffKeysFileName
	inputDiffKeysFileDir string
	// output directory for mutation differ
	mutationDifferDir string
	// size of batch used by mutation differ
	mutationDifferBatchSize uint64
	// timeout, in seconds, used by mutation differ
	mutationDifferTimeout uint64
	// size of source dcp handler channel
	sourceDcpHandlerChanSize uint64
	// size of target dcp handler channel
	targetDcpHandlerChanSize uint64
	// timeout for bucket for stats collection, in seconds
	bucketOpTimeout uint64
	// max number of retry for get stats
	maxNumOfGetStatsRetry uint64
	// max number of retry for send batch
	maxNumOfSendBatchRetry uint64
	// retry interval for get stats, in seconds
	getStatsRetryInterval uint64
	// retry interval for send batch, in milliseconds
	sendBatchRetryInterval uint64
	// max backoff for get stats, in seconds
	getStatsMaxBackoff uint64
	// max backoff for send batch, in seconds
	sendBatchMaxBackoff uint64
	// delay between source cluster start up and target cluster start up, in seconds
	delayBetweenSourceAndTarget uint64
	//interval for periodical checkpointing, in seconds
	// value of 0 indicates no periodical checkpointing
	checkpointInterval uint64
	// whether to run data generation
	runDataGeneration bool
	// whether to run file differ
	runFileDiffer bool
	// whether to verify diff keys through aysnc Get on clusters
	runMutationDiffer bool
}

func argParse() {
	flag.StringVar(&options.sourceUrl, "sourceUrl", "",
		"url for source cluster")
	flag.StringVar(&options.sourceUsername, "sourceUsername", "",
		"username for source cluster")
	flag.StringVar(&options.sourcePassword, "sourcePassword", "",
		"password for source cluster")
	flag.StringVar(&options.sourceBucketName, "sourceBucketName", "",
		"bucket name for source cluster")
	flag.StringVar(&options.remoteClusterName, "remoteClusterName", "",
		"Remote cluster reference name used when creating it")
	flag.StringVar(&options.sourceFileDir, "sourceFileDir", base.SourceFileDir,
		"directory to store mutations in source cluster")
	flag.StringVar(&options.targetUrl, "targetUrl", "",
		"url for target cluster")
	flag.StringVar(&options.targetUsername, "targetUsername", "",
		"username for target cluster")
	flag.StringVar(&options.targetPassword, "targetPassword", "",
		"password for target cluster")
	flag.StringVar(&options.targetBucketName, "targetBucketName", "",
		"bucket name for target cluster")
	flag.StringVar(&options.targetFileDir, "targetFileDir", base.TargetFileDir,
		"directory to store mutations in target cluster")
	flag.Uint64Var(&options.numberOfSourceDcpClients, "numberOfSourceDcpClients", 4,
		"number of source dcp clients")
	flag.Uint64Var(&options.numberOfWorkersPerSourceDcpClient, "numberOfWorkersPerSourceDcpClient", 256,
		"number of workers for each source dcp client")
	flag.Uint64Var(&options.numberOfTargetDcpClients, "numberOfTargetDcpClients", 4,
		"number of target dcp clients")
	flag.Uint64Var(&options.numberOfWorkersPerTargetDcpClient, "numberOfWorkersPerTargetDcpClient", 256,
		"number of workers for each target dcp client")
	flag.Uint64Var(&options.numberOfWorkersForFileDiffer, "numberOfWorkersForFileDiffer", 30,
		"number of worker threads for file differ ")
	flag.Uint64Var(&options.numberOfWorkersForMutationDiffer, "numberOfWorkersForMutationDiffer", 30,
		"number of worker threads for mutation differ ")
	flag.Uint64Var(&options.numberOfBins, "numberOfBins", 10,
		"number of buckets per vbucket")
	flag.Uint64Var(&options.numberOfFileDesc, "numberOfFileDesc", 500,
		"number of file descriptors")
	flag.Uint64Var(&options.completeByDuration, "completeByDuration", 0,
		"duration that the tool should run")
	flag.BoolVar(&options.completeBySeqno, "completeBySeqno", true,
		"whether tool should automatically complete (after processing all mutations at start time)")
	flag.StringVar(&options.checkpointFileDir, "checkpointFileDir", base.CheckpointFileDir,
		"directory for checkpoint files")
	flag.StringVar(&options.oldSourceCheckpointFileName, "oldSourceCheckpointFileName", "",
		"old source checkpoint file to load from when tool starts")
	flag.StringVar(&options.oldTargetCheckpointFileName, "oldTargetCheckpointFileName", "",
		"old target checkpoint file to load from when tool starts")
	flag.StringVar(&options.newCheckpointFileName, "newCheckpointFileName", "",
		"new checkpoint file to write to when tool shuts down")
	flag.StringVar(&options.fileDifferDir, "fileDifferDir", base.FileDifferDir,
		" directory for storing diffs generated by file differ")
	flag.StringVar(&options.inputDiffKeysFileDir, "inputDiffKeysFileDir", "",
		" directory to load diff key file to be used by mutation differ")
	flag.StringVar(&options.mutationDifferDir, "mutationDifferDir", base.MutationDifferDir,
		" output directory for mutation differ")
	flag.Uint64Var(&options.mutationDifferBatchSize, "mutationDifferBatchSize", 100,
		"size of batch used by mutation differ")
	flag.Uint64Var(&options.mutationDifferTimeout, "mutationDifferTimeout", 30,
		"timeout, in seconds, used by mutation differ")
	flag.Uint64Var(&options.sourceDcpHandlerChanSize, "sourceDcpHandlerChanSize", base.DcpHandlerChanSize,
		"size of source dcp handler channel")
	flag.Uint64Var(&options.targetDcpHandlerChanSize, "targetDcpHandlerChanSize", base.DcpHandlerChanSize,
		"size of target dcp handler channel")
	flag.Uint64Var(&options.bucketOpTimeout, "bucketOpTimeout", base.BucketOpTimeout,
		" timeout for bucket for stats collection, in seconds")
	flag.Uint64Var(&options.maxNumOfGetStatsRetry, "maxNumOfGetStatsRetry", base.MaxNumOfGetStatsRetry,
		"max number of retry for get stats")
	flag.Uint64Var(&options.maxNumOfSendBatchRetry, "maxNumOfSendBatchRetry", base.MaxNumOfSendBatchRetry,
		"max number of retry for send batch")
	flag.Uint64Var(&options.getStatsRetryInterval, "getStatsRetryInterval", base.GetStatsRetryInterval,
		" retry interval for get stats, in seconds")
	flag.Uint64Var(&options.sendBatchRetryInterval, "sendBatchRetryInterval", base.SendBatchRetryInterval,
		"retry interval for send batch, in milliseconds")
	flag.Uint64Var(&options.getStatsMaxBackoff, "getStatsMaxBackoff", base.GetStatsMaxBackoff,
		"max backoff for get stats, in seconds")
	flag.Uint64Var(&options.sendBatchMaxBackoff, "sendBatchMaxBackoff", base.SendBatchMaxBackoff,
		"max backoff for send batch, in seconds")
	flag.Uint64Var(&options.delayBetweenSourceAndTarget, "delayBetweenSourceAndTarget", base.DelayBetweenSourceAndTarget,
		"delay between source cluster start up and target cluster start up, in seconds")
	flag.Uint64Var(&options.checkpointInterval, "checkpointInterval", base.CheckpointInterval,
		"interval for periodical checkpointing, in seconds")
	flag.BoolVar(&options.runDataGeneration, "runDataGeneration", true,
		" whether to run data generation")
	flag.BoolVar(&options.runFileDiffer, "runFileDiffer", true,
		" whether to file differ")
	flag.BoolVar(&options.runMutationDiffer, "runMutationDiffer", true,
		" whether to verify diff keys through aysnc Get on clusters")

	flag.Parse()
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage : %s [OPTIONS] \n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	argParse()

	difftool := NewDiffTool()

	if len(options.remoteClusterName) > 0 {
		err := difftool.retrieveReplicationSpecInfo()
		if err != nil {
			os.Exit(1)
		}
	} else {
		difftool.populateTemporarySpecAndRef()
	}
	difftool.setupXDCRCompTopologyMock()

	if options.runDataGeneration {
		//		err := difftool.startStatsMgr()
		//		if err != nil {
		//			difftool.logger.Errorf("Error starting statsMgr. err=%v\n", err)
		//			os.Exit(1)
		//		}
		err := difftool.generateDataFiles()
		if err != nil {
			difftool.logger.Errorf("Error generating data files. err=%v\n", err)
			os.Exit(1)
		}
	} else {
		difftool.logger.Warnf("Skipping  generating data files since it has been disabled\n")
	}

	if options.runFileDiffer {
		err := difftool.diffDataFiles()
		if err != nil {
			fmt.Printf("Error running file difftool. err=%v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Skipping file difftool since it has been disabled\n")
	}

	if options.runMutationDiffer {
		difftool.runMutationDiffer()
	} else {
		fmt.Printf("Skipping mutation diff since it has been disabled\n")
	}
}

func cleanUpAndSetup() error {
	err := os.MkdirAll(options.sourceFileDir, 0777)
	if err != nil {
		fmt.Printf("Error mkdir targetFileDir: %v\n", err)
	}
	err = os.MkdirAll(options.targetFileDir, 0777)
	if err != nil {
		fmt.Printf("Error mkdir targetFileDir: %v\n", err)
	}
	err = os.MkdirAll(options.checkpointFileDir, 0777)
	if err != nil {
		// it is ok for checkpoint dir to be existing, since we do not clean it up
		fmt.Printf("Error mkdir checkpointFileDir: %v\n", err)
	}
	return nil
}
