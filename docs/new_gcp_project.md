# How to Set Up a New Instance of the Backend Application

1. Setup the GCP project
    1. Sign into gcloud using your @senecacam.com email
        * `$ gcloud auth login`
    1. Create a new project within the senecacam.com organization.
        * `$ gcloud projects create ${PROJECT_ID} --organization=161538395481`
    1. Switch to the project's context.
        * `$ gcloud config set project ${PROJECT_ID}`
    1. [Optional] Grant some other account owner permissions.
        *  This cannot be done through gcloud. Visit console.cloud.google.com --> Top-left hamburger menu --> IAM & Admin --> IAM --> 'ADD' --> Roles --> Basic --> Owner
    1.  Link a billing account
        * This cannot be done through gcloud. Visit console.cloud.google.com --> Top-left hamburger menu --> Billing
    1. Enable datastore in Datastore mode
        * This cannot be done through gcloud. Visit console.cloud.google.com --> Top-left hamburger menu --> Datastore --> SELECT DATASTORE MODE --> nam5 --> Create database
    1. Clone the common repo
        * `$ git clone https://${GITHUB_TOKEN}@github.com/Seneca-AI/common.git`
    1. Setup the Datastore index
        * `$ gcloud datastore indexes create common/googlecloud/datastore/index.yaml`
    1. Setup service account.  Note EMAIL can be your email from step 4, or your senecacam.com email.  Also make sure to save the key in a safe place and note its location.
        * `$ gcloud iam service-accounts create full-admin`
        * `$ gcloud iam service-accounts add-iam-policy-binding full-admin@${PROJECT_ID}.iam.gserviceaccount.com --member='user:${EMAIL}' --role='roles/owner'`
        * `$ gcloud iam service-accounts keys create ${PROJECT_ID}-full-admin-google-application-credentials.json  --iam-account=full-admin@${PROJECT_ID}.iam.gserviceaccount.com`
    1. Setup OAuth consent screen for getting OAuth tokens from users.
        * This cannot be done through gcloud.  Visit console.cloud.google.com --> Top-left hamburger menu --> APIs & Services --> OAuth consent screen
            * Configure OAuth Consent Screen
                * Choose 'Internal'
                * App name: SenecaCam
                * User support email: your senecacam email
                * Developer contact informatino: your senecacam email
                * SAVE AND CONTINUE
                * SAVE AND CONTINUE
            * Create credentials
                * Left side menu --> Credentials --> + CREATE CREDENTIALS
                * Application type: Desktop app
                * Name: desktop-oauth-client
                * Download the OAuth Client Secret all the way to the right of the new list item
                * mv ~/Downloads/*.apps.googleusercontent.com.json ./${PROJECT_ID}-oauth-credentials.json
                * Mke sure to save this key in a safe place and note its location.
1. Setup the server
    1. Create a VM for the Seneca server, and give it some time to setup
        * `$ gcloud compute instances create singleserver --image=ubuntu-minimal-2104-hirsute-v20210511 --image-project=ubuntu-os-cloud --zone=northamerica-northeast1-a && sleep 30`
    1. Copy the JSON credentials you previously created into the VM
        * `$ gcloud compute scp ${APPLICATION_CREDENTIALS_JSON_FILE} singleserver:~`
        * `$ gcloud compute scp ${OAUTH_CREDENTIALS_JSON_FILE} singleserver:~`
    1. SSH into the VM
        * `$ gcloud compute ssh singleserver`
    1. Make a folder for credentials and copy them over
        * `$ mkdir credentials`
        * `$ mv *.json credentials/`
    1. Install git
        * `$ sudo apt-get update`
        * `$ sudo apt-get install -y apt-utils`
        * `$ sudo apt-get install git -y`
    1. Clone the seneca repo
        * `$ git clone https://${GITHUB_TOKEN}@github.com/Seneca-AI/seneca.git`
    1. Run the setup script
        * `$ cd seneca/devops/setup`
        * `$ bash setup.sh setup`
        * Wait a while for the script to complete...keep pressing 'Enter' if you see 'Which services should be restarted?'
    1. Log in to gcloud on the VM
        * `$ gcloud auth login`
    1. Open up port 6060 to external traffic
        * `$ bash setup.sh open_port`
    1. Start the server
        * `$ bash setup.sh start_singleserver`
1. You'll need to get a user's oauth token and add it to the DB yourself, for now.  To make requests to the VM, note the VM's external IP.

TODO: setup zone file, setup API key to hit single server, get token using golang
