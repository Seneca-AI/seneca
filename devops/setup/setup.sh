#!/bin/bash
# TODO(lucaloncar): fix this

# Set path variables that are always necessary.
export GOPATH=$HOME/go
export PATH=$PATH:$HOME/go/bin
export GO_VERSION=1.16.2
export PROTOC_ZIP=protoc-3.14.0-linux-x86_64.zip
export DEBIAN_FRONTEND=noninteractive

# Define a spinner.
spin() {
  spinner="/|\\-/|\\-"
  while :
  do
    for i in `seq 0 7`
    do
      echo -n "${spinner:$i:1}"
      echo -en "\010"
      sleep 1
    done
  done
}

# Download and install all necessary dependencies.
setup() {
	echo "Output stored in setup.log"
	touch setup.log

	read -p "Enter GitHub token: " GITHUB_TOKEN

	echo "Installing wget"
	sudo apt-get update > setup.log
	sudo apt-get upgrade -y > setup.log
	sudo apt-get install wget -y > setup.log

	echo "Installing golang"
	
	wget -c https://golang.org/dl/go$GO_VERSION.linux-amd64.tar.gz > setup.log
	sudo tar -C $HOME -xzf go$GO_VERSION.linux-amd64.tar.gz
	rm go$GO_VERSION*

	echo "Installing Exiftool"
	sudo apt-get install build-essential -y > setup.log
	wget -c https://exiftool.org/Image-ExifTool-12.22.tar.gz > setup.log
	gzip -dc Image-ExifTool-12.22.tar.gz | tar -xf - > setup.log
	cd Image-ExifTool-12.22 > setup.log
	perl Makefile.PL > setup.log
	make test > setup.log
	sudo make install > setup.log
	cd .. > setup.log
	sudo rm -r Image-ExifTool-12.22 
	rm Image-ExifTool-12.22.tar.gz 

	echo "Getting protos"
	cd ../../..
	git clone https://${GITHUB_TOKEN}@github.com/Seneca-AI/common.git > seneca/devops/setup/setup.log
	cp -r common/proto_out/go/api seneca > seneca/devops/setup/setup.log
	cd seneca

	echo "Installing golang libraries"
	sudo env "PATH=$PATH" "GOPATH=$GOPATH" go mod tidy > devops/setup/setup.log

	echo "Done"
}

# Open a port to allow incoming traffic.
open_port() {
	read -p "Enter the port: " PORT
	read -p "Enter the VM instance name: " VM_INSTANCE
	gcloud compute firewall-rules create rule-allow-tcp-$PORT --source-ranges 0.0.0.0/0 --target-tags allow-tcp-$PORT --allow tcp:$PORT
	gcloud compute instances add-tags $VM_INSTANCE --tags allow-tcp-$PORT
	echo "Done"
}

# Start the datagatherer.
start_datagatherer() {
	echo "Starting datagatherer server."
	read -p "Enter GOOGLE_CLOUD_PROJECT: " GOOGLE_CLOUD_PROJECT
	read -p "Enter absolute path to GOOGLE_APPLICATION_CREDENTIALS json file: " GOOGLE_APPLICATION_CREDENTIALS
	read -p "Enter absolute path to GOOGLE_OAUTH_CREDENTIALS json file: " GOOGLE_OAUTH_CREDENTIALS
	cd ../cmd/datagatherer
	sudo env "PATH=$PATH" "GOPATH=$GOPATH" "GOOGLE_CLOUD_PROJECT=$GOOGLE_CLOUD_PROJECT" "GOOGLE_APPLICATION_CREDENTIALS=$GOOGLE_APPLICATION_CREDENTIALS" "GOOGLE_OAUTH_CREDENTIALS=$GOOGLE_OAUTH_CREDENTIALS" go run .
}

# Start the singleserver.
start_singleserver() {
	if [ -z "$3" ]; then
		echo "Must specify 'bash setup.sh start_singleserver [GOOGLE_CLOUD_PROJECT] [ABSOLUTE_PATH_TO_GOOGLE_APPLICATION_CREDENTIALS] [ABSOLUTE_PATH_TO_GOOGLE_OAUTH_CREDENTIALS]'"
		exit 1
	fi

	export GOOGLE_CLOUD_PROJECT=$1
	export GOOGLE_APPLICATION_CREDENTIALS=$2
	export GOOGLE_OAUTH_CREDENTIALS=$3

	echo "Starting single server."
	cd ../../cmd/singleserver
	nohup sudo env "PATH=$PATH" "GOPATH=$GOPATH" "GOOGLE_CLOUD_PROJECT=$GOOGLE_CLOUD_PROJECT" "GOOGLE_APPLICATION_CREDENTIALS=$GOOGLE_APPLICATION_CREDENTIALS" "GOOGLE_OAUTH_CREDENTIALS=$GOOGLE_OAUTH_CREDENTIALS" go run . &
}

deploy_run_syncer() {
	# Create the pubsub topic.
	gcloud pubsub topics create run_syncer_topic
	gcloud scheduler jobs create pubsub run_syncer_job --schedule="*/10 * * * *" --topic="run_syncer_topic" --message-body="run"
	gcloud compute networks vpc-access connectors create default-connector --network=default --region=us-central1 --range=10.8.0.0/28
	gcloud functions deploy RunSyncer --runtime=go113 --trigger-topic="run_syncer_topic" --vpc-connector=default-connector --region=us-central1
}

if [ -z "$1" ]; then 
	echo "Must specify a command.  Options are [ help, setup, open_port, start_datagatherer, start_singleserver, deploy_run_syncer ]."
	exit 1
fi

if [ $1 == "help" ]; then
	if [ -z "$2" ]; then
		echo "Options are [ setup, open_port, start_datagatherer, start_singleserver, deploy_run_syncer ]. Type 'bash setup.sh help <command> to learn more."
	else
		if [ $2 == "setup" ]; then
			echo "Download and install all necessary dependencies."
		elif [ $2 == "open_port" ]; then 
			echo "Open a port to allow incoming traffic."
		elif [ $2 == "start_datagatherer" ]; then
			echo "Start the datagatherer."
		elif [ $2 == "start_singleserver" ]; then
			echo "Start the singleserver, specifying [GOOGLE_CLOUD_PROJECT] [ABSOLUTE_PATH_TO_GOOGLE_APPLICATION_CREDENTIALS] [ABSOLUTE_PATH_TO_GOOGLE_OAUTH_CREDENTIALS]"
		elif [ $2 == "deploy_run_syncer" ]; then
			echo "Start deploy run_syncer cloud function."
		else 
			echo "Invalid argument."
		fi
	fi
	exit 0
fi

if [ $1 == "setup" ]; then
	# Start the spinner.
	spin &
	SPIN_PID=$!
	trap "kill -9 $SPIN_PID" `seq 0 15`
	# Run setup.
	setup
	kill -9 $SPIN_PID
elif [ $1 == "open_port" ]; then 
	open_port
elif [ $1 == "start_datagatherer" ]; then
	start_datagatherer
elif [ $1 == "start_singleserver" ]; then
	start_singleserver $2 $3 $4
elif [ $1 == "deploy_run_syncer" ]; then
	deploy_run_syncer
else
	echo "Invalid argument."
fi