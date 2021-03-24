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
1. Start the server!
    1. export GOPATH=$HOME/go
    1. export PATH=$PATH:$HOME/go/bin
    1. export GOOGLE_CLOUD_PROJECT="senecacam-sandbox"
    1. export CREDENTIALS_JSON="your_file.json"
    1. export GOOGLE_APPLICATION_CREDENTIALS="$PWD/configs/credentials/$CREDENTIALS_JSON"
    1. cd cmd/datagatherer
    1. sudo env "PATH=$PATH" "GOPATH=$GOPATH" "GOOGLE_CLOUD_PROJECT=$GOOGLE_CLOUD_PROJECT" "GOOGLE_APPLICATION_CREDENTIALS=$GOOGLE_APPLICATION_CREDENTIALS" go run .
