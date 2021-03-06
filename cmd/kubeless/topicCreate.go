/*
Copyright (c) 2016-2017 Bitnami

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

package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/spf13/cobra"

	k8scmd "k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

var topicCreateCmd = &cobra.Command{
	Use:   "create <topic_name> FLAG",
	Short: "create a topic to Kubeless",
	Long:  `create a topic to Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - topic name")
		}
		ctlNamespace, err := cmd.Flags().GetString("kafka-namespace")
		if err != nil {
			logrus.Fatal(err)
		}

		topicName := args[0]
		command := []string{"bash", "/opt/bitnami/kafka/bin/kafka-topics.sh", "--zookeeper", "zookeeper." + ctlNamespace + ":2181", "--replication-factor", "1", "--partitions", "1", "--create", "--topic", topicName}

		execCommand(command, ctlNamespace)
	},
}

// wrapper of kubectl exec
// execCommand executes a command to kafka pod
func execCommand(command []string, ctlNamespace string) {
	f := cmdutil.NewFactory(nil)

	k8sClientSet := utils.GetClientOutOfCluster()
	pods, err := utils.GetPodsByLabel(k8sClientSet, ctlNamespace, "kubeless", "kafka")
	if err != nil {
		logrus.Fatalf("Can't find the kafka pod: %v", err)
	} else if len(pods.Items) == 0 {
		logrus.Fatalln("Can't find any kafka pod")
	}
	params := &k8scmd.ExecOptions{
		StreamOptions: k8scmd.StreamOptions{
			Namespace:     ctlNamespace,
			PodName:       pods.Items[0].Name,
			ContainerName: "broker",
			In:            nil,
			Out:           os.Stdout,
			Err:           os.Stderr,
			TTY:           false,
		},
		Executor: &k8scmd.DefaultRemoteExecutor{},
		Command:  command,
	}
	config, err := f.ClientConfig()
	if err != nil {
		logrus.Fatalln(err)
	}
	params.Config = config

	fClientset, err := f.ClientSet()
	if err != nil {
		logrus.Fatalln(err)
	}
	params.PodClient = fClientset.Core()

	if err := params.Run(); err != nil {
		logrus.Fatalln(err)
	}
}
