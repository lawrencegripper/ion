#!/bin/sh -e

SKIP_BUILD=$1
SKIP_TERRAFORM=$2

echo "--------------------------------------------------------"
echo "WARNING: This script will deploy into your currently selected Azure Subscription, Kubernetes clusters and Docker hub user"
echo "WARNING: This script will deploy into your currently selected Azure Subscription, Kubernetes clusters and Docker hub user"
echo "WARNING: This script will deploy into your currently selected Azure Subscription, Kubernetes clusters and Docker hub user"
echo "WARNING: This script will deploy into your currently selected Azure Subscription, Kubernetes clusters and Docker hub user"
echo "WARNING: This script will deploy into your currently selected Azure Subscription, Kubernetes clusters and Docker hub user"
echo "It expects you to have already run terraform to deploy ion and connected kubectl to the deployed cluster"
echo " Must: Be logged into Azure CLI and have the right subscription set as your default"
echo " Must: Be logged into docker cli and have set $DOCKER_USER to your username"
echo "--------------------------------------------------------"

sleep 5

if [ -z "$DOCKER_USER" ]
then
      echo "You must specify a $DOCKER_USER environment variable to which the ion images can be pushed"
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

    kubectl delete deployments --all || true
    kubectl delete jobs --all || true

    echo "--------------------------------------------------------"
    echo "Deploying terraform"
    echo "--------------------------------------------------------"

    cd ./deployment
    if [ ! -f ./vars.private.tfvars ]; then
        echo "vars.private.tfvars not found in deployment file!"
        echo "WARNING.... you'll need to create it some of the fields in ./deployment/vars.private.tfvars without it the terraform deployment will fail"
        return
    fi

    sed -i "s/docker_root.*/docker_root=\"$DOCKER_USER\"/g" vars.private.tfvars
    sed -i "s/docker_user.*/docker_user=\"$ION_IMAGE_TAG\"/g" vars.private.tfvars
    terraform apply -var-file ./vars.private.tfvars -auto-approve
    cd -
    echo "--------------------------------------------------------"
    echo "Wait for the pods to start"
    echo "--------------------------------------------------------"

    sleep 15
    kubectl get pods || true
else
    echo "--------------------------------------------------------"
    echo "Cleaning up k8s, removing all jobs and pods"
    echo "--------------------------------------------------------"
    kubectl delete jobs --all || true
fi

echo "--------------------------------------------------------"
echo "Forwarding ports for management api and front api"
echo "--------------------------------------------------------"

#Cleanup any leftover listeners
ps aux | grep [k]ubectl | awk '{print $2}' | xargs kill || true

kubectl get pods | grep ion-front | awk '{print $1}' | xargs -I % kubectl port-forward % 9001:9001 &
FORWARD_PID1=$!
kubectl get pods | grep ion-management | awk '{print $1}' | xargs -I % kubectl port-forward % 9000:9000 &
FORWARD_PID2=$!


echo "--------------------------------------------------------"
echo "Deploying downloader and transcoder module with tag $ION_IMAGE_TAG"
echo "--------------------------------------------------------"

docker run --network host ion-cli module create -i frontapi.new_link -o file_downloaded -n downloader -m $DOCKER_USER/ion-module-download-file:$ION_IMAGE_TAG -p kubernetes --handler-image $DOCKER_USER/ion-handler:$ION_IMAGE_TAG
docker run --network host ion-cli module create -i file_downloaded -o file_transcoded -n transcodegpu -m $DOCKER_USER/ion-module-transcode:$ION_IMAGE_TAG -p azurebatch --handler-image $DOCKER_USER/ion-handler:$ION_IMAGE_TAG

read -p "Press enter to to stop forwarding ports to management api and front api and exit..." key
kill $FORWARD_PID1
kill $FORWARD_PID2