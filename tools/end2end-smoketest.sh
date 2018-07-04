#!/bin/sh -e

echo "--------------------------------------------------------"
echo "WARNING: This script expects you to have already run terraform to deploy ion and connected kubectl to the deployed cluster"
echo "--------------------------------------------------------"

if [ -z "$DOCKER_USER" ]
then
      echo "You must specify a $DOCKER_USER environment variable to which the ion images can be pushed"
fi

echo "--------------------------------------------------------"
echo "Building source and pushing images"
echo "--------------------------------------------------------"

make
./build/pushimages.sh
export ION_IMAGE_TAG=$(cat imagetag.temp)
echo "-> Using tag $ION_IMAGE_TAG" 

echo "--------------------------------------------------------"
echo "Deploying terraform"
echo "--------------------------------------------------------"

cd ./deployment
if [ ! -f ./vars.private.tfvars ]; then
    echo "vars.private.tfvars not found in deployment file!"
    echo "WARNING.... you'll need to go fill in some of the fields in ./deployment/vars.private.tfvars without it the terraform deployment will fail"
    cp ./vars.example.tfvars ./vars.private.tfvars
fi

sed -i "s/docker_root.*/docker_root=\"$ION_IMAGE_TAG\"/g" vars.private.tfvars
sed -i "s/docker_user.*/docker_user=\"$DOCKER_USER\"/g" vars.private.tfvars
terraform apply -var-file ./vars.private.tfvars -auto-approve
cd -

echo "--------------------------------------------------------"
echo "Cleaning up k8s, removing all jobs, deployments and pods"
echo "--------------------------------------------------------"

kubectl delete deployments --all
kubectl delete jobs --all 
kubectl delete pods --all 

echo "--------------------------------------------------------"
echo "Forwarding ports for management api and front api"
echo "--------------------------------------------------------"

kubectl get pods | grep ion-front | awk '{print $1}' | xargs -I % kubectl port-forward % 9001:9001 &
kubectl get pods | grep ion-management | awk '{print $1}' | xargs -I % kubectl port-forward % 9000:9000 &

echo "--------------------------------------------------------"
echo "Deploying downloader module with tag $ION_IMAGE_TAG"
echo "--------------------------------------------------------"

docker run --network host ion-cli module create -i frontapi.new_link -o file_downloaded -n downloader -m $DOCKER_USER/ion-module-download-file:$ION_IMAGE_TAG -p kubernetes --handler-image $DOCKER_USER/ion-handler:$ION_IMAGE_TAG

read -n1 -r -p "Press any key to stop forwarding ports to management api and front api..." key
rm *.temp
