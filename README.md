# seneca

Requirements:
* [Exiftool](https://exiftool.org/install.html#Unix) must be installed on the server for the datagatherer server to work
* There are many setup steps, like downloading go requirements and generating protobuf files, that are not yet documented here.  They will soon be automated in a Makefile.

### Setup from a fresh VM
1. Get the repository
    1. `sudo apt-get install git -y`
    1. `git clone https://github.com/Seneca-AI/seneca.git`
    1. `cd seneca`
1. Run the setup script
    1. `./setup.sh`


### Open up VM for external traffic
1. gcloud compute firewall-rules create rule-allow-tcp-8080 --source-ranges 0.0.0.0/0 --target-tags allow-tcp-8080 --allow tcp:8080
1. gcloud compute instances add-tags $VM_INSTANCE --tags allow-tcp-8080

### Start the server
1. source env/ENV
1. export GOOGLE_APPLICATION_CREDENTIALS="absolute/path/to/credentials"
1. cd cmd/datagatherer
1. sudo env "PATH=$PATH" "GOPATH=$GOPATH" "GOOGLE_CLOUD_PROJECT=$GOOGLE_CLOUD_PROJECT" "GOOGLE_APPLICATION_CREDENTIALS=$GOOGLE_APPLICATION_CREDENTIALS" go run .
