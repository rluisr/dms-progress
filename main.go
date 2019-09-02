package main

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/databasemigrationservice"
)

func _main() {
	body := getStats()
	slack(body)
}

func getStats() string {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("REGION"))})
	if err != nil {
		panic(err)
	}

	svc := databasemigrationservice.New(sess)

	drti := &databasemigrationservice.DescribeReplicationTasksInput{
		Filters:         nil,
		Marker:          nil,
		MaxRecords:      nil,
		WithoutSettings: nil,
	}

	result, err := svc.DescribeReplicationTasks(drti)
	if err != nil {
		panic(err)
	}

	/*
			{
		  MigrationType: "full-load-and-cdc",
		  ReplicationInstanceArn: "arn:aws:dms:ap-northeast-1:",
		  ReplicationTaskArn: "arn:aws:dms:ap-northeast-1:",
		  ReplicationTaskCreationDate: 2019-07-26 07:53:33 +0000 UTC,
		  ReplicationTaskIdentifier: "",
		  ReplicationTaskSettings: "{\"TargetMetadata\":{\"TargetSchema\":\"\",\"SupportLobs\":true,\"FullLobMode\":true,\"LobChunkSize\":0,\"LimitedSizeLobMode\":false,\"LobMaxSize\":0,\"InlineLobMaxSize\":0,\"LoadMaxFileSize\":0,\"ParallelLoadThreads\":0,\"ParallelLoadBufferSize\":0,\"BatchApplyEnabled\":false,\"TaskRecoveryTableEnabled\":false},\"FullLoadSettings\":{\"TargetTablePrepMode\":\"TRUNCATE_BEFORE_LOAD\",\"CreatePkAfterFullLoad\":false,\"StopTaskCachedChangesApplied\":false,\"StopTaskCachedChangesNotApplied\":false,\"MaxFullLoadSubTasks\":16,\"TransactionConsistencyTimeout\":600,\"CommitRate\":10000},\"Logging\":{\"EnableLogging\":true,\"LogComponents\":[{\"Id\":\"SOURCE_UNLOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_LOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_CAPTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_APPLY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TASK_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"}],\"CloudWatchLogGroup\":\"dms-tasks-prd\",\"CloudWatchLogStream\":\"dms-task-B2H6SCTOBDY6RBBDE6GYRTA7ZI\"},\"ControlTablesSettings\":{\"historyTimeslotInMinutes\":5,\"ControlSchema\":\"\",\"HistoryTimeslotInMinutes\":5,\"HistoryTableEnabled\":false,\"SuspendedTablesTableEnabled\":false,\"StatusTableEnabled\":false},\"StreamBufferSettings\":{\"StreamBufferCount\":3,\"StreamBufferSizeInMB\":8,\"CtrlStreamBufferSizeInMB\":5},\"ChangeProcessingDdlHandlingPolicy\":{\"HandleSourceTableDropped\":true,\"HandleSourceTableTruncated\":true,\"HandleSourceTableAltered\":true},\"ErrorBehavior\":{\"DataErrorPolicy\":\"LOG_ERROR\",\"DataTruncationErrorPolicy\":\"LOG_ERROR\",\"DataErrorEscalationPolicy\":\"SUSPEND_TABLE\",\"DataErrorEscalationCount\":0,\"TableErrorPolicy\":\"SUSPEND_TABLE\",\"TableErrorEscalationPolicy\":\"STOP_TASK\",\"TableErrorEscalationCount\":0,\"RecoverableErrorCount\":-1,\"RecoverableErrorInterval\":5,\"RecoverableErrorThrottling\":true,\"RecoverableErrorThrottlingMax\":1800,\"ApplyErrorDeletePolicy\":\"IGNORE_RECORD\",\"ApplyErrorInsertPolicy\":\"LOG_ERROR\",\"ApplyErrorUpdatePolicy\":\"LOG_ERROR\",\"ApplyErrorEscalationPolicy\":\"LOG_ERROR\",\"ApplyErrorEscalationCount\":0,\"ApplyErrorFailOnTruncationDdl\":false,\"FullLoadIgnoreConflicts\":true,\"FailOnTransactionConsistencyBreached\":false,\"FailOnNoTablesCaptured\":false},\"ChangeProcessingTuning\":{\"BatchApplyPreserveTransaction\":true,\"BatchApplyTimeoutMin\":1,\"BatchApplyTimeoutMax\":30,\"BatchApplyMemoryLimit\":500,\"BatchSplitSize\":0,\"MinTransactionSize\":1000,\"CommitTimeout\":1,\"MemoryLimitTotal\":1024,\"MemoryKeepTime\":60,\"StatementCacheSize\":50},\"ValidationSettings\":{\"EnableValidation\":true,\"ValidationMode\":\"ROW_LEVEL\",\"ThreadCount\":5,\"PartitionSize\":10000,\"FailureMaxCount\":10000,\"RecordFailureDelayInMinutes\":5,\"RecordSuspendDelayInMinutes\":30,\"MaxKeyColumnSize\":8096,\"TableFailureMaxCount\":1000,\"ValidationOnly\":false,\"HandleCollationDiff\":false,\"RecordFailureDelayLimitInMinutes\":0},\"PostProcessingRules\":null,\"CharacterSetSettings\":null,\"LoopbackPreventionSettings\":null}",
		  ReplicationTaskStartDate: 2019-09-01 10:56:39 +0000 UTC,
		  ReplicationTaskStats: {
		    ElapsedTimeMillis: 63496609,
		    FullLoadProgressPercent: 100,
		    TablesErrored: 0,
		    TablesLoaded: 1187,
		    TablesLoading: 0,
		    TablesQueued: 0
		  },
		  SourceEndpointArn: "arn:aws:dms:ap-northeast-1::",
		  Status: "running",
		  TableMappings: "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"\",\"table-name\":\"%\"},\"rule-action\":\"include\",\"filters\":[]}]}",
		  TargetEndpointArn: "arn:aws:dms:ap-northeast-1:"
		}
	*/
	var body string
	for _, task := range result.ReplicationTasks {
		body += fmt.Sprintf("name: %s, status: %s, fullload_progress: %d, table_error: %d\n", *task.ReplicationTaskIdentifier, *task.Status, *task.ReplicationTaskStats.FullLoadProgressPercent, *task.ReplicationTaskStats.TablesErrored)
	}

	return body
}

func slack(body string) {
	name := "dms-progress"
	channel := os.Getenv("SLACK_CHANNEL")

	jsonStr := `{"channel":"` + channel + `","username":"` + name + `","text":"` + body + `"}`

	req, _ := http.NewRequest(
		"POST",
		os.Getenv("SLACK_INCOMING_URL"),
		bytes.NewBuffer([]byte(jsonStr)),
	)

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		log.Fatalf("response code is not 200 got: %d\n%s", resp.StatusCode, string(b))
	}
}

func main() {
	lambda.Start(_main)
	//_main()
}
