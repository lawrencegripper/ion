#!/bin/sh -e
cd "$(dirname "$0")"
cd ..

SKIP_BUILD=$1
SKIP_TERRAFORM=$2

echo "--------------------------------------------------------"
echo "WARNING: This script will deploy into your currently selected Azure Subscription, Kubernetes clusters and Docker hub user"
echo "WARNING: This script will deploy into your currently selected Azure Subscription, Kubernetes clusters and Docker hub user"
echo "WARNING: This script will deploy into your currently selected Azure Subscription, Kubernetes clusters and Docker hub user"
echo "WARNING: This script will deploy into your currently selected Azure Subscription, Kubernetes clusters and Docker hub user"
echo "WARNING: This script will deploy into your currently selected Azure Subscription, Kubernetes clusters and Docker hub user"
echo "You must have already:"
echo " MUST: Run terraform init in the ./deployment folder"
echo " MUST: have kubectl installed and available in your path"
echo " Must: Be logged into Azure CLI and have the right subscription set as your default"
echo " Must: Be logged into Docker CLI and have set $DOCKER_USER to your username"
echo "--------------------------------------------------------"

sleep 5

if [ -z "$DOCKER_USER" ]
then
      echo "You must specify a $DOCKER_USER environment variable to which the ion images can be pushed"
      exit 1
fi

if [ -z "$SKIP_BUILD" ]
then
    echo "--------------------------------------------------------"
    echo "Building source and pushing images"
    echo "--------------------------------------------------------"

    make
    ./build/pushimages.sh
fi

export ION_IMAGE_TAG=$(cat imagetag.temp)
echo "-> Using tag $ION_IMAGE_TAG"

if [ -z "$SKIP_TERRAFORM" ]
then
    #Refresh the azurecli token
    az group list >> /dev/null

    echo "--------------------------------------------------------"
    echo "Cleaning up k8s, removing all deployments"
    echo "--------------------------------------------------------"

    if [ -f "./kubeconfig.private.yaml" ]
    then
        echo "Kubeconfig found cleaning up cluster."
        export KUBECONFIG=./kubeconfig.private.yaml
        kubectl delete deployments --all || true
        kubectl delete jobs --all || true
        kubectl delete pods --all || true
        kubectl delete secrets --all || true
        kubectl delete service --all || true
    else
        echo "Kubeconfig not found, no cluster created skipping cleanup..."
    fi


    echo "--------------------------------------------------------"
    echo "Deploying terraform"
    echo "--------------------------------------------------------"

    cd ./deployment
    if [ ! -f ./vars.private.tfvars ]; then
        echo "vars.private.tfvars not found in deployment file!"
        echo "WARNING.... you'll need to create it some of the fields in ./deployment/vars.private.tfvars without it the terraform deployment will fail"
        return
    fi

    terraform init
    terraform apply -var-file ./vars.private.tfvars -auto-approve -var docker_root=$DOCKER_USER -var docker_tag=$ION_IMAGE_TAG
    terraform output kubeconfig > ../kubeconfig.private.yaml

    echo "--------------------------------------------------------"
    echo "Setting kubectl context to new cluster"
    echo "--------------------------------------------------------"
    az aks get-credentials -n $(terraform output cluster_name) -g $(terraform output resource_group_name)
    cd -

    echo "--------------------------------------------------------"
    echo "Wait for the pods to start"
    echo "--------------------------------------------------------"

    sleep 15

    export KUBECONFIG=./kubeconfig.private.yaml
    kubectl get pods || true

else
    echo "--------------------------------------------------------"
    echo "Cleaning up k8s, removing all jobs and pods"
    echo "--------------------------------------------------------"
    kubectl delete jobs --all || true
fi

CERT_DIR="$PWD/tf"
mkdir -p $CERT_DIR
cd ./deployment
terraform output client_cert > "$CERT_DIR/client.crt"
terraform output client_key > "$CERT_DIR/client.key"
terraform output cluster_ca > "$CERT_DIR/rootCA.pem"
export FQDN=$(terraform output fqdn)
cd -

echo "--------------------------------------------------------"
echo "Forwarding ports for front api"
echo "--------------------------------------------------------"

#Cleanup any leftover listeners
ps aux | grep [k]ubectl | awk '{print $2}' | xargs kill || true

kubectl get pods | grep ion-front | awk '{print $1}' | xargs -I % kubectl port-forward % 9001:9001 &
FORWARD_PID1=$!

echo "--------------------------------------------------------"
echo "Deploying downloader and transcoder module with tag $ION_IMAGE_TAG"
echo "--------------------------------------------------------"

docker run --rm --network host -v ${PWD}/tf:/src/tf ion-cli module create -i frontapi.new_link -o file_downloaded \
-n downloader -m $DOCKER_USER/ion-module-download-file:$ION_IMAGE_TAG \
-p kubernetes --handler-image $DOCKER_USER/ion-handler:$ION_IMAGE_TAG \
--endpoint $FQDN:9000 \
--certfile /src/tf/client.crt \
--keyfile /src/tf/client.key \
--cacertfile /src/tf/rootCA.pem

docker run --rm --network host -v ${PWD}:/src ion-cli module create -i file_downloaded \
-o file_transcoded -n transcode -m $DOCKER_USER/ion-module-transcode:$ION_IMAGE_TAG -p azurebatch \
--handler-image $DOCKER_USER/ion-handler:$ION_IMAGE_TAG \
--config-map-file /src/tools/transcoder.env \
--endpoint $FQDN:9000 \
--certfile /src/tf/client.crt \
--keyfile /src/tf/client.key \
--cacertfile /src/tf/rootCA.pem

sleep 30

echo "--------------------------------------------------------"
echo "Submitting a video for processing to the frontapi"
echo "--------------------------------------------------------"

curl --header "Content-Type: application/json"   --request POST   --data '{"url": "http://download.blender.org/peach/bigbuckbunny_movies/BigBuckBunny_320x180.mp4"}'   http://localhost:9001/

if [ -x "$(command -v beep)" ]; then
    beep
fi

if [ -x "$(command -v notify-send)" ]; then
    notify-send -u critical ion-end2end "Ion ready for testing"
fi

read -p "Press enter to to stop forwarding ports to management api and front api and exit..." key
ps aux | grep [k]ubectl | awk '{print $2}' | xargs kill || true

