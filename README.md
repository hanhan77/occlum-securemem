# occlum-securemem
# build images
docker build -t occlum-go-seal .

# exec image to change json
docker run --privileged --device /dev/sgx/enclave --device /dev/sgx/provision -it --entrypoint /bin/bash occlum-securemem

cat Occlum.json 
vim Occlum.json 
rm -rf build run
occlum build

#  run app
occlum run /bin/app