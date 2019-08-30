package main

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
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

	var body string
	for _, task := range result.ReplicationTasks {
		body += fmt.Sprintf("name: %s, fullload_progress: %d, table_error: %d\n", *task.ReplicationTaskIdentifier, *task.ReplicationTaskStats.FullLoadProgressPercent, *task.ReplicationTaskStats.TablesErrored)
	}

	return body
}

func slack(body string) {
	name := "dms-progress"
	channel := "bf_dms"

	jsonStr := `{"channel":"` + channel + `","username":"` + name + `","text":"` + body + `"}`

	req, err := http.NewRequest(
		"POST",
		os.Getenv("SLACK_INCOMING_URL"),
		bytes.NewBuffer([]byte(jsonStr)),
	)

	if err != nil {
		fmt.Print(err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err)
	}

	fmt.Print(resp)
	defer resp.Body.Close()
}

func main() {
	lambda.Start(_main)
}
